package ssm

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (a *parameterResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	service, _, err := a.getSSMServiceWithRegion(ctx, input.ProviderContext, nil)
	if err != nil {
		return nil, err
	}

	nameValue, hasName := pluginutils.GetValueByPath("$.name", input.CurrentResourceSpec)
	if !hasName {
		return nil, errors.New("name is required to get external state for an SSM parameter")
	}
	name := core.StringValue(nameValue)

	// WithDecryption returns the plaintext for SecureString parameters (a no-op for the
	// plaintext types) so the value can be diffed for drift. This requires the provider's
	// principal to hold kms:Decrypt on the encrypting key.
	getOutput, err := service.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		if _, ok := errors.AsType[*ssmtypes.ParameterNotFound](err); ok {
			return emptyExternalState(), nil
		}
		return nil, err
	}
	if getOutput.Parameter == nil {
		return emptyExternalState(), nil
	}
	parameter := getOutput.Parameter

	metadata, err := describeParameter(ctx, service, name)
	if err != nil {
		return nil, err
	}

	tags, err := getParameterTags(ctx, service, name)
	if err != nil {
		return nil, err
	}

	fields := map[string]*core.MappingNode{
		"name":    core.MappingNodeFromString(name),
		"type":    core.MappingNodeFromString(string(parameter.Type)),
		"arn":     core.MappingNodeFromString(aws.ToString(parameter.ARN)),
		"version": core.MappingNodeFromInt(int(parameter.Version)),
	}

	if parameter.Type == ssmtypes.ParameterTypeSecureString {
		fields["secureValue"] = core.MappingNodeFromString(aws.ToString(parameter.Value))
	} else {
		fields["value"] = core.MappingNodeFromString(aws.ToString(parameter.Value))
	}

	if dataType := aws.ToString(parameter.DataType); dataType != "" {
		fields["dataType"] = core.MappingNodeFromString(dataType)
	}

	if metadata != nil {
		if description := aws.ToString(metadata.Description); description != "" {
			fields["description"] = core.MappingNodeFromString(description)
		}
		if tier := string(metadata.Tier); tier != "" {
			fields["tier"] = core.MappingNodeFromString(tier)
		}
		if keyID := aws.ToString(metadata.KeyId); keyID != "" {
			fields["keyId"] = core.MappingNodeFromString(keyID)
		}
		if allowedPattern := aws.ToString(metadata.AllowedPattern); allowedPattern != "" {
			fields["allowedPattern"] = core.MappingNodeFromString(allowedPattern)
		}
	}

	if len(tags.Fields) > 0 {
		fields["tags"] = tags
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{Fields: fields},
	}, nil
}

func emptyExternalState() *provider.ResourceGetExternalStateOutput {
	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
	}
}

func describeParameter(
	ctx context.Context,
	service ssmservice.Service,
	name string,
) (*ssmtypes.ParameterMetadata, error) {
	output, err := service.DescribeParameters(ctx, &ssm.DescribeParametersInput{
		ParameterFilters: []ssmtypes.ParameterStringFilter{
			{
				Key:    aws.String("Name"),
				Option: aws.String("Equals"),
				Values: []string{name},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if len(output.Parameters) == 0 {
		return nil, nil
	}

	return &output.Parameters[0], nil
}

func getParameterTags(
	ctx context.Context,
	service ssmservice.Service,
	name string,
) (*core.MappingNode, error) {
	output, err := service.ListTagsForResource(ctx, &ssm.ListTagsForResourceInput{
		ResourceType: ssmtypes.ResourceTypeForTaggingParameter,
		ResourceId:   aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	tags := &core.MappingNode{Fields: map[string]*core.MappingNode{}}
	for _, tag := range output.TagList {
		key := aws.ToString(tag.Key)
		if utils.IsBluelinkTag(key, utils.DefaultBluelinkTagPrefix) {
			continue
		}
		tags.Fields[key] = core.MappingNodeFromString(aws.ToString(tag.Value))
	}
	return tags, nil
}
