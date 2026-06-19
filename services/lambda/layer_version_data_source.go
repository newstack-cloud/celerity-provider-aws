package lambda

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// LayerVersionDataSource returns a data source implementation for an AWS Lambda Layer Version.
func LayerVersionDataSource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.DataSource {
	bundledExample, _ := examples.ReadFile("examples/datasources/lambda_layer_version_basic.md")

	lambdaLayerVersionFetcher := &lambdaLayerVersionDataSourceFetcher{
		lambdaServiceFactory,
		awsConfigStore,
	}
	return &providerv1.DataSourceDefinition{
		Type:             "aws/lambda/layerVersion",
		Label:            "AWS Lambda Layer Version",
		PlainTextSummary: "A data source for retrieving an AWS Lambda layer version.",
		FormattedDescription: "The data source type used to define a [Lambda layer version](https://docs.aws.amazon.com/lambda/latest/api/API_GetLayerVersion.html) " +
			"managed externally in AWS.",
		MarkdownExamples: []string{
			string(bundledExample),
		},
		Fields: lambdaLayerVersionDataSourceSchema(),
		FilterFields: map[string]*provider.DataSourceFilterSchema{
			"layerName": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The name or ARN of the layer.",
				FormattedDescription: "The name or ARN of the layer. " +
					"For example: `my-layer`, `arn:aws:lambda:us-east-2:123456789012:layer:my-layer`.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
			"versionNumber": {
				Type:        provider.DataSourceFilterSearchValueTypeInteger,
				Description: "The version number of the layer.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
			"region": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The region of the Lambda layer version to retrieve. Defaults to the region of the provider.",
				FormattedDescription: "The [region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) " +
					"of the Lambda layer version to retrieve. Defaults to the region of the provider.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
		},
		FetchFunc: lambdaLayerVersionFetcher.Fetch,
	}
}

type lambdaLayerVersionDataSourceFetcher struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaLayerVersionDataSourceFetcher) getLambdaService(
	ctx context.Context,
	input *provider.DataSourceFetchInput,
) (lambdaservice.Service, error) {
	meta := map[string]*core.MappingNode{
		"region": extractRegionFromFilters(input.DataSourceWithResolvedSubs.Filter),
	}
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		input.ProviderContext,
		meta,
	)
	if err != nil {
		return nil, err
	}

	return l.lambdaServiceFactory(awsConfig, input.ProviderContext), nil
}

func (l *lambdaLayerVersionDataSourceFetcher) Fetch(
	ctx context.Context,
	input *provider.DataSourceFetchInput,
) (*provider.DataSourceFetchOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda service: %w", err)
	}

	layerName := extractLayerNameFromFilters(input.DataSourceWithResolvedSubs.Filter)
	versionNumber := extractVersionNumberFromFilters(input.DataSourceWithResolvedSubs.Filter)

	if layerName == nil {
		return nil, fmt.Errorf("layerName filter is required for the lambda layer version data source")
	}
	if versionNumber == nil {
		return nil, fmt.Errorf("versionNumber filter is required for the lambda layer version data source")
	}

	versionInt, err := strconv.ParseInt(core.StringValue(versionNumber), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("versionNumber must be a valid integer: %w", err)
	}

	getLayerVersionInput := &lambda.GetLayerVersionInput{
		LayerName:     aws.String(core.StringValue(layerName)),
		VersionNumber: aws.Int64(versionInt),
	}

	layerVersionOutput, err := lambdaService.GetLayerVersion(ctx, getLayerVersionInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda layer version: %w", err)
	}

	data := l.createBaseData(layerVersionOutput)

	err = l.addOptionalConfigurationsToData(layerVersionOutput, data.Fields)
	if err != nil {
		return nil, err
	}

	return &provider.DataSourceFetchOutput{
		Data: data.Fields,
	}, nil
}

func (l *lambdaLayerVersionDataSourceFetcher) createBaseData(
	layerVersionOutput *lambda.GetLayerVersionOutput,
) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":     core.MappingNodeFromString(aws.ToString(layerVersionOutput.LayerArn)),
			"version": core.MappingNodeFromInt(int(layerVersionOutput.Version)),
		},
	}
}

func (l *lambdaLayerVersionDataSourceFetcher) addOptionalConfigurationsToData(
	layerVersionOutput *lambda.GetLayerVersionOutput,
	targetData map[string]*core.MappingNode,
) error {
	extractors := []pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput]{
		layerVersionDescriptionValueExtractor(),
		layerVersionLicenseInfoValueExtractor(),
		layerVersionCreatedDateValueExtractor(),
		layerVersionCompatibleRuntimesValueExtractor(),
		layerVersionCompatibleArchitecturesValueExtractor(),
		layerVersionContentValueExtractor(),
		{
			Name: "layerVersionArn",
			Condition: func(output *lambda.GetLayerVersionOutput) bool {
				return output.LayerVersionArn != nil
			},
			Fields: []string{"layerVersionArn"},
			Values: func(output *lambda.GetLayerVersionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(aws.ToString(output.LayerVersionArn)),
				}, nil
			},
		},
	}

	return pluginutils.RunOptionalValueExtractors(
		layerVersionOutput,
		targetData,
		extractors,
	)
}

func extractLayerNameFromFilters(
	filters *provider.ResolvedDataSourceFilters,
) *core.MappingNode {
	return pluginutils.ExtractMatchFromFilters(
		filters,
		"layerName",
	)
}

func extractVersionNumberFromFilters(
	filters *provider.ResolvedDataSourceFilters,
) *core.MappingNode {
	return pluginutils.ExtractMatchFromFilters(
		filters,
		"versionNumber",
	)
}
