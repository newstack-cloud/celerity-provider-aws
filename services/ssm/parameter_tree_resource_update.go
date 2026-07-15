package ssm

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Update reconciles the tree against the prior applied state, never against the cloud:
// an entry is only re-put when there is positive evidence of a blueprint change (its
// value hash differs from the hash recorded at the last apply, it is new, or it moved
// between values and secureValues). Out-of-band value writes therefore always survive
// an update, matching the tree's write-only blob semantics.
func (a *parameterTreeResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	specData := pluginutils.GetResolvedResourceSpecData(input.Changes)
	priorSpecData := pluginutils.GetCurrentResourceStateSpecData(input.Changes)

	service, _, err := a.getSSMServiceWithRegion(ctx, input.ProviderContext, parameterRegionMeta(specData))
	if err != nil {
		return nil, err
	}

	path, err := parameterTreePath(specData)
	if err != nil {
		return nil, err
	}

	desired := parameterTreeEntries(specData)
	prior := parameterTreeEntries(priorSpecData)
	storedHashes := parameterTreeStoredHashes(priorSpecData)

	settingsChanged := pluginutils.HasModifiedField(input.Changes, "spec.keyId") ||
		pluginutils.HasModifiedField(input.Changes, "spec.tier") ||
		pluginutils.HasModifiedField(input.Changes, "spec.description")
	tagsChanged := pluginutils.HasModifiedField(input.Changes, "spec.tags")

	desiredKeysSorted := sortedParameterTreeKeys(desired)

	for _, key := range desiredKeysSorted {
		entry := desired[key]
		storedHash, hasStoredHash := storedHashes[key]
		priorEntry, inPrior := prior[key]

		isNew := !hasStoredHash
		valueChanged := hasStoredHash && storedHash != parameterTreeValueHash(entry.value)
		typeMoved := inPrior && priorEntry.secure != entry.secure

		if isNew || valueChanged || typeMoved {
			if err := putTreeParameter(ctx, service, path, key, entry, specData); err != nil {
				return nil, err
			}
		} else if settingsChanged {
			if err := reputTreeParameterPreservingValue(ctx, service, path, key, entry, specData); err != nil {
				return nil, err
			}
		}

		if isNew || tagsChanged {
			name := parameterTreeFullName(path, key)
			if err := reconcileParameterTagsForName(ctx, service, input, specData, name); err != nil {
				return nil, err
			}
		}
	}

	for _, key := range sortedParameterTreeKeys(prior) {
		if _, stillDesired := desired[key]; stillDesired {
			continue
		}
		if err := deleteTreeParameter(ctx, service, parameterTreeFullName(path, key)); err != nil {
			return nil, err
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

func putTreeParameter(
	ctx context.Context,
	service ssmservice.Service,
	path string,
	key string,
	entry parameterTreeEntry,
	specData *core.MappingNode,
) error {
	putInput := parameterTreePutInput(path, key, entry, specData)
	// Tags cannot be sent alongside an overwriting PutParameter; they are reconciled
	// separately by the caller where needed.
	putInput.Overwrite = aws.Bool(true)
	_, err := service.PutParameter(ctx, putInput)
	return err
}

// Re-puts an entry whose blueprint value is unchanged
// but whose shared settings (keyId/tier/description) were modified. PutParameter requires
// a value, and sending the blueprint value would overwrite any out-of-band override, so the
// current cloud value is fetched transiently and re-applied. The fetched value is used
// only within this call, it is never stored in state or surfaced in external state, so
// the tree's intended behaviour to not read back values is unaffected.
func reputTreeParameterPreservingValue(
	ctx context.Context,
	service ssmservice.Service,
	path string,
	key string,
	entry parameterTreeEntry,
	specData *core.MappingNode,
) error {
	name := parameterTreeFullName(path, key)

	getOutput, err := service.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		if _, notFound := errors.AsType[*ssmtypes.ParameterNotFound](err); !notFound {
			return err
		}
		// The parameter was deleted out-of-band so fall back to (re)applying the
		// blueprint value with the new settings.
	} else if getOutput.Parameter != nil {
		entry.value = aws.ToString(getOutput.Parameter.Value)
	}

	return putTreeParameter(ctx, service, path, key, entry, specData)
}

func deleteTreeParameter(
	ctx context.Context,
	service ssmservice.Service,
	name string,
) error {
	_, err := service.DeleteParameter(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(name),
	})
	if err != nil {
		if _, notFound := errors.AsType[*ssmtypes.ParameterNotFound](err); notFound {
			return nil
		}
		return err
	}
	return nil
}
