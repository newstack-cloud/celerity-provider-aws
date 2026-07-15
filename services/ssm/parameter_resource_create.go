package ssm

import (
	"context"
	"errors"
	"fmt"
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

func (a *parameterResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	specData := pluginutils.GetResolvedResourceSpecData(input.Changes)

	service, _, err := a.getSSMServiceWithRegion(ctx, input.ProviderContext, parameterRegionMeta(specData))
	if err != nil {
		return nil, err
	}

	putInput, name, err := parameterPutInputFromSpec(specData)
	if err != nil {
		return nil, err
	}

	// PutParameter only accepts Tags when it is not overwriting an existing parameter.
	// On create Overwrite is false, so the merged Bluelink + user tags are included in the
	// initial call rather than a separate tagging operation.
	putInput.Tags = parameterTags(input, specData)

	putOutput, err := service.PutParameter(ctx, putInput)
	if err != nil {
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

// Build a PutParameterInput from the resolved resource spec,
// returning the input and the parameter name. The value/secureValue split is driven by the
// type; schema validation guarantees exactly the matching one is set.
func parameterPutInputFromSpec(
	specData *core.MappingNode,
) (*ssm.PutParameterInput, string, error) {
	nameNode, hasName := pluginutils.GetValueByPath("$.name", specData)
	if !hasName {
		return nil, "", errors.New("name is required to create an SSM parameter")
	}
	name := core.StringValue(nameNode)

	typeNode, hasType := pluginutils.GetValueByPath("$.type", specData)
	if !hasType {
		return nil, "", errors.New("type is required to create an SSM parameter")
	}
	paramType := core.StringValue(typeNode)

	putInput := &ssm.PutParameterInput{
		Name:      aws.String(name),
		Type:      ssmtypes.ParameterType(paramType),
		Overwrite: aws.Bool(false),
	}

	if paramType == string(ssmtypes.ParameterTypeSecureString) {
		secureValue, _ := pluginutils.GetValueByPath("$.secureValue", specData)
		putInput.Value = aws.String(core.StringValue(secureValue))
	} else {
		value, _ := pluginutils.GetValueByPath("$.value", specData)
		putInput.Value = aws.String(core.StringValue(value))
	}

	if description, ok := pluginutils.GetValueByPath("$.description", specData); ok {
		putInput.Description = aws.String(core.StringValue(description))
	}
	if tier, ok := pluginutils.GetValueByPath("$.tier", specData); ok {
		putInput.Tier = ssmtypes.ParameterTier(core.StringValue(tier))
	}
	if keyID, ok := pluginutils.GetValueByPath("$.keyId", specData); ok {
		putInput.KeyId = aws.String(core.StringValue(keyID))
	}
	if allowedPattern, ok := pluginutils.GetValueByPath("$.allowedPattern", specData); ok {
		putInput.AllowedPattern = aws.String(core.StringValue(allowedPattern))
	}
	if dataType, ok := pluginutils.GetValueByPath("$.dataType", specData); ok {
		putInput.DataType = aws.String(core.StringValue(dataType))
	}
	if policies, ok := pluginutils.GetValueByPath("$.policies", specData); ok {
		putInput.Policies = aws.String(core.StringValue(policies))
	}

	return putInput, name, nil
}

func parameterTags(
	input *provider.ResourceDeployInput,
	specData *core.MappingNode,
) []ssmtypes.Tag {
	merged := utils.MergeBluelinkTagsWithUserTags(input, userTagsFromSpec(specData))

	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	tags := make([]ssmtypes.Tag, 0, len(merged))
	for _, key := range keys {
		tags = append(tags, ssmtypes.Tag{
			Key:   aws.String(key),
			Value: aws.String(merged[key]),
		})
	}
	return tags
}

func userTagsFromSpec(specData *core.MappingNode) map[string]string {
	userTags := map[string]string{}
	if tagsNode, ok := pluginutils.GetValueByPath("$.tags", specData); ok && tagsNode != nil {
		for key, value := range tagsNode.Fields {
			userTags[key] = core.StringValue(value)
		}
	}
	return userTags
}

func getParameterARN(
	ctx context.Context,
	service ssmservice.Service,
	name string,
) (string, error) {
	output, err := service.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(false),
	})
	if err != nil {
		return "", err
	}
	if output.Parameter == nil {
		return "", fmt.Errorf("parameter %q was not found after creation", name)
	}
	return aws.ToString(output.Parameter.ARN), nil
}
