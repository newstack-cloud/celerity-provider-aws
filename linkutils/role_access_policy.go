package linkutils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
)

// This file is the IAM I/O layer of the role-access permission allocator. It
// discovers the role's current Bluelink-managed policy slots, asks the pure
// planner ([planRoleAccess]) where the grant belongs, and applies the resulting
// mutations across inline and managed policies. Callers must hold the per-role
// lock (acquired via the link framework's ResourceService) around this call so the
// read-modify-write over the role's shared policy set is serialised.
//
// The IAM client retries throttling/transient errors via the AWS SDK's default
// retryer, so no manual retry is layered here.

// The IAM limit on versions per managed policy; the
// oldest non-default versions are pruned before a new version is created.
const maxManagedPolicyVersions = 5

// RoleAccessGrant describes a single link's access grant to apply to a role.
type RoleAccessGrant struct {
	// RoleName is the IAM role to grant access on (the resolved execution role).
	RoleName string
	// SID identifies this link's statement within the role's policies.
	SID string
	// Statement is the IAM policy statement to upsert; nil removes the grant.
	Statement map[string]any
	// Limits overrides the packing budgets; the zero value uses the defaults.
	Limits AccessPolicyLimits
}

// InlineAccessPolicyName returns the name of the shared inline policy the
// allocator manages on a role. Links use it to build the ResourceDataMappings path
// that attributes their statement to the link (suppressing role drift).
func InlineAccessPolicyName() string {
	return inlineAccessPolicyName
}

// RoleAccessResult reports where the grant was placed (empty when removed).
type RoleAccessResult struct {
	// PlacedSlot is the name of the policy the grant now lives in (empty on remove).
	PlacedSlot string
	// PlacedSlotInline reports whether the grant landed in the role's inline policy
	// (true) versus an attached managed policy (false). Meaningful only when
	// PlacedSlot is non-empty.
	PlacedSlotInline bool
	// PlacedSlotARN is the ARN of the managed policy the grant landed in, set only
	// for managed (overflow) placements. Links use it to build the
	// managedPolicyArns ResourceDataMapping that attributes the attachment to the
	// link (suppressing role drift).
	PlacedSlotARN string
}

// ReconcileRoleAccessPolicy applies a single access grant to a role, packing it
// into the role's inline policy or a managed policy slot per the allocator's
// strategy. The caller must hold the per-role lock.
func ReconcileRoleAccessPolicy(
	ctx context.Context,
	iamService iamservice.Service,
	grant RoleAccessGrant,
) (RoleAccessResult, error) {
	limits := grant.Limits
	if limits == (AccessPolicyLimits{}) {
		limits = DefaultAccessPolicyLimits()
	}

	slots, err := discoverRoleAccessSlots(ctx, iamService, grant.RoleName)
	if err != nil {
		return RoleAccessResult{}, err
	}

	var statement json.RawMessage
	removeOnly := grant.Statement == nil
	if !removeOnly {
		statement, err = json.Marshal(grant.Statement)
		if err != nil {
			return RoleAccessResult{}, err
		}
	}

	plan, err := planRoleAccess(slots, grant.SID, statement, removeOnly, limits)
	if err != nil {
		return RoleAccessResult{}, fmt.Errorf("on role %q: %w", grant.RoleName, err)
	}

	placedARN, err := applyAccessPolicyPlan(ctx, iamService, grant.RoleName, plan)
	if err != nil {
		return RoleAccessResult{}, err
	}

	inline := plan.PlacedSlot == inlineAccessPolicyName
	if !inline && plan.PlacedSlot != "" && placedARN == "" {
		// The grant landed in an existing managed slot whose document did not change
		// (no upsert), so recover its ARN from the discovered slots.
		placedARN = managedSlotARN(slots, plan.PlacedSlot)
	}

	return RoleAccessResult{
		PlacedSlot:       plan.PlacedSlot,
		PlacedSlotInline: inline,
		PlacedSlotARN:    placedARN,
	}, nil
}

func managedSlotARN(slots []AccessPolicySlot, name string) string {
	for _, slot := range slots {
		if slot.Kind == AccessPolicyManaged && slot.Name == name {
			return slot.ARN
		}
	}
	return ""
}

func discoverRoleAccessSlots(
	ctx context.Context,
	iamService iamservice.Service,
	roleName string,
) ([]AccessPolicySlot, error) {
	var slots []AccessPolicySlot

	inline, found, err := readInlineSlot(ctx, iamService, roleName)
	if err != nil {
		return nil, err
	}
	if found {
		slots = append(slots, inline)
	}

	managed, err := readManagedSlots(ctx, iamService, roleName)
	if err != nil {
		return nil, err
	}
	return append(slots, managed...), nil
}

func readInlineSlot(
	ctx context.Context,
	iamService iamservice.Service,
	roleName string,
) (AccessPolicySlot, bool, error) {
	var marker *string
	for {
		out, err := iamService.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
			RoleName: aws.String(roleName),
			Marker:   marker,
		})
		if err != nil {
			return AccessPolicySlot{}, false, err
		}
		for _, name := range out.PolicyNames {
			if name != inlineAccessPolicyName {
				continue
			}
			policy, err := iamService.GetRolePolicy(ctx, &iam.GetRolePolicyInput{
				RoleName:   aws.String(roleName),
				PolicyName: aws.String(name),
			})
			if err != nil {
				return AccessPolicySlot{}, false, err
			}
			statements, err := parseStatements(aws.ToString(policy.PolicyDocument))
			if err != nil {
				return AccessPolicySlot{}, false, err
			}
			return AccessPolicySlot{
				Kind:       AccessPolicyInline,
				Name:       inlineAccessPolicyName,
				Statements: statements,
			}, true, nil
		}
		if !out.IsTruncated {
			return AccessPolicySlot{}, false, nil
		}
		marker = out.Marker
	}
}

func readManagedSlots(
	ctx context.Context,
	iamService iamservice.Service,
	roleName string,
) ([]AccessPolicySlot, error) {
	var slots []AccessPolicySlot
	var marker *string
	for {
		out, err := iamService.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
			RoleName: aws.String(roleName),
			Marker:   marker,
		})
		if err != nil {
			return nil, err
		}
		for _, attached := range out.AttachedPolicies {
			name := aws.ToString(attached.PolicyName)
			if !strings.HasPrefix(name, managedAccessPolicyPrefix) {
				continue
			}
			slot, err := readManagedSlot(ctx, iamService, name, aws.ToString(attached.PolicyArn))
			if err != nil {
				return nil, err
			}
			slots = append(slots, slot)
		}
		if !out.IsTruncated {
			break
		}
		marker = out.Marker
	}
	return slots, nil
}

func readManagedSlot(
	ctx context.Context,
	iamService iamservice.Service,
	name, arn string,
) (AccessPolicySlot, error) {
	policy, err := iamService.GetPolicy(ctx, &iam.GetPolicyInput{PolicyArn: aws.String(arn)})
	if err != nil {
		return AccessPolicySlot{}, err
	}

	version := ""
	if policy.Policy != nil {
		version = aws.ToString(policy.Policy.DefaultVersionId)
	}

	policyVersion, err := iamService.GetPolicyVersion(ctx, &iam.GetPolicyVersionInput{
		PolicyArn: aws.String(arn),
		VersionId: aws.String(version),
	})
	if err != nil {
		return AccessPolicySlot{}, err
	}

	document := ""
	if policyVersion.PolicyVersion != nil {
		document = aws.ToString(policyVersion.PolicyVersion.Document)
	}
	statements, err := parseStatements(document)
	if err != nil {
		return AccessPolicySlot{}, err
	}

	return AccessPolicySlot{
		Kind:       AccessPolicyManaged,
		Name:       name,
		ARN:        arn,
		Statements: statements,
	}, nil
}

// Applies the plan and returns the ARN of the placed managed
// slot (empty for inline placements or when the placed slot was not upserted).
func applyAccessPolicyPlan(
	ctx context.Context,
	iamService iamservice.Service,
	roleName string,
	plan AccessPolicyPlan,
) (string, error) {
	placedARN := ""
	for _, slot := range plan.Upserts {
		arn, err := upsertSlot(ctx, iamService, roleName, slot)
		if err != nil {
			return "", err
		}
		if slot.Kind == AccessPolicyManaged && slot.Name == plan.PlacedSlot {
			placedARN = arn
		}
	}

	for _, slot := range plan.Deletes {
		if err := deleteSlot(ctx, iamService, roleName, slot); err != nil {
			return "", err
		}
	}

	return placedARN, nil
}

// Writes a slot's document and returns the managed policy ARN it wrote to
// (empty for inline slots).
func upsertSlot(
	ctx context.Context,
	iamService iamservice.Service,
	roleName string,
	slot AccessPolicySlot,
) (string, error) {
	document, err := serialiseSlotDocument(slot)
	if err != nil {
		return "", err
	}

	if slot.Kind == AccessPolicyInline {
		_, err = iamService.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
			RoleName:       aws.String(roleName),
			PolicyName:     aws.String(slot.Name),
			PolicyDocument: aws.String(document),
		})
		return "", err
	}

	if slot.ARN == "" {
		out, err := iamService.CreatePolicy(ctx, &iam.CreatePolicyInput{
			PolicyName:     aws.String(slot.Name),
			PolicyDocument: aws.String(document),
		})
		if err != nil {
			return "", err
		}
		arn := ""
		if out.Policy != nil {
			arn = aws.ToString(out.Policy.Arn)
		}
		_, err = iamService.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: aws.String(arn),
		})
		return arn, err
	}

	if err := pruneManagedPolicyVersions(ctx, iamService, slot.ARN); err != nil {
		return "", err
	}
	_, err = iamService.CreatePolicyVersion(ctx, &iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(slot.ARN),
		PolicyDocument: aws.String(document),
		SetAsDefault:   true,
	})
	return slot.ARN, err
}

func deleteSlot(
	ctx context.Context,
	iamService iamservice.Service,
	roleName string,
	slot AccessPolicySlot,
) error {
	if slot.Kind == AccessPolicyInline {
		_, err := iamService.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
			RoleName:   aws.String(roleName),
			PolicyName: aws.String(slot.Name),
		})
		return err
	}

	_, err := iamService.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(slot.ARN),
	})
	if err != nil {
		return err
	}
	if err := deleteNonDefaultVersions(ctx, iamService, slot.ARN); err != nil {
		return err
	}
	_, err = iamService.DeletePolicy(ctx, &iam.DeletePolicyInput{PolicyArn: aws.String(slot.ARN)})
	return err
}

// Removes the oldest non-default versions so a new
// version can be created within the IAM five-versions-per-policy limit.
func pruneManagedPolicyVersions(
	ctx context.Context,
	iamService iamservice.Service,
	arn string,
) error {
	out, err := iamService.ListPolicyVersions(ctx, &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(arn),
	})
	if err != nil {
		return err
	}

	if len(out.Versions) < maxManagedPolicyVersions {
		return nil
	}

	// Oldest versions have the lowest version numbers; delete oldest non-default
	// until there is room for a new version.
	versions := append([]iamPolicyVersion(nil), nonDefaultVersions(out.Versions)...)
	sort.Slice(versions, func(i, j int) bool { return versions[i].order < versions[j].order })
	for i := 0; i < len(versions) && len(out.Versions)-i >= maxManagedPolicyVersions; i++ {
		_, err := iamService.DeletePolicyVersion(ctx, &iam.DeletePolicyVersionInput{
			PolicyArn: aws.String(arn),
			VersionId: aws.String(versions[i].id),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteNonDefaultVersions(
	ctx context.Context,
	iamService iamservice.Service,
	arn string,
) error {
	out, err := iamService.ListPolicyVersions(ctx, &iam.ListPolicyVersionsInput{
		PolicyArn: aws.String(arn),
	})
	if err != nil {
		return err
	}

	for _, version := range nonDefaultVersions(out.Versions) {
		_, err := iamService.DeletePolicyVersion(ctx, &iam.DeletePolicyVersionInput{
			PolicyArn: aws.String(arn),
			VersionId: aws.String(version.id),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

type iamPolicyVersion struct {
	id    string
	order int
}

func nonDefaultVersions(versions []iamtypes.PolicyVersion) []iamPolicyVersion {
	var out []iamPolicyVersion
	for _, version := range versions {
		if version.IsDefaultVersion {
			continue
		}
		id := aws.ToString(version.VersionId)
		out = append(out, iamPolicyVersion{id: id, order: versionOrder(id)})
	}
	return out
}

// Extracts the numeric ordering from an IAM version id like "v3".
func versionOrder(versionID string) int {
	digits := strings.TrimPrefix(versionID, "v")
	order := 0
	for _, r := range digits {
		if r < '0' || r > '9' {
			return order
		}
		order = order*10 + int(r-'0')
	}
	return order
}

// Renders a slot's statements (sorted by Sid for
// determinism) into an IAM policy document JSON string.
func serialiseSlotDocument(slot AccessPolicySlot) (string, error) {
	sids := make([]string, 0, len(slot.Statements))
	for sid := range slot.Statements {
		sids = append(sids, sid)
	}
	sort.Strings(sids)

	statements := make([]json.RawMessage, 0, len(sids))
	for _, sid := range sids {
		statements = append(statements, slot.Statements[sid])
	}

	document := struct {
		Version   string            `json:"Version"`
		Statement []json.RawMessage `json:"Statement"`
	}{
		Version:   "2012-10-17",
		Statement: statements,
	}
	data, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Decodes an IAM policy document (URL-encoded as returned by IAM
// reads) into its statements keyed by Sid. Statements without a Sid are ignored.
// The allocator only manages its own keyed statements.
func parseStatements(document string) (map[string]json.RawMessage, error) {
	if document == "" {
		return map[string]json.RawMessage{}, nil
	}
	decoded, err := url.PathUnescape(document)
	if err != nil {
		decoded = document
	}

	var parsed struct {
		Statement []json.RawMessage `json:"Statement"`
	}
	if err := json.Unmarshal([]byte(decoded), &parsed); err != nil {
		return nil, fmt.Errorf("parsing managed access policy document: %w", err)
	}

	statements := map[string]json.RawMessage{}
	for _, raw := range parsed.Statement {
		var meta struct {
			Sid string `json:"Sid"`
		}
		if err := json.Unmarshal(raw, &meta); err != nil || meta.Sid == "" {
			continue
		}
		statements[meta.Sid] = raw
	}

	return statements, nil
}
