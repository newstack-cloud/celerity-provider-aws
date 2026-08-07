package flex

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink-provider-aws/ec2util"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Named security groups are empty groups the VPC prepares for resources that need an
// identity within it, exposed as securityGroupIdsByName for a resource's own spec to
// reference.
//
// The VPC prepares them and nothing more. Attaching a group to a database is the database's
// own spec field, so it stays in the blueprint where state, spec and live AWS all agree
// on it. A link that attached the group instead would be writing a whole-list property on
// a resource it does not own, and the target's next unrelated update would diff that list
// against the blueprint and silently detach it.
//
// The rules come later and from elsewhere: the link between a workload and the resource
// opens a path between the workload's group and this one. The VPC doesn't write rules.

// A referenced VPC is shared and its topology is owned by another blueprint, so the only
// thing deployed here is this application's own named security groups. Reconciling
// against the groups discovered for this instance covers the first deploy, a re-deploy
// after a partial failure, and any later change alike.
func (l *vpcResourceActions) referenceModeDeployOutput(
	ctx context.Context,
	service ec2service.Service,
	input *provider.ResourceDeployInput,
	resourceSpecData *core.MappingNode,
	externalStateOutput *provider.ResourceGetExternalStateOutput,
) (*provider.ResourceDeployOutput, error) {
	computedFields := extractComputedFieldsFromExternalState(externalStateOutput.ResourceSpecState)

	name, hasName := pluginutils.GetValueByPath("$.name", resourceSpecData)
	if !hasName {
		return nil, errors.New("name is required for a referenced flex VPC")
	}
	vpcName := core.StringValue(name)

	vpcID, _ := pluginutils.GetValueByPath("$.vpcId", externalStateOutput.ResourceSpecState)
	vpcCtx := &vpcContext{
		resourceSpecData: resourceSpecData,
		vpcName:          vpcName,
		// Built the same way as every other path, minus the user tags, which
		// describe the owner's VPC and are ignored in reference mode.
		tags: vpcElementTags(
			vpcName,
			core.MappingNodeFields(),
			getBluelinkTagsAsEC2Tags(input),
		),
		instanceID: input.InstanceID,
		vpcID:      aws.String(core.StringValue(vpcID)),
	}

	groupIDs, err := l.reconcileNamedSecurityGroups(ctx, service, vpcCtx)
	if err != nil {
		return nil, err
	}

	computedFields["spec.securityGroupIdsByName"] = toSpecComputedSecurityGroupIdsByName(
		groupIDs,
	)

	return &provider.ResourceDeployOutput{ComputedFieldValues: computedFields}, nil
}

// Reads the declared group names, in declaration order with duplicates removed.
func parseSecurityGroupNames(resourceSpecData *core.MappingNode) []string {
	node, has := pluginutils.GetValueByPath("$.securityGroups", resourceSpecData)
	if !has || node == nil {
		return nil
	}

	names := []string{}
	seen := map[string]bool{}
	for _, item := range node.Items {
		name := core.StringValue(item)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}

	return names
}

func (l *vpcResourceActions) createNamedSecurityGroups(
	ctx context.Context,
	service ec2service.Service,
	vpcCtx *vpcContext,
) (map[string]string, error) {
	names := parseSecurityGroupNames(vpcCtx.resourceSpecData)
	if len(names) == 0 {
		return map[string]string{}, nil
	}

	groupIDs := map[string]string{}
	for _, name := range names {
		groupID, err := l.createNamedSecurityGroup(ctx, service, vpcCtx, name)
		if err != nil {
			return nil, err
		}
		groupIDs[name] = groupID
	}

	return groupIDs, nil
}

func (l *vpcResourceActions) createNamedSecurityGroup(
	ctx context.Context,
	service ec2service.Service,
	vpcCtx *vpcContext,
	name string,
) (string, error) {
	groupOutput, err := service.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(fmt.Sprintf("%s-%s", vpcCtx.vpcName, name)),
		Description: aws.String(fmt.Sprintf("Group %s for %s", name, vpcCtx.vpcName)),
		VpcId:       vpcCtx.vpcID,
		TagSpecifications: createSecurityGroupTagSpecifications(
			tagsWithSecurityGroupName(
				vpcCtx.tags,
				vpcCtx.vpcName,
				name,
				vpcCtx.instanceID,
			),
		),
	})
	if err != nil {
		return "", err
	}

	// A group is an identity, not a grant. AWS attaches allow-all egress to a new group,
	// which would let anything referencing this group reach the whole internet on the
	// strength of having been given a name.
	_, err = service.RevokeSecurityGroupEgress(ctx, &ec2.RevokeSecurityGroupEgressInput{
		GroupId:       groupOutput.GroupId,
		IpPermissions: allTrafficPermissions(),
	})
	if err != nil {
		return "", err
	}

	return aws.ToString(groupOutput.GroupId), nil
}

func tagsWithSecurityGroupName(
	tags []types.Tag,
	vpcName string,
	name string,
	instanceID string,
) []types.Tag {
	return flexVPCResourceTags(tags, vpcName,
		flexVPCTag(TagFlexVPCSecurityGroupName, name),
		flexVPCTag(TagFlexVPCSecurityGroupNameOwner, instanceID),
	)
}

func toSpecComputedSecurityGroupIdsByName(groupIDs map[string]string) *core.MappingNode {
	fields := map[string]*core.MappingNode{}
	for name, groupID := range groupIDs {
		fields[name] = core.MappingNodeFromString(groupID)
	}

	return &core.MappingNode{Fields: fields}
}

// The named groups this blueprint instance owns in the VPC, read back from their tags.
func namedSecurityGroupIDs(
	ctx context.Context,
	service ec2service.Service,
	vpcName string,
	instanceID string,
) (map[string]string, error) {
	output, err := service.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", TagFlexVPCName)),
				Values: []string{vpcName},
			},
			{
				Name:   aws.String(fmt.Sprintf("tag:%s", TagFlexVPCSecurityGroupNameOwner)),
				Values: []string{instanceID},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if output == nil {
		return map[string]string{}, nil
	}

	groupIDs := map[string]string{}
	for _, group := range output.SecurityGroups {
		name := tagValue(group.Tags, TagFlexVPCSecurityGroupName)
		if name == "" {
			continue
		}
		groupIDs[name] = aws.ToString(group.GroupId)
	}

	return groupIDs, nil
}

func tagValue(tags []types.Tag, key string) string {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key {
			return aws.ToString(tag.Value)
		}
	}

	return ""
}

// Brings the VPC's named groups in line with what is declared: prepare the ones that are
// new, delete the ones that are gone.
//
// A group that is still referenced by a resource cannot be deleted, and EC2 says so with
// DependencyViolation. That is the right outcome as it means the blueprint dropped the name
// while a resource still points at the group, and deleting it would strip that resource
// of its identity in the VPC.
func (l *vpcResourceActions) reconcileNamedSecurityGroups(
	ctx context.Context,
	service ec2service.Service,
	vpcCtx *vpcContext,
) (map[string]string, error) {
	existing, err := namedSecurityGroupIDs(ctx, service, vpcCtx.vpcName, vpcCtx.instanceID)
	if err != nil {
		return nil, err
	}

	declared := parseSecurityGroupNames(vpcCtx.resourceSpecData)
	declaredSet := map[string]bool{}
	for _, name := range declared {
		declaredSet[name] = true
	}

	groupIDs := map[string]string{}
	for _, name := range declared {
		if groupID, ok := existing[name]; ok {
			groupIDs[name] = groupID
			continue
		}
		groupID, err := l.createNamedSecurityGroup(ctx, service, vpcCtx, name)
		if err != nil {
			return nil, err
		}
		groupIDs[name] = groupID
	}

	names := sortedNames(existing)
	for _, name := range names {
		if declaredSet[name] {
			continue
		}
		_, err := service.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
			GroupId: aws.String(existing[name]),
		})
		if err != nil {
			return nil, fmt.Errorf(
				"the %q security group could not be removed from flex VPC %q: %w",
				name,
				vpcCtx.vpcName,
				err,
			)
		}
	}

	return groupIDs, nil
}

// Removes the named groups this instance owns.
//
// Read back from AWS by tag rather than from the spec, so a group created by an update
// after the state in hand was written is still removed. Every rule is revoked across all
// of them before any group is deleted: the workload group holds egress to these and these
// hold ingress from it, and a group referenced by a rule on another group cannot be
// deleted. Revoking one at a time in turn would deadlock on the first pair.
func (l *vpcResourceActions) deleteNamedSecurityGroups(
	ctx context.Context,
	service ec2service.Service,
	vpcName string,
	instanceID string,
) error {
	groupIDs, err := namedSecurityGroupIDs(ctx, service, vpcName, instanceID)
	if err != nil {
		return err
	}
	if len(groupIDs) == 0 {
		return nil
	}

	ordered := make([]string, 0, len(groupIDs))
	names := sortedNames(groupIDs)
	for _, name := range names {
		ordered = append(ordered, groupIDs[name])
	}

	for _, groupID := range ordered {
		if err := revokeAllRules(ctx, service, groupID); err != nil {
			return err
		}
	}

	for _, groupID := range ordered {
		if err := ec2util.DeleteSecurityGroupWhenUnused(ctx, service, groupID); err != nil {
			return err
		}
	}

	return nil
}

// Deleted in a deterministic order so a teardown that fails part way through leaves the
// same state it would on any other run.
func sortedNames(groupIDs map[string]string) []string {
	names := make([]string, 0, len(groupIDs))
	for name := range groupIDs {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}
