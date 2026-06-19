package lambda

import (
	"context"
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

// FunctionUrlDataSource returns a data source implementation for an AWS Lambda Function URL.
func FunctionUrlDataSource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.DataSource {
	bundledExample, _ := examples.ReadFile("examples/datasources/lambda_function_url_yaml.md")

	lambdaFunctionUrlFetcher := &lambdaFunctionUrlDataSourceFetcher{
		lambdaServiceFactory,
		awsConfigStore,
	}
	return &providerv1.DataSourceDefinition{
		Type:             "aws/lambda/function_url",
		Label:            "AWS Lambda Function URL",
		PlainTextSummary: "A data source for retrieving an AWS Lambda function URL configuration.",
		FormattedDescription: "The data source type used to define a [Lambda function URL](https://docs.aws.amazon.com/lambda/latest/api/API_GetFunctionUrlConfig.html) " +
			"managed externally in AWS.",
		MarkdownExamples: []string{
			string(bundledExample),
		},
		Fields: lambdaFunctionUrlDataSourceSchema(),
		FilterFields: map[string]*provider.DataSourceFilterSchema{
			"functionName": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The name or ARN of the Lambda function to retrieve the function URL for.",
				FormattedDescription: "The name or ARN of the Lambda function to retrieve the function URL for. " +
					"Can be a function name, function ARN, or partial ARN.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
			"qualifier": {
				Type: provider.DataSourceFilterSearchValueTypeString,
				Description: "The qualifier of the Lambda function to retrieve the function URL for (e.g. an-alias). " +
					"If not provided, the function URL for the unqualified function will be retrieved.",
				FormattedDescription: "The qualifier of the Lambda function to retrieve the function URL for (e.g. `an-alias`). " +
					"If not provided, the function URL for the unqualified function will be retrieved.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
			"region": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The region of the Lambda function URL to retrieve. Defaults to the region of the provider.",
				FormattedDescription: "The [region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) " +
					"of the Lambda function URL to retrieve. Defaults to the region of the provider.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
		},
		FetchFunc: lambdaFunctionUrlFetcher.Fetch,
	}
}

type lambdaFunctionUrlDataSourceFetcher struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaFunctionUrlDataSourceFetcher) getLambdaService(
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

func (l *lambdaFunctionUrlDataSourceFetcher) Fetch(
	ctx context.Context,
	input *provider.DataSourceFetchInput,
) (*provider.DataSourceFetchOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input)
	if err != nil {
		return nil, err
	}

	functionName := extractFunctionNameFromFilters(
		input.DataSourceWithResolvedSubs.Filter,
	)
	if functionName == nil {
		return nil, fmt.Errorf(
			"function name is required for the lambda function URL data source",
		)
	}

	qualifier := extractQualifierFromFilters(input.DataSourceWithResolvedSubs.Filter)

	getFunctionUrlInput := &lambda.GetFunctionUrlConfigInput{
		FunctionName: aws.String(core.StringValue(functionName)),
	}
	if qualifier != nil {
		getFunctionUrlInput.Qualifier = aws.String(core.StringValue(qualifier))
	}

	functionUrlOutput, err := lambdaService.GetFunctionUrlConfig(ctx, getFunctionUrlInput)
	if err != nil {
		return nil, err
	}

	data := l.createBaseData(functionUrlOutput)

	err = l.addOptionalConfigurationsToData(functionUrlOutput, data.Fields)
	if err != nil {
		return nil, err
	}

	return &provider.DataSourceFetchOutput{
		Data: data.Fields,
	}, nil
}

func (l *lambdaFunctionUrlDataSourceFetcher) createBaseData(
	functionUrlOutput *lambda.GetFunctionUrlConfigOutput,
) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionUrl": core.MappingNodeFromString(
				aws.ToString(functionUrlOutput.FunctionUrl),
			),
			"functionArn": core.MappingNodeFromString(
				aws.ToString(functionUrlOutput.FunctionArn),
			),
			"authType": core.MappingNodeFromString(
				string(functionUrlOutput.AuthType),
			),
			"creationTime": core.MappingNodeFromString(
				aws.ToString(functionUrlOutput.CreationTime),
			),
			"lastModifiedTime": core.MappingNodeFromString(
				aws.ToString(functionUrlOutput.LastModifiedTime),
			),
		},
	}
}

func (l *lambdaFunctionUrlDataSourceFetcher) addOptionalConfigurationsToData(
	functionUrlOutput *lambda.GetFunctionUrlConfigOutput,
	targetData map[string]*core.MappingNode,
) error {
	configurations := []pluginutils.OptionalValueExtractor[*lambda.GetFunctionUrlConfigOutput]{
		{
			Name: "invokeMode",
			Condition: func(output *lambda.GetFunctionUrlConfigOutput) bool {
				return output.InvokeMode != ""
			},
			Fields: []string{"invokeMode"},
			Values: func(output *lambda.GetFunctionUrlConfigOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(string(output.InvokeMode)),
				}, nil
			},
		},
		{
			Name: "cors",
			Condition: func(output *lambda.GetFunctionUrlConfigOutput) bool {
				return output.Cors != nil
			},
			Fields: []string{
				"cors.allowCredentials",
				"cors.allowHeaders",
				"cors.allowMethods",
				"cors.allowOrigins",
				"cors.exposeHeaders",
				"cors.maxAge",
			},
			Values: func(output *lambda.GetFunctionUrlConfigOutput) ([]*core.MappingNode, error) {
				cors := output.Cors

				var allowCredentials *core.MappingNode
				if cors.AllowCredentials != nil {
					allowCredentials = core.MappingNodeFromBool(aws.ToBool(cors.AllowCredentials))
				}

				var allowHeaders *core.MappingNode
				if len(cors.AllowHeaders) > 0 {
					headers := make([]*core.MappingNode, len(cors.AllowHeaders))
					for i, header := range cors.AllowHeaders {
						headers[i] = core.MappingNodeFromString(header)
					}
					allowHeaders = &core.MappingNode{Items: headers}
				}

				var allowMethods *core.MappingNode
				if len(cors.AllowMethods) > 0 {
					methods := make([]*core.MappingNode, len(cors.AllowMethods))
					for i, method := range cors.AllowMethods {
						methods[i] = core.MappingNodeFromString(method)
					}
					allowMethods = &core.MappingNode{Items: methods}
				}

				var allowOrigins *core.MappingNode
				if len(cors.AllowOrigins) > 0 {
					origins := make([]*core.MappingNode, len(cors.AllowOrigins))
					for i, origin := range cors.AllowOrigins {
						origins[i] = core.MappingNodeFromString(origin)
					}
					allowOrigins = &core.MappingNode{Items: origins}
				}

				var exposeHeaders *core.MappingNode
				if len(cors.ExposeHeaders) > 0 {
					headers := make([]*core.MappingNode, len(cors.ExposeHeaders))
					for i, header := range cors.ExposeHeaders {
						headers[i] = core.MappingNodeFromString(header)
					}
					exposeHeaders = &core.MappingNode{Items: headers}
				}

				var maxAge *core.MappingNode
				if cors.MaxAge != nil {
					maxAge = core.MappingNodeFromInt(int(aws.ToInt32(cors.MaxAge)))
				}

				return []*core.MappingNode{
					allowCredentials,
					allowHeaders,
					allowMethods,
					allowOrigins,
					exposeHeaders,
					maxAge,
				}, nil
			},
		},
	}

	return pluginutils.RunOptionalValueExtractors(
		functionUrlOutput,
		targetData,
		configurations,
	)
}
