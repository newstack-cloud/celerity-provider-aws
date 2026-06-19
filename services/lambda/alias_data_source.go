package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// AliasDataSource returns a data source implementation for an AWS Lambda Alias.
func AliasDataSource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.DataSource {
	bundledExample, _ := examples.ReadFile("examples/datasources/lambda_alias_basic.md")

	lambdaAliasFetcher := &lambdaAliasDataSourceFetcher{
		lambdaServiceFactory,
		awsConfigStore,
	}
	return &providerv1.DataSourceDefinition{
		Type:             "aws/lambda/alias",
		Label:            "AWS Lambda Alias",
		PlainTextSummary: "A data source for retrieving an AWS Lambda alias.",
		FormattedDescription: "The data source type used to define a [Lambda alias](https://docs.aws.amazon.com/lambda/latest/api/API_GetAlias.html) " +
			"managed externally in AWS.",
		MarkdownExamples: []string{
			string(bundledExample),
		},
		Fields: lambdaAliasDataSourceSchema(),
		FilterFields: map[string]*provider.DataSourceFilterSchema{
			"functionName": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The name of the Lambda function.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
			"name": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The name of the alias.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
		},
		FetchFunc: lambdaAliasFetcher.Fetch,
	}
}

type lambdaAliasDataSourceFetcher struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaAliasDataSourceFetcher) getLambdaService(
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

func (l *lambdaAliasDataSourceFetcher) Fetch(
	ctx context.Context,
	input *provider.DataSourceFetchInput,
) (*provider.DataSourceFetchOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda service: %w", err)
	}

	// Extract filter values
	functionName := extractFunctionNameFromFilters(input.DataSourceWithResolvedSubs.Filter)
	aliasName := extractAliasNameFromFilters(input.DataSourceWithResolvedSubs.Filter)

	// Validate required filters
	if functionName == nil {
		return nil, errors.New("function_name filter is required")
	}
	if aliasName == nil {
		return nil, errors.New("name filter is required")
	}

	// Get the alias
	getAliasInput := &lambda.GetAliasInput{
		FunctionName: aws.String(core.StringValue(functionName)),
		Name:         aws.String(core.StringValue(aliasName)),
	}

	aliasOutput, err := lambdaService.GetAlias(ctx, getAliasInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda alias: %w", err)
	}

	// Build the data from the alias output
	data := l.createBaseData(aliasOutput, core.StringValue(functionName))

	// Add optional configurations
	err = l.addOptionalConfigurationsToData(aliasOutput, data.Fields)
	if err != nil {
		return nil, err
	}

	return &provider.DataSourceFetchOutput{
		Data: data.Fields,
	}, nil
}

func (l *lambdaAliasDataSourceFetcher) createBaseData(
	aliasOutput *lambda.GetAliasOutput,
	functionName string,
) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":             core.MappingNodeFromString(aws.ToString(aliasOutput.AliasArn)),
			"functionName":    core.MappingNodeFromString(functionName),
			"functionVersion": core.MappingNodeFromString(aws.ToString(aliasOutput.FunctionVersion)),
			"invokeArn":       core.MappingNodeFromString(aws.ToString(aliasOutput.AliasArn)),
			"name":            core.MappingNodeFromString(aws.ToString(aliasOutput.Name)),
		},
	}
}

func (l *lambdaAliasDataSourceFetcher) addOptionalConfigurationsToData(
	aliasOutput *lambda.GetAliasOutput,
	targetData map[string]*core.MappingNode,
) error {
	extractors := []pluginutils.OptionalValueExtractor[*lambda.GetAliasOutput]{
		aliasDescriptionValueExtractor(),
		{
			Name: "routingConfig.additionalVersionWeights",
			Condition: func(output *lambda.GetAliasOutput) bool {
				return output.RoutingConfig != nil && len(output.RoutingConfig.AdditionalVersionWeights) > 0
			},
			Fields: []string{"routingConfig.additionalVersionWeights"},
			Values: func(output *lambda.GetAliasOutput) ([]*core.MappingNode, error) {
				// Convert the routing config to JSON string
				routingConfig := map[string]float64{}
				maps.Copy(
					routingConfig,
					output.RoutingConfig.AdditionalVersionWeights,
				)

				jsonBytes, err := json.Marshal(routingConfig)
				if err != nil {
					return nil, fmt.Errorf("failed to marshal routing config: %w", err)
				}

				return []*core.MappingNode{
					core.MappingNodeFromString(string(jsonBytes)),
				}, nil
			},
		},
	}

	return pluginutils.RunOptionalValueExtractors(
		aliasOutput,
		targetData,
		extractors,
	)
}

func extractFunctionNameFromFilters(
	filters *provider.ResolvedDataSourceFilters,
) *core.MappingNode {
	return pluginutils.ExtractMatchFromFilters(
		filters,
		"functionName",
	)
}

func extractAliasNameFromFilters(
	filters *provider.ResolvedDataSourceFilters,
) *core.MappingNode {
	return pluginutils.ExtractMatchFromFilters(
		filters,
		"name",
	)
}
