package cloudcontrol

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscc "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/wI2L/jsondiff"
)

func (a *ccResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	service, err := a.getCloudControlService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	identifier := a.currentIdentifier(input)
	if identifier == "" {
		return nil, fmt.Errorf(
			"cannot update %s: no Cloud Control identifier found in current state",
			a.config.CFNType,
		)
	}

	if err := a.validateResolvedSpec(input); err != nil {
		return nil, err
	}

	currentProperties, err := a.fetchCurrentProperties(ctx, service, identifier)
	if err != nil {
		return nil, err
	}

	patchDocument, hasChanges, err := a.buildPatchDocument(input, currentProperties)
	if err != nil {
		return nil, err
	}
	if !hasChanges {
		// Nothing to update, but the computed fields must still be captured into
		// state from the live resource. An empty token short-circuits the wait.
		return a.captureComputedFields(ctx, service, input, "", identifier)
	}

	output, err := a.submitUpdate(ctx, service, identifier, patchDocument, input)
	if err != nil {
		return nil, err
	}

	return a.captureComputedFields(
		ctx,
		service,
		input,
		aws.ToString(output.ProgressEvent.RequestToken),
		identifier,
	)
}

func (a *ccResourceActions) fetchCurrentProperties(
	ctx context.Context,
	service cloudcontrolservice.Service,
	identifier string,
) (string, error) {
	getResource := pluginutils.RetryableReturnValue(
		func(ctx context.Context, in *awscc.GetResourceInput) (*awscc.GetResourceOutput, error) {
			return service.GetResource(ctx, in)
		},
		isCCErrorRetryable,
	)
	current, err := getResource(ctx, &awscc.GetResourceInput{
		TypeName:   aws.String(a.config.CFNType),
		Identifier: aws.String(identifier),
	})
	if err != nil {
		return "", err
	}

	if current.ResourceDescription == nil {
		return "", nil
	}
	return aws.ToString(current.ResourceDescription.Properties), nil
}

func (a *ccResourceActions) submitUpdate(
	ctx context.Context,
	service cloudcontrolservice.Service,
	identifier string,
	patchDocument string,
	input *provider.ResourceDeployInput,
) (*awscc.UpdateResourceOutput, error) {
	updateResource := pluginutils.RetryableReturnValue(
		func(
			ctx context.Context,
			in *awscc.UpdateResourceInput,
		) (*awscc.UpdateResourceOutput, error) {
			return service.UpdateResource(ctx, in)
		},
		isCCErrorRetryable,
	)
	output, err := updateResource(ctx, &awscc.UpdateResourceInput{
		TypeName:      aws.String(a.config.CFNType),
		Identifier:    aws.String(identifier),
		PatchDocument: aws.String(patchDocument),
		ClientToken: aws.String(
			hashedClientToken("update", input.InstanceID, input.ResourceID, patchDocument),
		),
	})
	if err != nil {
		return nil, err
	}

	if output.ProgressEvent == nil {
		return nil, fmt.Errorf(
			"cloud control update for %s returned no progress event",
			a.config.CFNType,
		)
	}
	return output, nil
}

func (a *ccResourceActions) currentIdentifier(input *provider.ResourceDeployInput) string {
	if input.Changes == nil {
		return ""
	}

	currentState := input.Changes.AppliedResourceInfo.CurrentResourceState
	if currentState == nil {
		return ""
	}

	return primaryIdentifier(currentState.SpecData, a.config.Meta)
}

// Both sides have computed and write-only fields removed so the patch never targets
// properties Cloud Control will not accept, and any operation rooted at a create-only
// property is dropped (those route through replacement, not update).
func (a *ccResourceActions) buildPatchDocument(
	input *provider.ResourceDeployInput,
	currentProperties string,
) (string, bool, error) {
	currentCFN, err := a.managedCurrentCFNNode(currentProperties)
	if err != nil {
		return "", false, err
	}
	desiredCFN := a.managedDesiredCFNNode(input)

	currentJSON, err := json.Marshal(currentCFN)
	if err != nil {
		return "", false, err
	}
	desiredJSON, err := json.Marshal(desiredCFN)
	if err != nil {
		return "", false, err
	}

	patch, err := jsondiff.CompareJSON(currentJSON, desiredJSON)
	if err != nil {
		return "", false, err
	}

	patch = a.dropCreateOnlyOperations(patch)
	patch = a.dropUnmanagedOperations(patch)
	if len(patch) == 0 {
		return "", false, nil
	}

	patchJSON, err := json.Marshal(patch)
	if err != nil {
		return "", false, err
	}
	return string(patchJSON), true, nil
}

func (a *ccResourceActions) managedDesiredCFNNode(
	input *provider.ResourceDeployInput,
) *core.MappingNode {
	desired := stripComputedFields(resolvedSpec(input), a.config.Schema)
	if desired == nil {
		desired = &core.MappingNode{Fields: map[string]*core.MappingNode{}}
	}

	if a.config.Meta.IsTaggable() {
		a.injectTags(desired, input)
	}

	a.stripWriteOnlyFields(desired)
	return SpecToCFN(desired, a.config.Schema, a.config.Meta)
}

func (a *ccResourceActions) managedCurrentCFNNode(
	currentProperties string,
) (*core.MappingNode, error) {
	node, err := cfnJSONToMappingNode(currentProperties)
	if err != nil {
		return nil, err
	}

	current := stripComputedFields(
		CFNToSpec(node, a.config.Schema, a.config.Meta),
		a.config.Schema,
	)
	if current == nil {
		current = &core.MappingNode{Fields: map[string]*core.MappingNode{}}
	}

	a.stripWriteOnlyFields(current)
	return SpecToCFN(current, a.config.Schema, a.config.Meta), nil
}

func (a *ccResourceActions) dropCreateOnlyOperations(patch jsondiff.Patch) jsondiff.Patch {
	if len(a.config.Meta.CreateOnlyFields) == 0 {
		return patch
	}
	createOnlyRoots := make(map[string]bool, len(a.config.Meta.CreateOnlyFields))
	for _, field := range a.config.Meta.CreateOnlyFields {
		createOnlyRoots["/"+upperFirst(field)] = true
	}

	kept := make(jsondiff.Patch, 0, len(patch))
	for _, operation := range patch {
		if createOnlyRoots[patchRoot(operation.Path)] {
			continue
		}
		kept = append(kept, operation)
	}
	return kept
}

// Read-back state can carry properties the generated schema does not model,
// typically read-only fields AWS added to the live resource type after its
// schema was vendored (e.g. AWS::Events::Rule's Id). Ops rooted at them would
// be rejected by AWS ("readOnlyProperties ... cannot be updated") or add
// state the engine does not manage, so they are dropped.
func (a *ccResourceActions) dropUnmanagedOperations(patch jsondiff.Patch) jsondiff.Patch {
	if a.config.Schema == nil || a.config.Schema.Attributes == nil {
		return patch
	}
	managedRoots := make(map[string]bool, len(a.config.Schema.Attributes))
	for name := range a.config.Schema.Attributes {
		cfnName := a.config.Meta.FieldNameOverrides[name]
		if cfnName == "" {
			cfnName = upperFirst(name)
		}
		managedRoots["/"+cfnName] = true
	}

	kept := make(jsondiff.Patch, 0, len(patch))
	for _, operation := range patch {
		if !managedRoots[patchRoot(operation.Path)] {
			continue
		}
		kept = append(kept, operation)
	}
	return kept
}

func patchRoot(path string) string {
	trimmed := strings.TrimPrefix(path, "/")
	if index := strings.IndexByte(trimmed, '/'); index >= 0 {
		trimmed = trimmed[:index]
	}
	return "/" + trimmed
}
