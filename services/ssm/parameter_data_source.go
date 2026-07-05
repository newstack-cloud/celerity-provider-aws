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
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// ParameterDataSource returns a data source implementation for an AWS Systems Manager
// (SSM) parameter, used to reference an existing parameter from a blueprint.
func ParameterDataSource(
	ssmServiceFactory pluginutils.ServiceFactory[*aws.Config, ssmservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.DataSource {
	fetcher := &parameterDataSourceFetcher{
		ssmServiceFactory: ssmServiceFactory,
		awsConfigStore:    awsConfigStore,
	}

	bundledExample, _ := examples.ReadFile("examples/datasources/parameter.md")

	return &providerv1.DataSourceDefinition{
		Type:             "aws/ssm/parameter",
		Label:            "AWS SSM Parameter",
		PlainTextSummary: "A data source for retrieving an AWS Systems Manager parameter.",
		FormattedDescription: "The data source type used to define an [SSM parameter](https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_GetParameter.html) " +
			"managed externally in AWS. SecureString values are never returned.",
		Fields:           parameterDataSourceSchema(),
		FetchFunc:        fetcher.Fetch,
		MarkdownExamples: []string{string(bundledExample)},
		FilterFields: map[string]*provider.DataSourceFilterSchema{
			"name": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The name of the parameter to retrieve.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
			"region": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The AWS region the parameter resides in.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
		},
	}
}

type parameterDataSourceFetcher struct {
	ssmServiceFactory pluginutils.ServiceFactory[*aws.Config, ssmservice.Service]
	awsConfigStore    pluginutils.ServiceConfigStore[*aws.Config]
}

func (f *parameterDataSourceFetcher) getSSMService(
	ctx context.Context,
	input *provider.DataSourceFetchInput,
) (ssmservice.Service, error) {
	meta := map[string]*core.MappingNode{
		"region": pluginutils.ExtractMatchFromFilters(
			input.DataSourceWithResolvedSubs.Filter,
			"region",
		),
	}
	awsConfig, err := f.awsConfigStore.FromProviderContext(ctx, input.ProviderContext, meta)
	if err != nil {
		return nil, err
	}
	return f.ssmServiceFactory(awsConfig, input.ProviderContext), nil
}

func (f *parameterDataSourceFetcher) Fetch(
	ctx context.Context,
	input *provider.DataSourceFetchInput,
) (*provider.DataSourceFetchOutput, error) {
	service, err := f.getSSMService(ctx, input)
	if err != nil {
		return nil, err
	}

	nameValue := pluginutils.ExtractMatchFromFilters(
		input.DataSourceWithResolvedSubs.Filter,
		"name",
	)
	if nameValue == nil {
		return nil, errors.New("name is required for the SSM parameter data source")
	}
	name := core.StringValue(nameValue)

	// SecureString values are never decrypted for a data source lookup.
	getOutput, err := service.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(false),
	})
	if err != nil {
		return nil, err
	}
	if getOutput.Parameter == nil {
		return nil, errors.New("no SSM parameter found matching the provided filters")
	}
	parameter := getOutput.Parameter

	metadata, err := describeParameter(ctx, service, name)
	if err != nil {
		return nil, err
	}

	fields := map[string]*core.MappingNode{
		"name":    core.MappingNodeFromString(name),
		"arn":     core.MappingNodeFromString(aws.ToString(parameter.ARN)),
		"type":    core.MappingNodeFromString(string(parameter.Type)),
		"version": core.MappingNodeFromInt(int(parameter.Version)),
	}

	if parameter.Type != ssmtypes.ParameterTypeSecureString {
		fields["value"] = core.MappingNodeFromString(aws.ToString(parameter.Value))
	}
	if dataType := aws.ToString(parameter.DataType); dataType != "" {
		fields["dataType"] = core.MappingNodeFromString(dataType)
	}

	if metadata != nil {
		if tier := string(metadata.Tier); tier != "" {
			fields["tier"] = core.MappingNodeFromString(tier)
		}
		if keyID := aws.ToString(metadata.KeyId); keyID != "" {
			fields["keyId"] = core.MappingNodeFromString(keyID)
		}
		if description := aws.ToString(metadata.Description); description != "" {
			fields["description"] = core.MappingNodeFromString(description)
		}
		if allowedPattern := aws.ToString(metadata.AllowedPattern); allowedPattern != "" {
			fields["allowedPattern"] = core.MappingNodeFromString(allowedPattern)
		}
	}

	return &provider.DataSourceFetchOutput{
		Data: fields,
	}, nil
}
