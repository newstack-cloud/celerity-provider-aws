package ssm

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (a *parameterResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	specData := pluginutils.GetResolvedResourceSpecData(input.Changes)

	service, _, err := a.getSSMServiceWithRegion(
		ctx,
		input.ProviderContext,
		parameterRegionMeta(specData),
	)
	if err != nil {
		return nil, err
	}

	putInput, name, err := parameterPutInputFromSpec(specData)
	if err != nil {
		return nil, err
	}
	// Overwrite the existing parameter value/metadata. Tags cannot be sent alongside an
	// overwriting PutParameter, so they are reconciled separately below.
	putInput.Overwrite = aws.Bool(true)

	putOutput, err := service.PutParameter(ctx, putInput)
	if err != nil {
		return nil, err
	}

	if err := a.reconcileParameterTags(ctx, service, input, specData, name); err != nil {
		return nil, err
	}

	arn, err := getParameterARN(ctx, service, name)
	if err != nil {
		return nil, err
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.arn":     core.MappingNodeFromString(arn),
			"spec.version": core.MappingNodeFromInt(int(putOutput.Version)),
		},
	}, nil
}

// Brings the parameter's tags in line with the desired set (merged
// Bluelink + user tags): adding/overwriting changed tags and removing tags no longer desired.
func (a *parameterResourceActions) reconcileParameterTags(
	ctx context.Context,
	service ssmservice.Service,
	input *provider.ResourceDeployInput,
	specData *core.MappingNode,
	name string,
) error {
	return reconcileParameterTagsForName(ctx, service, input, specData, name)
}

// Package-level form shared with the parameter tree resource, which reconciles
// tags for many parameter names against a single desired set.
func reconcileParameterTagsForName(
	ctx context.Context,
	service ssmservice.Service,
	input *provider.ResourceDeployInput,
	specData *core.MappingNode,
	name string,
) error {
	desired := utils.MergeBluelinkTagsWithUserTags(input, userTagsFromSpec(specData))

	listOutput, err := service.ListTagsForResource(ctx, &ssm.ListTagsForResourceInput{
		ResourceType: ssmtypes.ResourceTypeForTaggingParameter,
		ResourceId:   aws.String(name),
	})
	if err != nil {
		return err
	}
	current := map[string]string{}
	for _, tag := range listOutput.TagList {
		current[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	addKeys := make([]string, 0, len(desired))
	for key, value := range desired {
		if currentValue, ok := current[key]; !ok || currentValue != value {
			addKeys = append(addKeys, key)
		}
	}
	sort.Strings(addKeys)
	if len(addKeys) > 0 {
		tags := make([]ssmtypes.Tag, 0, len(addKeys))
		for _, key := range addKeys {
			tags = append(tags, ssmtypes.Tag{
				Key:   aws.String(key),
				Value: aws.String(desired[key]),
			})
		}
		if _, err := service.AddTagsToResource(ctx, &ssm.AddTagsToResourceInput{
			ResourceType: ssmtypes.ResourceTypeForTaggingParameter,
			ResourceId:   aws.String(name),
			Tags:         tags,
		}); err != nil {
			return err
		}
	}

	removeKeys := make([]string, 0, len(current))
	for key := range current {
		if _, ok := desired[key]; !ok {
			removeKeys = append(removeKeys, key)
		}
	}
	sort.Strings(removeKeys)
	if len(removeKeys) > 0 {
		if _, err := service.RemoveTagsFromResource(ctx, &ssm.RemoveTagsFromResourceInput{
			ResourceType: ssmtypes.ResourceTypeForTaggingParameter,
			ResourceId:   aws.String(name),
			TagKeys:      removeKeys,
		}); err != nil {
			return err
		}
	}

	return nil
}
