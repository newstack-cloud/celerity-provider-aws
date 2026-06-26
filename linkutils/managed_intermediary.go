package linkutils

import (
	"context"
	"strings"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
)

// This file provides a reusable mechanism for a link to manage a single,
// per-link Cloud Control resource as an intermediary (e.g. an AWS::Lambda::Permission
// or AWS::SQS::QueueInlinePolicy that "activates" the link). The resource is
// deployed/destroyed through the link framework's ResourceService so the engine
// owns its lifecycle, drift and stabilisation. Unlike the role-access allocator
// (which packs many links' statements into a role's shared policies), this is for
// resources each link owns outright.

// The Cloud Control engine's internal computed spec
// fields returned at deploy. They are declared as the deploy's computed fields so
// the framework merges them into the intermediary's persisted state for later
// updates and destroys.
var ccInternalComputedFields = []string{"spec.__ccRequestToken", "spec.__ccPrimaryIdentifier"}

// ManagedIntermediary describes a single Cloud Control resource a link owns.
type ManagedIntermediary struct {
	// ResourceType is the Bluelink resource type to deploy, e.g. "aws/lambda/permission".
	ResourceType string
	// ResourceID is a deterministic, instance-stable identifier for the resource so
	// repeated reconciles target the same intermediary.
	ResourceID string
	// ResourceName is a human-readable logical name for the resource.
	ResourceName string
	// Spec is the resource's desired spec (free-form fields may be JSON strings; the
	// engine parses them at the Cloud Control boundary).
	Spec *core.MappingNode
}

// DeployManagedIntermediary creates or updates the link-owned resource and returns
// its persisted intermediary state (carrying the Cloud Control identifier needed
// for later updates/destroys). A nil priorState yields a create; an existing one
// yields an update.
func DeployManagedIntermediary(
	ctx context.Context,
	resourceService provider.ResourceService,
	instanceID, instanceName string,
	providerCtx provider.Context,
	priorState *state.LinkIntermediaryResourceState,
	intermediary ManagedIntermediary,
) (*state.LinkIntermediaryResourceState, error) {
	info := provider.ResourceInfo{
		ResourceID:   intermediary.ResourceID,
		ResourceName: intermediary.ResourceName,
		InstanceID:   instanceID,
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Type: &schema.ResourceTypeWrapper{Value: intermediary.ResourceType},
			Spec: intermediary.Spec,
		},
	}
	if priorState != nil {
		info.CurrentResourceState = intermediaryToResourceState(priorState)
	}

	output, err := resourceService.Deploy(
		ctx,
		intermediary.ResourceType,
		&provider.ResourceDeployServiceInput{
			DeployInput: &provider.ResourceDeployInput{
				InstanceID:      instanceID,
				InstanceName:    instanceName,
				ResourceID:      intermediary.ResourceID,
				Changes:         &provider.Changes{AppliedResourceInfo: info, ComputedFields: ccInternalComputedFields},
				ProviderContext: providerCtx,
			},
			WaitUntilStable: true,
		},
	)
	if err != nil {
		return nil, err
	}

	return &state.LinkIntermediaryResourceState{
		ResourceID:       intermediary.ResourceID,
		ResourceType:     intermediary.ResourceType,
		InstanceID:       instanceID,
		Status:           core.ResourceStatusCreated,
		PreciseStatus:    core.PreciseResourceStatusCreated,
		ResourceSpecData: mergeIntermediaryComputedFields(intermediary.Spec, output),
	}, nil
}

// DestroyManagedIntermediary removes a previously-deployed link-owned resource.
// A nil priorState is a no-op (nothing was recorded to remove).
func DestroyManagedIntermediary(
	ctx context.Context,
	resourceService provider.ResourceService,
	instanceID string,
	providerCtx provider.Context,
	priorState *state.LinkIntermediaryResourceState,
) error {
	if priorState == nil {
		return nil
	}
	return resourceService.Destroy(
		ctx,
		priorState.ResourceType,
		&provider.ResourceDestroyInput{
			InstanceID:      instanceID,
			ResourceID:      priorState.ResourceID,
			ResourceState:   intermediaryToResourceState(priorState),
			ProviderContext: providerCtx,
		},
	)
}

// FindIntermediaryState returns the recorded intermediary state with the given
// resource ID from the current link state, or nil.
func FindIntermediaryState(
	linkState *state.LinkState,
	resourceID string,
) *state.LinkIntermediaryResourceState {
	if linkState == nil {
		return nil
	}
	for _, intermediary := range linkState.IntermediaryResourceStates {
		if intermediary != nil && intermediary.ResourceID == resourceID {
			return intermediary
		}
	}
	return nil
}

func intermediaryToResourceState(
	intermediary *state.LinkIntermediaryResourceState,
) *state.ResourceState {
	return &state.ResourceState{
		ResourceID:    intermediary.ResourceID,
		Name:          intermediary.ResourceID,
		Type:          intermediary.ResourceType,
		InstanceID:    intermediary.InstanceID,
		Status:        intermediary.Status,
		PreciseStatus: intermediary.PreciseStatus,
		SpecData:      intermediary.ResourceSpecData,
	}
}

// mergeIntermediaryComputedFields combines the resolved spec with the computed
// field values returned at deploy (e.g. the Cloud Control primary identifier) so
// the persisted state carries the handle needed for later updates and destroys.
func mergeIntermediaryComputedFields(
	spec *core.MappingNode,
	output *provider.ResourceDeployOutput,
) *core.MappingNode {
	merged := &core.MappingNode{Fields: map[string]*core.MappingNode{}}
	if spec != nil {
		for key, value := range spec.Fields {
			merged.Fields[key] = value
		}
	}
	if output != nil {
		for path, value := range output.ComputedFieldValues {
			merged.Fields[strings.TrimPrefix(path, "spec.")] = value
		}
	}
	return merged
}
