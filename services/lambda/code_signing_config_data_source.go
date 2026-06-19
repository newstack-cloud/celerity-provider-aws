package lambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// CodeSigningConfigDataSource returns a data source implementation for an AWS Lambda Code Signing Configuration.
func CodeSigningConfigDataSource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.DataSource {
	bundledExample, _ := examples.ReadFile("examples/datasources/lambda_code_signing_config_basic.md")

	lambdaCodeSigningConfigFetcher := &lambdaCodeSigningConfigDataSourceFetcher{
		lambdaServiceFactory,
		awsConfigStore,
	}
	return &providerv1.DataSourceDefinition{
		Type:             "aws/lambda/codeSigningConfig",
		Label:            "AWS Lambda Code Signing Configuration",
		PlainTextSummary: "A data source for retrieving an AWS Lambda code signing configuration.",
		FormattedDescription: "The data source type used to define a [Lambda code signing configuration](https://docs.aws.amazon.com/lambda/latest/dg/configuration-codesigning.html) " +
			"managed externally in AWS.",
		MarkdownExamples: []string{
			string(bundledExample),
		},
		Fields: lambdaCodeSigningConfigDataSourceSchema(),
		FilterFields: map[string]*provider.DataSourceFilterSchema{
			"arn": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The ARN of the Lambda code signing configuration to retrieve.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
		},
		FetchFunc: lambdaCodeSigningConfigFetcher.Fetch,
	}
}

type lambdaCodeSigningConfigDataSourceFetcher struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaCodeSigningConfigDataSourceFetcher) getLambdaService(
	ctx context.Context,
	input *provider.DataSourceFetchInput,
) (lambdaservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		input.ProviderContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.lambdaServiceFactory(awsConfig, input.ProviderContext), nil
}

func (l *lambdaCodeSigningConfigDataSourceFetcher) Fetch(
	ctx context.Context,
	input *provider.DataSourceFetchInput,
) (*provider.DataSourceFetchOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda service: %w", err)
	}

	arn := extractARNFromFilters(input.DataSourceWithResolvedSubs.Filter)
	if arn == nil {
		return nil, errors.New("arn filter is required")
	}

	getCodeSigningConfigInput := &lambda.GetCodeSigningConfigInput{
		CodeSigningConfigArn: aws.String(core.StringValue(arn)),
	}

	codeSigningConfigOutput, err := lambdaService.GetCodeSigningConfig(ctx, getCodeSigningConfigInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda code signing config: %w", err)
	}

	data := l.createBaseData(codeSigningConfigOutput)

	err = l.addOptionalConfigurationsToData(codeSigningConfigOutput, data.Fields)
	if err != nil {
		return nil, err
	}

	listTagsInput := &lambda.ListTagsInput{
		Resource: aws.String(core.StringValue(arn)),
	}
	tagsOutput, err := lambdaService.ListTags(ctx, listTagsInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda code signing config tags: %w", err)
	}

	if tagsOutput != nil && len(tagsOutput.Tags) > 0 {
		tagNodes := make([]*core.MappingNode, 0, len(tagsOutput.Tags))
		for key, value := range tagsOutput.Tags {
			tagNodes = append(tagNodes, &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"key":   core.MappingNodeFromString(key),
					"value": core.MappingNodeFromString(value),
				},
			})
		}
		data.Fields["tags"] = &core.MappingNode{
			Items: tagNodes,
		}
	}

	return &provider.DataSourceFetchOutput{
		Data: data.Fields,
	}, nil
}

func (l *lambdaCodeSigningConfigDataSourceFetcher) createBaseData(
	output *lambda.GetCodeSigningConfigOutput,
) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                 core.MappingNodeFromString(aws.ToString(output.CodeSigningConfig.CodeSigningConfigArn)),
			"codeSigningConfigId": core.MappingNodeFromString(aws.ToString(output.CodeSigningConfig.CodeSigningConfigId)),
		},
	}
}

func (l *lambdaCodeSigningConfigDataSourceFetcher) addOptionalConfigurationsToData(
	output *lambda.GetCodeSigningConfigOutput,
	targetData map[string]*core.MappingNode,
) error {
	extractors := []pluginutils.OptionalValueExtractor[*lambda.GetCodeSigningConfigOutput]{
		{
			Name: "description",
			Condition: func(output *lambda.GetCodeSigningConfigOutput) bool {
				return output.CodeSigningConfig.Description != nil
			},
			Fields: []string{"description"},
			Values: func(output *lambda.GetCodeSigningConfigOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(aws.ToString(output.CodeSigningConfig.Description)),
				}, nil
			},
		},
		{
			Name: "allowedPublishers.signingProfileVersionArns",
			Condition: func(output *lambda.GetCodeSigningConfigOutput) bool {
				return output.CodeSigningConfig.AllowedPublishers != nil && len(output.CodeSigningConfig.AllowedPublishers.SigningProfileVersionArns) > 0
			},
			Fields: []string{"allowedPublishers.signingProfileVersionArns"},
			Values: func(output *lambda.GetCodeSigningConfigOutput) ([]*core.MappingNode, error) {
				arnNodes := make([]*core.MappingNode, len(output.CodeSigningConfig.AllowedPublishers.SigningProfileVersionArns))
				for i, arn := range output.CodeSigningConfig.AllowedPublishers.SigningProfileVersionArns {
					arnNodes[i] = core.MappingNodeFromString(arn)
				}
				return []*core.MappingNode{
					{Items: arnNodes},
				}, nil
			},
		},
		{
			Name: "codeSigningPolicies.untrustedArtifactOnDeployment",
			Condition: func(output *lambda.GetCodeSigningConfigOutput) bool {
				return output.CodeSigningConfig.CodeSigningPolicies != nil && output.CodeSigningConfig.CodeSigningPolicies.UntrustedArtifactOnDeployment != ""
			},
			Fields: []string{"codeSigningPolicies.untrustedArtifactOnDeployment"},
			Values: func(output *lambda.GetCodeSigningConfigOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(string(output.CodeSigningConfig.CodeSigningPolicies.UntrustedArtifactOnDeployment)),
				}, nil
			},
		},
	}

	return pluginutils.RunOptionalValueExtractors(output, targetData, extractors)
}

func extractARNFromFilters(
	filters *provider.ResolvedDataSourceFilters,
) *core.MappingNode {
	return pluginutils.ExtractMatchFromFilters(
		filters,
		"arn",
	)
}
