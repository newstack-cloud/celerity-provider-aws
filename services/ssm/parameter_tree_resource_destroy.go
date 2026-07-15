package ssm

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (a *parameterTreeResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	// The parameters must be deleted from the region they were created in, so the region
	// recorded in the resource state (when set) re-targets the client.
	service, _, err := a.getSSMServiceWithRegion(
		ctx,
		input.ProviderContext,
		parameterRegionMeta(input.ResourceState.SpecData),
	)
	if err != nil {
		return err
	}

	path, err := parameterTreePath(input.ResourceState.SpecData)
	if err != nil {
		return err
	}

	// Only keys the tree owns per its recorded state are deleted; parameters written
	// beneath the prefix out-of-band are never touched. The computed parameters map is
	// unioned in as the resource's own record of what it applied, in case the value maps
	// are unavailable in state.
	owned := map[string]struct{}{}
	for key := range parameterTreeEntries(input.ResourceState.SpecData) {
		owned[key] = struct{}{}
	}
	parametersNode, ok := pluginutils.GetValueByPath(
		"$.parameters",
		input.ResourceState.SpecData,
	)
	if ok && parametersNode != nil {
		for key := range parametersNode.Fields {
			owned[key] = struct{}{}
		}
	}

	for _, key := range sortedParameterTreeKeys(owned) {
		if err := deleteTreeParameter(ctx, service, parameterTreeFullName(path, key)); err != nil {
			return err
		}
	}

	return nil
}
