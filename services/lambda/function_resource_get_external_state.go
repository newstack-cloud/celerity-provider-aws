package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Try to get ARN from current spec first
	functionARN := ""
	if input.CurrentResourceSpec != nil && input.CurrentResourceSpec.Fields != nil {
		functionARN = core.StringValue(input.CurrentResourceSpec.Fields["arn"])
	}

	// If no ARN, attempt fallback lookup by Bluelink tags
	if functionARN == "" {
		fallbackArn, err := l.lookupFunctionARNByTags(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup function by tags: %w", err)
		}
		if fallbackArn == "" {
			// Resource doesn't exist yet
			return &provider.ResourceGetExternalStateOutput{
				ResourceSpecState: &core.MappingNode{
					Fields: map[string]*core.MappingNode{},
				},
			}, nil
		}
		functionARN = fallbackArn
	}

	functionOutput, err := lambdaService.GetFunction(
		ctx,
		&lambda.GetFunctionInput{
			FunctionName: &functionARN,
		},
	)
	if err != nil {
		return nil, err
	}

	resourceSpecState := l.buildBaseResourceSpecState(
		functionOutput,
		input.CurrentResourceSpec.Fields["code"],
	)

	err = l.addOptionalConfigurationsToSpec(
		functionOutput,
		resourceSpecState.Fields,
		input.ProviderContext,
	)
	if err != nil {
		return nil, err
	}

	err = l.addAdditionalConfigurationsToSpec(
		ctx,
		functionARN,
		resourceSpecState.Fields,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	l.addComputedFieldsToSpec(functionOutput, resourceSpecState.Fields)

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}

func (l *lambdaFunctionResourceActions) buildBaseResourceSpecState(
	functionOutput *lambda.GetFunctionOutput,
	inputSpecCode *core.MappingNode,
) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(
				aws.ToString(functionOutput.Configuration.FunctionArn),
			),
			"architecture": core.MappingNodeFromString(
				string(functionOutput.Configuration.Architectures[0]),
			),
			"code": functionCodeConfigToMappingNode(
				functionOutput.Code,
				inputSpecCode,
			),
			"functionName": core.MappingNodeFromString(
				aws.ToString(functionOutput.Configuration.FunctionName),
			),
		},
	}
}

func (l *lambdaFunctionResourceActions) addOptionalConfigurationsToSpec(
	functionOutput *lambda.GetFunctionOutput,
	specFields map[string]*core.MappingNode,
	providerContext provider.Context,
) error {
	extractors := []pluginutils.OptionalValueExtractor[*lambda.GetFunctionOutput]{
		{
			Name: "deadLetterConfig",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.DeadLetterConfig != nil
			},
			Fields: []string{"deadLetterConfig"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					functionDeadLetterConfigToMappingNode(
						output.Configuration.DeadLetterConfig,
					)}, nil
			},
		},
		{
			Name: "description",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.Description != nil
			},
			Fields: []string{"description"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(
						*output.Configuration.Description,
					)}, nil
			},
		},
		{
			Name: "environment",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.Environment != nil
			},
			Fields: []string{"environment"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					functionEnvToMappingNode(output.Configuration.Environment),
				}, nil
			},
		},
		{
			Name: "ephemeralStorage",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.EphemeralStorage != nil
			},
			Fields: []string{"ephemeralStorage"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					functionEphemeralStorageToMappingNode(output.Configuration.EphemeralStorage),
				}, nil
			},
		},
		{
			Name: "fileSystemConfig",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.FileSystemConfigs != nil
			},
			Fields: []string{"fileSystemConfig"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					functionFileSystemConfigsToMappingNode(output.Configuration.FileSystemConfigs),
				}, nil
			},
		},
		functionHandlerValueExtractor(),
		{
			Name: "imageConfig",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.ImageConfigResponse != nil
			},
			Fields: []string{"imageConfig"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					functionImageConfigToMappingNode(output.Configuration.ImageConfigResponse),
				}, nil
			},
		},
		functionKMSKeyArnValueExtractor(),
		{
			Name: "layers",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.Layers != nil
			},
			Fields: []string{"layers"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					functionLayersToMappingNode(output.Configuration.Layers),
				}, nil
			},
		},
		{
			Name: "loggingConfig",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.LoggingConfig != nil
			},
			Fields: []string{"loggingConfig"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					functionLoggingConfigToMappingNode(output.Configuration.LoggingConfig),
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
			Name: "packageType",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.PackageType != ""
			},
			Fields: []string{"packageType"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(string(output.Configuration.PackageType)),
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
			Name: "runtimeManagementConfig",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.RuntimeVersionConfig != nil
			},
			Fields: []string{"runtimeManagementConfig"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					functionRuntimeVersionConfigToMappingNode(
						output.Configuration.RuntimeVersionConfig,
						specFields["runtimeManagementConfig"],
					),
				}, nil
			},
		},
		{
			Name: "snapStart",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.SnapStart != nil
			},
			Fields: []string{"snapStart"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					functionSnapStartConfigToMappingNode(output.Configuration.SnapStart),
				}, nil
			},
		},
		// Note: tags are handled separately below to filter Bluelink provenance tags
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
			Fields: []string{"tracingConfig"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					functionTracingConfigToMappingNode(output.Configuration.TracingConfig),
				}, nil
			},
		},
		{
			Name: "vpcConfig",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.VpcConfig != nil
			},
			Fields: []string{"vpcConfig"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					functionVPCConfigToMappingNode(output.Configuration.VpcConfig),
				}, nil
			},
		},
	}

	err := pluginutils.RunOptionalValueExtractors(
		functionOutput,
		specFields,
		extractors,
	)
	if err != nil {
		return err
	}

	// Handle tags separately to filter out Bluelink provenance tags
	if len(functionOutput.Tags) > 0 {
		specFields["tags"] = utils.UserTagsToMappingNode(functionOutput.Tags, providerContext)
	}

	return nil
}

func (l *lambdaFunctionResourceActions) addAdditionalConfigurationsToSpec(
	ctx context.Context,
	functionARN string,
	specFields map[string]*core.MappingNode,
	lambdaService lambdaservice.Service,
) error {
	extractors := []pluginutils.AdditionalValueExtractor[lambdaservice.Service]{
		{
			Name: "code signing config",
			Extract: func(
				ctx context.Context,
				filters *provider.ResolvedDataSourceFilters,
				specFields map[string]*core.MappingNode,
				lambdaService lambdaservice.Service,
			) error {
				return l.addCodeSigningConfigToSpec(ctx, filters, specFields, lambdaService)
			},
		},
		{
			Name: "recursion config",
			Extract: func(
				ctx context.Context,
				filters *provider.ResolvedDataSourceFilters,
				specFields map[string]*core.MappingNode,
				lambdaService lambdaservice.Service,
			) error {
				return l.addRecursionConfigToSpec(ctx, filters, specFields, lambdaService)
			},
		},
		{
			Name: "concurrency config",
			Extract: func(
				ctx context.Context,
				filters *provider.ResolvedDataSourceFilters,
				specFields map[string]*core.MappingNode,
				lambdaService lambdaservice.Service,
			) error {
				return l.addConcurrencyConfigToSpec(ctx, filters, specFields, lambdaService)
			},
		},
	}

	filters := pluginutils.CreateStringEqualsFilter("arn", functionARN)

	err := pluginutils.RunAdditionalValueExtractors(
		ctx,
		filters,
		specFields,
		extractors,
		lambdaService,
	)
	if err != nil {
		return err
	}

	return nil
}

func (l *lambdaFunctionResourceActions) addComputedFieldsToSpec(
	functionOutput *lambda.GetFunctionOutput,
	specFields map[string]*core.MappingNode,
) {
	specFields["arn"] = core.MappingNodeFromString(
		aws.ToString(functionOutput.Configuration.FunctionArn),
	)

	if functionOutput.Configuration.SnapStart != nil {
		specFields["snapStartResponseApplyOn"] = core.MappingNodeFromString(
			string(functionOutput.Configuration.SnapStart.ApplyOn),
		)
		specFields["snapStartResponseOptimizationStatus"] = core.MappingNodeFromString(
			string(functionOutput.Configuration.SnapStart.OptimizationStatus),
		)
	}
}

func (l *lambdaFunctionResourceActions) addCodeSigningConfigToSpec(
	ctx context.Context,
	filters *provider.ResolvedDataSourceFilters,
	specFields map[string]*core.MappingNode,
	lambdaService lambdaservice.Service,
) error {
	nameOrARN := extractFunctionNameOrARNFromFilters(filters)
	codeSigningConfigOutput, err := lambdaService.GetFunctionCodeSigningConfig(
		ctx,
		&lambda.GetFunctionCodeSigningConfigInput{
			FunctionName: aws.String(core.StringValue(nameOrARN)),
		},
	)
	if err != nil {
		return err
	}

	if codeSigningConfigOutput.CodeSigningConfigArn != nil {
		specFields["codeSigningConfigArn"] = core.MappingNodeFromString(
			*codeSigningConfigOutput.CodeSigningConfigArn,
		)
	}

	return nil
}

func (l *lambdaFunctionResourceActions) addRecursionConfigToSpec(
	ctx context.Context,
	filters *provider.ResolvedDataSourceFilters,
	specFields map[string]*core.MappingNode,
	lambdaService lambdaservice.Service,
) error {
	nameOrARN := extractFunctionNameOrARNFromFilters(filters)
	recursionConfigOutput, err := lambdaService.GetFunctionRecursionConfig(
		ctx,
		&lambda.GetFunctionRecursionConfigInput{
			FunctionName: aws.String(core.StringValue(nameOrARN)),
		},
	)
	if err != nil {
		return err
	}

	if recursionConfigOutput.RecursiveLoop != "" {
		specFields["recursiveLoop"] = core.MappingNodeFromString(
			string(recursionConfigOutput.RecursiveLoop),
		)
	}

	return nil
}

func (l *lambdaFunctionResourceActions) addConcurrencyConfigToSpec(
	ctx context.Context,
	filters *provider.ResolvedDataSourceFilters,
	specFields map[string]*core.MappingNode,
	lambdaService lambdaservice.Service,
) error {
	nameOrARN := extractFunctionNameOrARNFromFilters(filters)
	concurrencyConfigOutput, err := lambdaService.GetFunctionConcurrency(
		ctx,
		&lambda.GetFunctionConcurrencyInput{
			FunctionName: aws.String(core.StringValue(nameOrARN)),
		},
	)
	if err != nil {
		return err
	}

	if concurrencyConfigOutput.ReservedConcurrentExecutions != nil {
		specFields["reservedConcurrentExecutions"] = core.MappingNodeFromInt(
			int(aws.ToInt32(concurrencyConfigOutput.ReservedConcurrentExecutions)),
		)
	}

	return nil
}

func functionCodeConfigToMappingNode(
	code *types.FunctionCodeLocation,
	inputSpecCode *core.MappingNode,
) *core.MappingNode {
	fields := map[string]*core.MappingNode{}

	// For code source fields for a `Zip` package type, the source config is
	// not available in the FunctionCodeLocation
	// in the response when fetching the function, a pre-signed URL is returned instead.
	// When retrieving external state for resources, if fields in the spec are not available
	// in the upstream provider response, they will be set to the value in the input spec.
	if inputSpecCode != nil {
		if s3Bucket, hasBucket := inputSpecCode.Fields["s3Bucket"]; hasBucket {
			fields["s3Bucket"] = s3Bucket
		}
		if s3Key, hasKey := inputSpecCode.Fields["s3Key"]; hasKey {
			fields["s3Key"] = s3Key
		}
		if s3ObjectVersion, hasVersion := inputSpecCode.Fields["s3ObjectVersion"]; hasVersion {
			fields["s3ObjectVersion"] = s3ObjectVersion
		}
		if zipFile, hasZipFile := inputSpecCode.Fields["zipFile"]; hasZipFile {
			fields["zipFile"] = zipFile
		}
	}

	if code.ImageUri != nil {
		fields["imageUri"] = core.MappingNodeFromString(aws.ToString(code.ImageUri))
	}

	if code.SourceKMSKeyArn != nil {
		fields["sourceKMSKeyArn"] = core.MappingNodeFromString(aws.ToString(code.SourceKMSKeyArn))
	}

	return &core.MappingNode{Fields: fields}
}

func functionDeadLetterConfigToMappingNode(
	deadLetterConfig *types.DeadLetterConfig,
) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"targetArn": core.MappingNodeFromString(
				aws.ToString(deadLetterConfig.TargetArn),
			),
		},
	}
}

func functionEnvToMappingNode(
	environment *types.EnvironmentResponse,
) *core.MappingNode {
	if environment.Variables == nil {
		return &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		}
	}

	variables := make(map[string]*core.MappingNode, len(environment.Variables))
	for key, value := range environment.Variables {
		variables[key] = core.MappingNodeFromString(value)
	}

	return &core.MappingNode{
		Fields: variables,
	}
}

func functionEphemeralStorageToMappingNode(
	ephemeralStorage *types.EphemeralStorage,
) *core.MappingNode {
	if ephemeralStorage.Size == nil {
		return &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		}
	}

	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"size": core.MappingNodeFromInt(int(
				aws.ToInt32(ephemeralStorage.Size),
			)),
		},
	}
}

func functionFileSystemConfigsToMappingNode(
	fileSystemConfigs []types.FileSystemConfig,
) *core.MappingNode {
	if len(fileSystemConfigs) == 0 {
		return &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		}
	}
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(
				aws.ToString(fileSystemConfigs[0].Arn),
			),
			"localMountPath": core.MappingNodeFromString(
				aws.ToString(fileSystemConfigs[0].LocalMountPath),
			),
		},
	}
}

func functionImageConfigToMappingNode(
	imageConfigResponse *types.ImageConfigResponse,
) *core.MappingNode {
	if imageConfigResponse.ImageConfig == nil {
		return &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		}
	}

	fields := map[string]*core.MappingNode{}

	if imageConfigResponse.ImageConfig.Command != nil {
		fields["command"] = core.MappingNodeFromStringSlice(
			imageConfigResponse.ImageConfig.Command,
		)
	}

	if imageConfigResponse.ImageConfig.EntryPoint != nil {
		fields["entryPoint"] = core.MappingNodeFromStringSlice(
			imageConfigResponse.ImageConfig.EntryPoint,
		)
	}

	if imageConfigResponse.ImageConfig.WorkingDirectory != nil {
		fields["workingDirectory"] = core.MappingNodeFromString(
			aws.ToString(imageConfigResponse.ImageConfig.WorkingDirectory),
		)
	}

	return &core.MappingNode{
		Fields: fields,
	}
}

func functionLayersToMappingNode(
	layers []types.Layer,
) *core.MappingNode {
	if len(layers) == 0 {
		return &core.MappingNode{
			Items: []*core.MappingNode{},
		}
	}

	items := make([]*core.MappingNode, len(layers))
	for i, layer := range layers {
		items[i] = core.MappingNodeFromString(aws.ToString(layer.Arn))
	}

	return &core.MappingNode{
		Items: items,
	}
}

func functionLoggingConfigToMappingNode(
	loggingConfig *types.LoggingConfig,
) *core.MappingNode {
	if loggingConfig == nil {
		return &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		}
	}

	fields := map[string]*core.MappingNode{}

	if loggingConfig.ApplicationLogLevel != "" {
		fields["applicationLogLevel"] = core.MappingNodeFromString(
			string(loggingConfig.ApplicationLogLevel),
		)
	}

	if loggingConfig.LogFormat != "" {
		fields["logFormat"] = core.MappingNodeFromString(
			string(loggingConfig.LogFormat),
		)
	}

	if loggingConfig.LogGroup != nil {
		fields["logGroup"] = core.MappingNodeFromString(
			aws.ToString(loggingConfig.LogGroup),
		)
	}

	if loggingConfig.SystemLogLevel != "" {
		fields["systemLogLevel"] = core.MappingNodeFromString(
			string(loggingConfig.SystemLogLevel),
		)
	}

	return &core.MappingNode{
		Fields: fields,
	}
}

func functionRuntimeVersionConfigToMappingNode(
	runtimeVersionConfig *types.RuntimeVersionConfig,
	inputSpecRuntimeVersionConfig *core.MappingNode,
) *core.MappingNode {
	fields := map[string]*core.MappingNode{}

	if runtimeVersionConfig.RuntimeVersionArn != nil {
		fields["runtimeVersionArn"] = core.MappingNodeFromString(
			aws.ToString(runtimeVersionConfig.RuntimeVersionArn),
		)
	}

	// The `updateRuntimeOn` field is an input when saving a lambda function but is not persisted
	// as part of the resource state in AWS, so like other fields that are input-only,
	// it is sourced from the input spec.
	if inputSpecRuntimeVersionConfig != nil {
		if updateRuntimeOn, ok := inputSpecRuntimeVersionConfig.Fields["updateRuntimeOn"]; ok {
			fields["updateRuntimeOn"] = updateRuntimeOn
		}
	}

	return &core.MappingNode{
		Fields: fields,
	}
}

func functionSnapStartConfigToMappingNode(
	snapStartConfig *types.SnapStartResponse,
) *core.MappingNode {
	fields := map[string]*core.MappingNode{}

	if snapStartConfig.ApplyOn != "" {
		fields["applyOn"] = core.MappingNodeFromString(
			string(snapStartConfig.ApplyOn),
		)
	}

	return &core.MappingNode{
		Fields: fields,
	}
}

func functionTracingConfigToMappingNode(
	tracingConfig *types.TracingConfigResponse,
) *core.MappingNode {
	fields := map[string]*core.MappingNode{}

	if tracingConfig.Mode != "" {
		fields["mode"] = core.MappingNodeFromString(
			string(tracingConfig.Mode),
		)
	}

	return &core.MappingNode{
		Fields: fields,
	}
}

func functionVPCConfigToMappingNode(
	vpcConfig *types.VpcConfigResponse,
) *core.MappingNode {
	fields := map[string]*core.MappingNode{}

	if vpcConfig.SecurityGroupIds != nil {
		fields["securityGroupIds"] = core.MappingNodeFromStringSlice(
			vpcConfig.SecurityGroupIds,
		)
	}

	if vpcConfig.SubnetIds != nil {
		fields["subnetIds"] = core.MappingNodeFromStringSlice(
			vpcConfig.SubnetIds,
		)
	}

	if vpcConfig.Ipv6AllowedForDualStack != nil {
		fields["ipv6AllowedForDualStack"] = core.MappingNodeFromBool(
			aws.ToBool(vpcConfig.Ipv6AllowedForDualStack),
		)
	}

	return &core.MappingNode{
		Fields: fields,
	}
}

// lookupFunctionARNByTags attempts to find a Lambda function by its Bluelink provenance tags.
// This is used as a fallback when the ARN is not available (e.g., interrupted resource creation).
// Returns empty string if no matching resource is found.
func (l *lambdaFunctionResourceActions) lookupFunctionARNByTags(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (string, error) {
	tagFilters := utils.BuildBluelinkTagFiltersForLookup(input)
	if tagFilters == nil {
		// Tagging is not enabled, cannot perform fallback lookup
		return "", nil
	}

	taggingService, err := l.getResourceGroupTaggingService(ctx, input.ProviderContext)
	if err != nil {
		return "", err
	}

	result, err := taggingService.GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters:          tagFilters,
		ResourceTypeFilters: []string{"lambda:function"},
	})
	if err != nil {
		return "", err
	}

	if len(result.ResourceTagMappingList) == 0 {
		return "", nil
	}

	return aws.ToString(result.ResourceTagMappingList[0].ResourceARN), nil
}
