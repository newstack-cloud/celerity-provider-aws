package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/two-hundred/celerity-provider-aws/utils"
	"github.com/two-hundred/celerity/libs/blueprint/core"
	"github.com/two-hundred/celerity/libs/blueprint/provider"
)

func (l *lambdaFunctionResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	functionARN := core.StringValue(
		input.CurrentResourceSpec.Fields["arn"],
	)

	functionOutput, err := lambdaService.GetFunction(
		ctx,
		&lambda.GetFunctionInput{
			FunctionName: &functionARN,
		},
	)
	if err != nil {
		return nil, err
	}

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"architecture": core.MappingNodeFromString(
				string(functionOutput.Configuration.Architectures[0]),
			),
			"code": functionCodeConfigToMappingNode(
				functionOutput.Code,
				input.CurrentResourceSpec.Fields["code"],
			),
			"functionName": core.MappingNodeFromString(
				aws.ToString(functionOutput.Configuration.FunctionName),
			),
		},
	}

	if functionOutput.Configuration.DeadLetterConfig != nil {
		resourceSpecState.Fields["deadLetterConfig"] = functionDeadLetterConfigToMappingNode(
			functionOutput.Configuration.DeadLetterConfig,
		)
	}

	if functionOutput.Configuration.Description != nil {
		resourceSpecState.Fields["description"] = core.MappingNodeFromString(
			*functionOutput.Configuration.Description,
		)
	}

	if functionOutput.Configuration.Environment != nil {
		resourceSpecState.Fields["environment"] = functionEnvToMappingNode(
			functionOutput.Configuration.Environment,
		)
	}

	if functionOutput.Configuration.EphemeralStorage != nil {
		resourceSpecState.Fields["ephemeralStorage"] = functionEphemeralStorageToMappingNode(
			functionOutput.Configuration.EphemeralStorage,
		)
	}

	if functionOutput.Configuration.FileSystemConfigs != nil {
		resourceSpecState.Fields["fileSystemConfig"] = functionFileSystemConfigsToMappingNode(
			functionOutput.Configuration.FileSystemConfigs,
		)
	}

	if functionOutput.Configuration.Handler != nil {
		resourceSpecState.Fields["handler"] = core.MappingNodeFromString(
			aws.ToString(functionOutput.Configuration.Handler),
		)
	}

	if functionOutput.Configuration.ImageConfigResponse != nil {
		resourceSpecState.Fields["imageConfig"] = functionImageConfigToMappingNode(
			functionOutput.Configuration.ImageConfigResponse,
		)
	}

	if functionOutput.Configuration.KMSKeyArn != nil {
		resourceSpecState.Fields["kmsKeyArn"] = core.MappingNodeFromString(
			aws.ToString(functionOutput.Configuration.KMSKeyArn),
		)
	}

	if functionOutput.Configuration.Layers != nil {
		resourceSpecState.Fields["layers"] = functionLayersToMappingNode(
			functionOutput.Configuration.Layers,
		)
	}

	if functionOutput.Configuration.LoggingConfig != nil {
		resourceSpecState.Fields["loggingConfig"] = functionLoggingConfigToMappingNode(
			functionOutput.Configuration.LoggingConfig,
		)
	}

	if functionOutput.Configuration.MemorySize != nil {
		resourceSpecState.Fields["memorySize"] = core.MappingNodeFromInt(
			int(aws.ToInt32(functionOutput.Configuration.MemorySize)),
		)
	}

	if functionOutput.Configuration.PackageType != "" {
		resourceSpecState.Fields["packageType"] = core.MappingNodeFromString(
			string(functionOutput.Configuration.PackageType),
		)
	}

	if functionOutput.Configuration.Role != nil {
		resourceSpecState.Fields["role"] = core.MappingNodeFromString(
			aws.ToString(functionOutput.Configuration.Role),
		)
	}

	if functionOutput.Configuration.Runtime != "" {
		resourceSpecState.Fields["runtime"] = core.MappingNodeFromString(
			string(functionOutput.Configuration.Runtime),
		)
	}

	if functionOutput.Configuration.RuntimeVersionConfig != nil {
		resourceSpecState.Fields["runtimeManagementConfig"] = functionRuntimeVersionConfigToMappingNode(
			functionOutput.Configuration.RuntimeVersionConfig,
			input.CurrentResourceSpec.Fields["runtimeManagementConfig"],
		)
	}

	if functionOutput.Configuration.SnapStart != nil {
		resourceSpecState.Fields["snapStart"] = functionSnapStartConfigToMappingNode(
			functionOutput.Configuration.SnapStart,
		)
	}

	if len(functionOutput.Tags) > 0 {
		resourceSpecState.Fields["tags"] = utils.TagsToMappingNode(
			functionOutput.Tags,
		)
	}

	if functionOutput.Configuration.Timeout != nil {
		resourceSpecState.Fields["timeout"] = core.MappingNodeFromInt(
			int(aws.ToInt32(functionOutput.Configuration.Timeout)),
		)
	}

	if functionOutput.Configuration.TracingConfig != nil {
		resourceSpecState.Fields["tracingConfig"] = functionTracingConfigToMappingNode(
			functionOutput.Configuration.TracingConfig,
		)
	}

	if functionOutput.Configuration.VpcConfig != nil {
		resourceSpecState.Fields["vpcConfig"] = functionVPCConfigToMappingNode(
			functionOutput.Configuration.VpcConfig,
		)
	}

	err = l.addCodeSigningConfigToSpec(
		ctx,
		functionARN,
		resourceSpecState.Fields,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	err = l.addRecursionConfigToSpec(
		ctx,
		functionARN,
		resourceSpecState.Fields,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	err = l.addConcurrencyConfigToSpec(
		ctx,
		functionARN,
		resourceSpecState.Fields,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	l.addComputedFieldsToSpec(
		functionOutput,
		resourceSpecState.Fields,
	)

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}

func (l *lambdaFunctionResourceActions) addComputedFieldsToSpec(
	functionOutput *lambda.GetFunctionOutput,
	specFields map[string]*core.MappingNode,
) {
	specFields["arn"] = core.MappingNodeFromString(
		aws.ToString(functionOutput.Configuration.FunctionArn),
	)

	if functionOutput.Configuration.SnapStart != nil {
		specFields["snapsnapStartResponseApplyOn"] = core.MappingNodeFromString(
			string(functionOutput.Configuration.SnapStart.ApplyOn),
		)
		specFields["snapStartResponseOptimizationStatus"] = core.MappingNodeFromString(
			string(functionOutput.Configuration.SnapStart.OptimizationStatus),
		)
	}
}

func (l *lambdaFunctionResourceActions) addCodeSigningConfigToSpec(
	ctx context.Context,
	functionARN string,
	specFields map[string]*core.MappingNode,
	lambdaService Service,
) error {
	codeSigningConfigOutput, err := lambdaService.GetFunctionCodeSigningConfig(
		ctx,
		&lambda.GetFunctionCodeSigningConfigInput{
			FunctionName: &functionARN,
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
	functionARN string,
	specFields map[string]*core.MappingNode,
	lambdaService Service,
) error {
	recursionConfigOutput, err := lambdaService.GetFunctionRecursionConfig(
		ctx,
		&lambda.GetFunctionRecursionConfigInput{
			FunctionName: &functionARN,
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
	functionARN string,
	specFields map[string]*core.MappingNode,
	lambdaService Service,
) error {
	concurrencyConfigOutput, err := lambdaService.GetFunctionConcurrency(
		ctx,
		&lambda.GetFunctionConcurrencyInput{
			FunctionName: &functionARN,
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
	if code == nil {
		return nil
	}

	fields := map[string]*core.MappingNode{}

	// For code source fields for a `Zip` package type, the source config is
	// not available in the FunctionCodeLocation
	// in the response when fetching the function, a pre-signed URL is returned instead.
	// When retrieving external state for resources, if fields in the spec are not available
	// in the upstream provider response, they will be set to the value in the input spec.
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
	if updateRuntimeOn, ok := inputSpecRuntimeVersionConfig.Fields["updateRuntimeOn"]; ok {
		fields["updateRuntimeOn"] = updateRuntimeOn
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
