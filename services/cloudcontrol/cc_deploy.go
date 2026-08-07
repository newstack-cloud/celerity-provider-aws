package cloudcontrol

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscc "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	cctypes "github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

const clientTokenMaxLength = 128

func resolvedSpec(input *provider.ResourceDeployInput) *core.MappingNode {
	if input == nil || input.Changes == nil {
		return nil
	}

	resolved := input.Changes.AppliedResourceInfo.ResourceWithResolvedSubs
	if resolved == nil {
		return nil
	}

	return resolved.Spec
}

// Runs the ValidateResolvedSpec behaviour overlay against the raw resolved spec,
// before any engine transforms, so create and update validate the same view of
// the user's declared intent.
func (a *ccResourceActions) validateResolvedSpec(input *provider.ResourceDeployInput) error {
	if a.behaviour == nil || a.behaviour.ValidateResolvedSpec == nil {
		return nil
	}
	return a.behaviour.ValidateResolvedSpec(resolvedSpec(input))
}

func (a *ccResourceActions) buildDesiredCFNNode(
	input *provider.ResourceDeployInput,
) (*core.MappingNode, error) {
	if err := a.validateResolvedSpec(input); err != nil {
		return nil, err
	}

	desired := stripComputedFields(resolvedSpec(input), a.config.Schema)
	if desired == nil {
		desired = &core.MappingNode{Fields: map[string]*core.MappingNode{}}
	}

	if err := a.applyNameGeneration(desired, input); err != nil {
		return nil, err
	}

	if a.config.Meta.IsTaggable() {
		a.injectTags(desired, input)
	}

	if a.behaviour != nil && a.behaviour.BeforeCreate != nil {
		if err := a.behaviour.BeforeCreate(desired); err != nil {
			return nil, err
		}
	}

	return SpecToCFN(desired, a.config.Schema, a.config.Meta), nil
}

// Generates a name only when the user left the field empty.
func (a *ccResourceActions) applyNameGeneration(
	desired *core.MappingNode,
	input *provider.ResourceDeployInput,
) error {
	if a.behaviour == nil || a.behaviour.Name == nil {
		return nil
	}

	field := a.behaviour.Name.Field
	if existing := desired.Fields[field]; existing != nil && core.StringValue(existing) != "" {
		return nil
	}

	name, err := a.behaviour.Name.Generate(input)
	if err != nil {
		return err
	}

	if desired.Fields == nil {
		desired.Fields = map[string]*core.MappingNode{}
	}
	desired.Fields[field] = core.MappingNodeFromString(name)
	return nil
}

func (a *ccResourceActions) injectTags(
	desired *core.MappingNode,
	input *provider.ResourceDeployInput,
) {
	tagProperty := a.config.Meta.TagPropertyName
	userTags := extractUserTags(desired.Fields[tagProperty], a.config.Meta.TagShape)
	merged := utils.MergeBluelinkTagsWithUserTags(input, userTags)
	if len(merged) == 0 {
		return
	}

	if desired.Fields == nil {
		desired.Fields = map[string]*core.MappingNode{}
	}
	desired.Fields[tagProperty] = buildTagNode(merged, a.config.Meta.TagShape)
}

// Computed-state polling bounds. Cloud Control create/update are asynchronous; the
// deploy polls the request status to capture computed fields as early as they become
// available. The interval is the poll cadence; the timeout caps the wait for the rare
// case where a declared computed field (e.g. an endpoint address) is not populated
// until the operation completes, bounded generously for slow resources.
var (
	deployPollInterval = 5 * time.Second
	deployPollTimeout  = 30 * time.Minute
)

// Used for normal (fast) resources. Cloud Control only reliably exposes computed fields
// once the operation reaches SUCCESS (the IN_PROGRESS resource model is not guaranteed to
// carry them), so this effectively waits for completion before capturing.
func (a *ccResourceActions) captureComputedFields(
	ctx context.Context,
	service cloudcontrolservice.Service,
	input *provider.ResourceDeployInput,
	requestToken string,
	identifier string,
) (*provider.ResourceDeployOutput, error) {
	specState, resolvedID, err := a.awaitComputedState(
		ctx, service, input.ProviderContext, requestToken, identifier,
	)
	if err != nil {
		return nil, err
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: a.computedFieldValues(
			specState,
			resolvedID,
			requestToken,
			a.omittedAutoNamedFields(resolvedSpec(input)),
		),
	}, nil
}

// Returns the top-level fields marked computed-when-omitted (auto-named
// identifier components) that have no user-provided value in the given spec;
// the AWS-assigned values for these fields are reported as computed once the
// operation completes so references to them resolve from state.
func (a *ccResourceActions) omittedAutoNamedFields(spec *core.MappingNode) []string {
	if a.config.Schema == nil {
		return nil
	}

	var fields []string
	for name, attr := range a.config.Schema.Attributes {
		if attr == nil || !attr.ComputedWhenOmitted || attr.Computed {
			continue
		}
		if spec != nil && spec.Fields != nil {
			if value := spec.Fields[name]; value != nil && core.StringValue(value) != "" {
				continue
			}
		}
		fields = append(fields, name)
	}
	sort.Strings(fields)
	return fields
}

// Resolves the resource's primary identifier without waiting for the operation to
// complete. This is used for slow-to-stabilise resources so Create can return at config-complete
// instead of blocking for the whole provision. The identifier is usually already present
// on the create progress event (e.g. a user-provided DB identifier).
// This only polls for the identifier when it is not present in the progress event.
func (a *ccResourceActions) waitForIdentifier(
	ctx context.Context,
	service cloudcontrolservice.Service,
	token string,
	knownIdentifier string,
) (string, error) {
	if knownIdentifier != "" || token == "" {
		return knownIdentifier, nil
	}

	ctx, cancel := context.WithTimeout(ctx, deployPollTimeout)
	defer cancel()

	getStatus := pluginutils.RetryableReturnValue(
		func(
			ctx context.Context,
			in *awscc.GetResourceRequestStatusInput,
		) (*awscc.GetResourceRequestStatusOutput, error) {
			return service.GetResourceRequestStatus(ctx, in)
		},
		isCCErrorRetryable,
	)

	for {
		output, err := getStatus(ctx, &awscc.GetResourceRequestStatusInput{
			RequestToken: aws.String(token),
		})
		if err != nil {
			return "", err
		}

		if progress := output.ProgressEvent; progress != nil {
			if id := aws.ToString(progress.Identifier); id != "" {
				return id, nil
			}
			switch progress.OperationStatus {
			case cctypes.OperationStatusFailed,
				cctypes.OperationStatusCancelInProgress,
				cctypes.OperationStatusCancelComplete:
				return "", ccOperationFailedError(
					a.config.CFNType,
					progress.ErrorCode,
					aws.ToString(progress.StatusMessage),
				)
			}
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf(
				"timed out waiting for %s identifier to become available",
				a.config.CFNType,
			)
		case <-time.After(deployPollInterval):
		}
	}
}

func (a *ccResourceActions) awaitComputedState(
	ctx context.Context,
	service cloudcontrolservice.Service,
	providerContext provider.Context,
	token string,
	fallbackIdentifier string,
) (*core.MappingNode, string, error) {
	if token == "" {
		specState, err := a.fetchResourceSpecState(
			ctx,
			service,
			fallbackIdentifier,
			providerContext,
		)
		return specState, fallbackIdentifier, err
	}

	ctx, cancel := context.WithTimeout(ctx, deployPollTimeout)
	defer cancel()

	getStatus := pluginutils.RetryableReturnValue(
		func(
			ctx context.Context,
			in *awscc.GetResourceRequestStatusInput,
		) (*awscc.GetResourceRequestStatusOutput, error) {
			return service.GetResourceRequestStatus(ctx, in)
		},
		isCCErrorRetryable,
	)

	for {
		output, err := getStatus(ctx, &awscc.GetResourceRequestStatusInput{
			RequestToken: aws.String(token),
		})
		if err != nil {
			return nil, "", err
		}

		if progress := output.ProgressEvent; progress != nil {
			identifier := aws.ToString(progress.Identifier)
			if identifier == "" {
				identifier = fallbackIdentifier
			}

			switch progress.OperationStatus {
			case cctypes.OperationStatusFailed,
				cctypes.OperationStatusCancelInProgress,
				cctypes.OperationStatusCancelComplete:
				return nil, "", ccOperationFailedError(
					a.config.CFNType,
					progress.ErrorCode,
					aws.ToString(progress.StatusMessage),
				)
			case cctypes.OperationStatusSuccess:
				specState, err := a.completedSpecState(ctx, service, progress, identifier, providerContext)
				return specState, identifier, err
			default:
				specState, captured, err := a.tryEarlyCaptureState(progress, identifier, providerContext)
				if err != nil {
					return nil, "", err
				}
				if captured {
					return specState, identifier, nil
				}
			}
		}

		select {
		case <-ctx.Done():
			return nil, "", fmt.Errorf(
				"timed out waiting for %s computed fields to become available",
				a.config.CFNType,
			)
		case <-time.After(deployPollInterval):
		}
	}
}

// IN_PROGRESS / PENDING fast path, try early capture once the identifier and every declared
// computed field are available in the resource model.
func (a *ccResourceActions) tryEarlyCaptureState(
	progress *cctypes.ProgressEvent,
	identifier string,
	providerContext provider.Context,
) (specState *core.MappingNode, captured bool, err error) {
	if identifier == "" {
		return nil, false, nil
	}
	model := aws.ToString(progress.ResourceModel)
	if model == "" {
		return nil, false, nil
	}
	specState, err = a.cfnPropertiesToSpecState(model, identifier, providerContext)
	if err != nil {
		return nil, false, err
	}
	if !a.allComputedFieldsPresent(specState) {
		return nil, false, nil
	}
	return specState, true, nil
}

func (a *ccResourceActions) completedSpecState(
	ctx context.Context,
	service cloudcontrolservice.Service,
	progress *cctypes.ProgressEvent,
	identifier string,
	providerContext provider.Context,
) (*core.MappingNode, error) {
	if model := aws.ToString(progress.ResourceModel); model != "" {
		specState, err := a.cfnPropertiesToSpecState(model, identifier, providerContext)
		if err == nil && a.allComputedFieldsPresent(specState) {
			return specState, nil
		}
	}

	return a.fetchResourceSpecState(ctx, service, identifier, providerContext)
}

func (a *ccResourceActions) allComputedFieldsPresent(specState *core.MappingNode) bool {
	for _, field := range a.config.Meta.ComputedFields {
		value, ok := pluginutils.GetValueByPath(fmt.Sprintf("$.%s", field), specState)
		if !ok || !computedFieldPopulated(value) {
			return false
		}
	}

	return true
}

func computedFieldPopulated(value *core.MappingNode) bool {
	if value == nil {
		return false
	}

	if value.Fields != nil || value.Items != nil {
		return true
	}

	if value.Scalar == nil {
		return false
	}

	if value.Scalar.StringValue != nil {
		return *value.Scalar.StringValue != ""
	}

	// Non-string scalars (int/float/bool) are considered populated when set.
	return value.Scalar.IntValue != nil ||
		value.Scalar.FloatValue != nil ||
		value.Scalar.BoolValue != nil
}

func (a *ccResourceActions) computedFieldValues(
	specState *core.MappingNode,
	identifier string,
	requestToken string,
	omittedAutoNamedFields []string,
) map[string]*core.MappingNode {
	computed := map[string]*core.MappingNode{}
	if requestToken != "" {
		computed["spec."+fieldRequestToken] = core.MappingNodeFromString(requestToken)
	}

	if identifier != "" {
		computed["spec."+fieldPrimaryIdentifier] = core.MappingNodeFromString(identifier)
	}

	if specState == nil {
		return computed
	}

	for _, field := range a.config.Meta.ComputedFields {
		for _, path := range computedPathsForField(a.config.Schema, field) {
			if value, ok := pluginutils.GetValueByPath(fmt.Sprintf("$.%s", path), specState); ok {
				specFieldKey := fmt.Sprintf("spec.%s", path)
				computed[specFieldKey] = value
			}
		}
	}

	// Auto-named fields the user omitted carry the AWS-assigned (or engine
	// generated) value in the post-operation state; they are computed for this
	// deployment only.
	for _, field := range omittedAutoNamedFields {
		if value, ok := pluginutils.GetValueByPath(fmt.Sprintf("$.%s", field), specState); ok &&
			computedFieldPopulated(value) {
			computed[fmt.Sprintf("spec.%s", field)] = value
		}
	}
	return computed
}

func clientToken(prefix, instanceID, resourceID string) string {
	return sanitiseClientToken(prefix + "-" + instanceID + "-" + resourceID)
}

func hashedClientToken(prefix string, parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return sanitiseClientToken(prefix + "-" + hex.EncodeToString(sum[:]))
}

// Resolves the Cloud Control identifier from a spec, preferring the engine-owned
// __ccPrimaryIdentifier captured at deploy time. When that is absent (e.g. reconciling an
// existing resource by spec), it reconstructs the identifier from the user-facing id
// field(s): a single field directly, or a composite as the ordered components joined with
// "|" (Cloud Control's composite identifier delimiter). A composite returns "" unless
// every component is present.
func primaryIdentifier(spec *core.MappingNode, meta CCResourceMeta) string {
	if value, ok := pluginutils.GetValueByPath(fmt.Sprintf("$.%s", fieldPrimaryIdentifier), spec); ok {
		if id := core.StringValue(value); id != "" {
			return id
		}
	}

	fields := meta.PrimaryIdentifierFields
	if len(fields) == 0 {
		if meta.PrimaryIdentifierField == "" {
			return ""
		}
		fields = []string{meta.PrimaryIdentifierField}
	}

	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		value, ok := pluginutils.GetValueByPath(fmt.Sprintf("$.%s", field), spec)
		if !ok {
			return ""
		}
		part := core.StringValue(value)
		if part == "" {
			return ""
		}
		parts = append(parts, part)
	}

	return strings.Join(parts, "|")
}

func sanitiseClientToken(token string) string {
	cleaned := make([]rune, 0, len(token))
	for _, r := range token {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-':
			cleaned = append(cleaned, r)
		default:
			cleaned = append(cleaned, '-')
		}
	}
	if len(cleaned) > clientTokenMaxLength {
		cleaned = cleaned[:clientTokenMaxLength]
	}
	return string(cleaned)
}

// The spec paths to report as computed for a field listed in a resource type's metadata.
//
// The metadata lists top-level field names, but CloudFormation marks read-only properties
// at whatever depth they occur, and the generated schema follows suit. AWS::RDS::DBCluster
// is one such case: the metadata lists "masterUserSecret" while only its "secretArn"
// attribute is computed, the rest of the object being author-supplied.
//
// Reporting the parent makes the engine reject the whole deployment, since a computed
// value it never declared comes back. So a field that is not itself computed is expanded
// into whichever of its descendants are, and a field with no computed descendants reports
// nothing rather than claiming the parent.
func computedPathsForField(
	schema *provider.ResourceDefinitionsSchema,
	field string,
) []string {
	fieldSchema := schemaForFieldPath(schema, field)
	if fieldSchema == nil || fieldSchema.Computed {
		// Nothing known about the field is still reported under its own name, so a
		// resource type whose schema the engine has not augmented behaves as before.
		return []string{field}
	}

	return computedDescendantPaths(fieldSchema, field)
}

func schemaForFieldPath(
	schema *provider.ResourceDefinitionsSchema,
	field string,
) *provider.ResourceDefinitionsSchema {
	if schema == nil || schema.Attributes == nil {
		return nil
	}

	current := schema
	for _, segment := range strings.Split(field, ".") {
		if current.Attributes == nil {
			return nil
		}
		next, ok := current.Attributes[segment]
		if !ok {
			return nil
		}
		current = next
	}

	return current
}

// Depth-first so the returned paths are stable for a given schema.
func computedDescendantPaths(
	schema *provider.ResourceDefinitionsSchema,
	prefix string,
) []string {
	if schema == nil || schema.Attributes == nil {
		return nil
	}

	names := make([]string, 0, len(schema.Attributes))
	for name := range schema.Attributes {
		names = append(names, name)
	}
	sort.Strings(names)

	paths := []string{}
	for _, name := range names {
		attr := schema.Attributes[name]
		path := prefix + "." + name
		if attr.Computed {
			paths = append(paths, path)
			continue
		}
		paths = append(paths, computedDescendantPaths(attr, path)...)
	}

	return paths
}
