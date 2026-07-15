package ssm

import (
	"context"
	"errors"

	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (a *parameterTreeResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	specData := pluginutils.GetResolvedResourceSpecData(input.Changes)

	service, _, err := a.getSSMServiceWithRegion(ctx, input.ProviderContext, parameterRegionMeta(specData))
	if err != nil {
		return nil, err
	}

	path, err := parameterTreePath(specData)
	if err != nil {
		return nil, err
	}

	desired := parameterTreeEntries(specData)
	for _, key := range sortedParameterTreeKeys(desired) {
		putInput := parameterTreePutInput(path, key, desired[key], specData)
		// PutParameter only accepts Tags when it is not overwriting an existing parameter.
		// On create Overwrite is false, so the merged Bluelink + user tags are included in
		// the initial call rather than a separate tagging operation.
		putInput.Tags = parameterTags(input, specData)

		if _, err := service.PutParameter(ctx, putInput); err != nil {
			if _, exists := errors.AsType[*ssmtypes.ParameterAlreadyExists](err); !exists {
				return nil, err
			}
			// A parameter already at this name is either left over from a retried partial
			// create (already holding this value) or was written out-of-band before the
			// tree existed. Either way the existing value is preserved — the tree never
			// overwrites a value it has no record of applying — and only tags are synced
			// to mark ownership.
			name := parameterTreeFullName(path, key)
			if err := reconcileParameterTagsForName(ctx, service, input, specData, name); err != nil {
				return nil, err
			}
		}
	}

	metadataByKey, err := describeManagedTreeParameters(ctx, service, path, desired)
	if err != nil {
		return nil, err
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.parameters": parameterTreeComputedParameters(metadataByKey, desired),
		},
	}, nil
}
