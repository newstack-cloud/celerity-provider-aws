package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// FunctionDataSource returns a data source implementation for an AWS Lambda Function.
func FunctionDataSource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.DataSource {
	bundledExample, _ := examples.ReadFile("examples/datasources/lambda_function_yaml.md")
	exportAllExample, _ := examples.ReadFile("examples/datasources/lambda_function_yaml_export_all.md")

	lambdaFunctionFetcher := &lambdaFunctionDataSourceFetcher{
		lambdaServiceFactory,
		awsConfigStore,
	}
	return &providerv1.DataSourceDefinition{
		Type:             "aws/lambda/function",
		Label:            "AWS Lambda Function",
		PlainTextSummary: "A data source for retrieving an AWS Lambda function.",
		FormattedDescription: "The data source type used to define a [Lambda function](https://docs.aws.amazon.com/lambda/latest/api/API_GetFunction.html) " +
			"managed externally in AWS.",
		MarkdownExamples: []string{
			string(bundledExample),
			string(exportAllExample),
		},
		Fields: lambdaFunctionDataSourceSchema(),
		FilterFields: map[string]*provider.DataSourceFilterSchema{
			"arn": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The ARN of the Lambda function to retrieve.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The name of the Lambda function to retrieve.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
				ConflictsWith: []string{"arn"},
			},
			"qualifier": {
				Type: provider.DataSourceFilterSearchValueTypeString,
				Description: "The qualifier of the Lambda function to retrieve (e.g. $LATEST, an-alias, 2). " +
					"When not included, the latest published version will be retrieved, " +
					"if there isn't a published version, then the latest unpublished version will be retrieved.",
				FormattedDescription: "The qualifier of the Lambda function to retrieve (e.g. `$LATEST`, `an-alias`, `2`). " +
					"When not included, the latest published version will be retrieved, " +
					"if there isn't a published version, then the latest unpublished version will be retrieved.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
			"region": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The region of the Lambda function to retrieve. Defaults to the region of the provider.",
				FormattedDescription: "The [region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) " +
					"of the Lambda function to retrieve. Defaults to the region of the provider.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
		},
		FetchFunc: lambdaFunctionFetcher.Fetch,
	}
}

type lambdaFunctionDataSourceFetcher struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaFunctionDataSourceFetcher) getLambdaService(
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

func (l *lambdaFunctionDataSourceFetcher) Fetch(
	ctx context.Context,
	input *provider.DataSourceFetchInput,
) (*provider.DataSourceFetchOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input)
	if err != nil {
		return nil, err
	}

	functionNameOrARN := extractFunctionNameOrARNFromFilters(
		input.DataSourceWithResolvedSubs.Filter,
	)
	if functionNameOrARN == nil {
		return nil, fmt.Errorf(
			"function name or ARN is required for the lambda function data source",
		)
	}

	qualifier := extractQualifierFromFilters(input.DataSourceWithResolvedSubs.Filter)

	functionOutput, err := lambdaService.GetFunction(
		ctx,
		&lambda.GetFunctionInput{
			FunctionName: aws.String(core.StringValue(functionNameOrARN)),
			Qualifier:    aws.String(core.StringValue(qualifier)),
		},
	)
	if err != nil {
		return nil, err
	}

	data := l.createBaseData(
		functionOutput,
	)

	err = l.addOptionalConfigurationsToData(
		functionOutput,
		data.Fields,
	)
	if err != nil {
		return nil, err
	}

	err = l.addAdditionalConfigurationsToData(
		ctx,
		input.DataSourceWithResolvedSubs.Filter,
		data.Fields,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	return &provider.DataSourceFetchOutput{
		Data: data.Fields,
	}, nil
}

func (l *lambdaFunctionDataSourceFetcher) createBaseData(
	functionOutput *lambda.GetFunctionOutput,
) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"architecture": core.MappingNodeFromString(
				string(functionOutput.Configuration.Architectures[0]),
			),
			"arn": core.MappingNodeFromString(
				aws.ToString(functionOutput.Configuration.FunctionArn),
			),
			"codeSHA256": core.MappingNodeFromString(
				aws.ToString(functionOutput.Configuration.CodeSha256),
			),
			"name": core.MappingNodeFromString(
				aws.ToString(functionOutput.Configuration.FunctionName),
			),
			"qualifiedArn": core.MappingNodeFromString(
				aws.ToString(functionOutput.Configuration.FunctionArn) + ":" +
					aws.ToString(functionOutput.Configuration.Version),
			),
			"sourceCodeSize": core.MappingNodeFromInt(
				int(functionOutput.Configuration.CodeSize),
			),
			"version": core.MappingNodeFromString(
				aws.ToString(functionOutput.Configuration.Version),
			),
		},
	}
}

func (l *lambdaFunctionDataSourceFetcher) addOptionalConfigurationsToData(
	functionOutput *lambda.GetFunctionOutput,
	targetData map[string]*core.MappingNode,
) error {
	configurations := []pluginutils.OptionalValueExtractor[*lambda.GetFunctionOutput]{
		{
			Name: "deadLetterConfig",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.DeadLetterConfig != nil
			},
			Fields: []string{"deadLetterConfig.targetArn"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(
						aws.ToString(output.Configuration.DeadLetterConfig.TargetArn),
					),
				}, nil
			},
		},
		{
			Name: "environment",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.Environment != nil
			},
			Fields: []string{"environment.variables"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				envVars, err := serialiseFunctionEnvVars(output.Configuration)
				if err != nil {
					return nil, err
				}
				return []*core.MappingNode{envVars}, nil
			},
		},
		{
			Name: "ephemeralStorage",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.EphemeralStorage != nil
			},
			Fields: []string{"ephemeralStorage.size"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromInt(int(aws.ToInt32(output.Configuration.EphemeralStorage.Size))),
				}, nil
			},
		},
		{
			Name: "fileSystemConfig",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return len(output.Configuration.FileSystemConfigs) > 0
			},
			Fields: []string{"fileSystemConfig.arn", "fileSystemConfig.localMountPath"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(
						aws.ToString(output.Configuration.FileSystemConfigs[0].Arn),
					),
					core.MappingNodeFromString(
						aws.ToString(output.Configuration.FileSystemConfigs[0].LocalMountPath),
					),
				}, nil
			},
		},
		functionHandlerValueExtractor(),
		{
			Name: "imageUri",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Code != nil && output.Code.ImageUri != nil
			},
			Fields: []string{"imageUri"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(aws.ToString(output.Code.ImageUri)),
				}, nil
			},
		},
		functionKMSKeyArnValueExtractor(),
		{
			Name: "layers",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return len(output.Configuration.Layers) > 0
			},
			Fields: []string{"layers"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				layerArns := make([]*core.MappingNode, len(output.Configuration.Layers))
				for i, layer := range output.Configuration.Layers {
					layerArns[i] = core.MappingNodeFromString(aws.ToString(layer.Arn))
				}
				return []*core.MappingNode{
					{Items: layerArns},
				}, nil
			},
		},
		{
			Name: "loggingConfig",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.LoggingConfig != nil
			},
			Fields: []string{
				"loggingConfig.applicationLogLevel",
				"loggingConfig.logFormat",
				"loggingConfig.logGroup",
				"loggingConfig.systemLogLevel",
			},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				loggingConfig := output.Configuration.LoggingConfig
				return []*core.MappingNode{
					core.MappingNodeFromString(string(loggingConfig.ApplicationLogLevel)),
					core.MappingNodeFromString(string(loggingConfig.LogFormat)),
					core.MappingNodeFromString(aws.ToString(loggingConfig.LogGroup)),
					core.MappingNodeFromString(string(loggingConfig.SystemLogLevel)),
				}, nil
			},
		},
		{
			Name: "memorySize",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.MemorySize != nil
			},
			Fields: []string{"memorySize"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromInt(int(aws.ToInt32(output.Configuration.MemorySize))),
				}, nil
			},
		},
		{
			Name: "role",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.Role != nil
			},
			Fields: []string{"role"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(aws.ToString(output.Configuration.Role)),
				}, nil
			},
		},
		{
			Name: "runtime",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.Runtime != ""
			},
			Fields: []string{"runtime"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(string(output.Configuration.Runtime)),
				}, nil
			},
		},
		{
			Name: "signingJobArn",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.SigningJobArn != nil
			},
			Fields: []string{"signingJobArn"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(aws.ToString(output.Configuration.SigningJobArn)),
				}, nil
			},
		},
		{
			Name: "timeout",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.Timeout != nil
			},
			Fields: []string{"timeout"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromInt(int(aws.ToInt32(output.Configuration.Timeout))),
				}, nil
			},
		},
		{
			Name: "tracingConfig",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.TracingConfig != nil
			},
			Fields: []string{"tracingConfig.mode"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(string(output.Configuration.TracingConfig.Mode)),
				}, nil
			},
		},
		{
			Name: "vpcConfig",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.VpcConfig != nil
			},
			Fields: []string{
				"vpcConfig.ipv6AllowedForDualStack",
				"vpcConfig.securityGroupIds",
				"vpcConfig.subnetIds",
			},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				vpcConfig := output.Configuration.VpcConfig

				// Security Group IDs
				securityGroupIds := make([]*core.MappingNode, len(vpcConfig.SecurityGroupIds))
				for i, sgId := range vpcConfig.SecurityGroupIds {
					securityGroupIds[i] = core.MappingNodeFromString(sgId)
				}

				// Subnet IDs
				subnetIds := make([]*core.MappingNode, len(vpcConfig.SubnetIds))
				for i, subnetId := range vpcConfig.SubnetIds {
					subnetIds[i] = core.MappingNodeFromString(subnetId)
				}

				return []*core.MappingNode{
					core.MappingNodeFromBool(aws.ToBool(vpcConfig.Ipv6AllowedForDualStack)),
					{Items: securityGroupIds},
					{Items: subnetIds},
				}, nil
			},
		},
	}

	return pluginutils.RunOptionalValueExtractors(
		functionOutput,
		targetData,
		configurations,
	)
}

func (l *lambdaFunctionDataSourceFetcher) addAdditionalConfigurationsToData(
	ctx context.Context,
	filters *provider.ResolvedDataSourceFilters,
	targetData map[string]*core.MappingNode,
	lambdaService lambdaservice.Service,
) error {
	additionalConfigurations := []pluginutils.AdditionalValueExtractor[lambdaservice.Service]{
		{
			Name:    "code signing config",
			Extract: l.addCodeSigningConfigToData,
		},
		{
			Name:    "concurrency config",
			Extract: l.addConcurrencyConfigToData,
		},
	}

	return pluginutils.RunAdditionalValueExtractors(
		ctx,
		filters,
		targetData,
		additionalConfigurations,
		lambdaService,
	)
}

func (l *lambdaFunctionDataSourceFetcher) addCodeSigningConfigToData(
	ctx context.Context,
	filters *provider.ResolvedDataSourceFilters,
	targetData map[string]*core.MappingNode,
	lambdaService lambdaservice.Service,
) error {
	functionNameOrARN := extractFunctionNameOrARNFromFilters(filters)
	codeSigningConfigOutput, err := lambdaService.GetFunctionCodeSigningConfig(
		ctx,
		&lambda.GetFunctionCodeSigningConfigInput{
			FunctionName: aws.String(core.StringValue(functionNameOrARN)),
		},
	)
	if err != nil {
		var apiErr interface{ ErrorCode() string }
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "ResourceNotFoundException" {
			return nil
		}
		if strings.Contains(err.Error(), "ResourceNotFoundException") {
			return nil
		}
		return err
	}

	if codeSigningConfigOutput.CodeSigningConfigArn != nil {
		targetData["codeSigningConfigArn"] = core.MappingNodeFromString(
			aws.ToString(codeSigningConfigOutput.CodeSigningConfigArn),
		)
	}
	return nil
}

func (l *lambdaFunctionDataSourceFetcher) addConcurrencyConfigToData(
	ctx context.Context,
	filters *provider.ResolvedDataSourceFilters,
	targetData map[string]*core.MappingNode,
	lambdaService lambdaservice.Service,
) error {
	functionNameOrARN := extractFunctionNameOrARNFromFilters(filters)
	concurrencyConfigOutput, err := lambdaService.GetFunctionConcurrency(
		ctx,
		&lambda.GetFunctionConcurrencyInput{
			FunctionName: aws.String(core.StringValue(functionNameOrARN)),
		},
	)
	if err != nil {
		var apiErr interface{ ErrorCode() string }
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "ResourceNotFoundException" {
			return nil
		}
		if strings.Contains(err.Error(), "ResourceNotFoundException") {
			return nil
		}
		return err
	}

	if concurrencyConfigOutput.ReservedConcurrentExecutions != nil {
		targetData["reservedConcurrentExecutions"] = core.MappingNodeFromInt(
			int(aws.ToInt32(concurrencyConfigOutput.ReservedConcurrentExecutions)),
		)
	}
	return nil
}

func serialiseFunctionEnvVars(
	funcConfig *types.FunctionConfiguration,
) (*core.MappingNode, error) {
	if funcConfig.Environment == nil {
		return core.MappingNodeFromString("{}"), nil
	}

	envVars := funcConfig.Environment.Variables
	jsonBytes, err := json.Marshal(envVars)
	if err != nil {
		return nil, err
	}

	return core.MappingNodeFromString(string(jsonBytes)), nil
}
