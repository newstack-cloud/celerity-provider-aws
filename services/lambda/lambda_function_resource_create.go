package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[Service]{
		&functionCreate{},
		&functionConcurrencyUpdate{},
		&functionRecursionConfigUpdate{},
		&functionRuntimeManagementConfigUpdate{},
	}

	hasUpdates, saveOpCtx, err := pluginutils.RunSaveOperations(
		ctx,
		pluginutils.SaveOperationContext{
			Data: map[string]any{},
		},
		createOperations,
		input,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	if !hasUpdates {
		return nil, fmt.Errorf("no updates were made during function creation")
	}

	createFunctionOutput, ok := saveOpCtx.Data["createFunctionOutput"].(*lambda.CreateFunctionOutput)
	if !ok {
		return nil, fmt.Errorf("createFunctionOutput not found in save operation context")
	}

	computedFields := map[string]*core.MappingNode{
		"spec.arn": core.MappingNodeFromString(aws.ToString(createFunctionOutput.FunctionArn)),
	}

	if createFunctionOutput.SnapStart != nil {
		computedFields["spec.snapStartResponseApplyOn"] = core.MappingNodeFromString(
			string(createFunctionOutput.SnapStart.ApplyOn),
		)
		computedFields["spec.snapStartResponseOptimizationStatus"] = core.MappingNodeFromString(
			string(createFunctionOutput.SnapStart.OptimizationStatus),
		)
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}

func changesToCreateFunctionInput(
	specData *core.MappingNode,
) (*lambda.CreateFunctionInput, bool, error) {
	input := &lambda.CreateFunctionInput{}

	valueSetters := []*pluginutils.ValueSetter[*lambda.CreateFunctionInput]{
		pluginutils.NewValueSetter(
			"$.architecture",
			setCreateFunctionArchitecture,
		),
		pluginutils.NewValueSetter(
			"$.code",
			setCreateFunctionCode,
		),
		pluginutils.NewValueSetter(
			"$.codeSigningConfigArn",
			setCreateFunctionCodeSigningConfigARN,
		),
		pluginutils.NewValueSetter(
			"$.deadLetterConfig.targetArn",
			setCreateFunctionDeadLetterConfigTargetARN,
		),
		pluginutils.NewValueSetter(
			"$.description",
			setCreateFunctionDescription,
		),
		pluginutils.NewValueSetter(
			"$.environment.variables",
			setCreateFunctionEnvironmentVariables,
		),
		pluginutils.NewValueSetter(
			"$.ephemeralStorage.size",
			setCreateFunctionEphemeralStorageSize,
		),
		pluginutils.NewValueSetter(
			"$.fileSystemConfig",
			setCreateFunctionFileSystemConfig,
		),
		pluginutils.NewValueSetter(
			"$.functionName",
			setCreateFunctionName,
		),
		pluginutils.NewValueSetter(
			"$.handler",
			setCreateFunctionHandler,
		),
		pluginutils.NewValueSetter(
			"$.imageConfig",
			setCreateFunctionImageConfig,
		),
		pluginutils.NewValueSetter(
			"$.kmsKeyArn",
			setCreateFunctionKMSKeyARN,
		),
		pluginutils.NewValueSetter(
			"$.layers",
			setCreateFunctionLayers,
		),
		pluginutils.NewValueSetter(
			"$.loggingConfig",
			setCreateFunctionLoggingConfig,
		),
		pluginutils.NewValueSetter(
			"$.memorySize",
			setCreateFunctionMemorySize,
		),
		pluginutils.NewValueSetter(
			"$.packageType",
			setCreateFunctionPackageType,
		),
		pluginutils.NewValueSetter(
			"$.role",
			setCreateFunctionRole,
		),
		pluginutils.NewValueSetter(
			"$.runtime",
			setCreateFunctionRuntime,
		),
		pluginutils.NewValueSetter(
			"$.snapStart.applyOn",
			setCreateFunctionSnapStart,
		),
		pluginutils.NewValueSetter(
			"$.tags",
			setCreateFunctionTags,
		),
		pluginutils.NewValueSetter(
			"$.timeout",
			setCreateFunctionTimeout,
		),
		pluginutils.NewValueSetter(
			"$.tracingConfig.mode",
			setCreateFunctionTracingConfig,
		),
		pluginutils.NewValueSetter(
			"$.vpcConfig",
			setCreateFunctionVPCConfig,
		),
	}

	hasUpdates := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
		hasUpdates = hasUpdates || valueSetter.DidSet()
	}

	return input, hasUpdates, nil
}

func setCreateFunctionArchitecture(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.Architectures = []types.Architecture{
		types.Architecture(core.StringValue(value)),
	}
}

func setCreateFunctionCode(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	code := &types.FunctionCode{}

	valueSetters := []*pluginutils.ValueSetter[*types.FunctionCode]{
		pluginutils.NewValueSetter(
			"$.imageUri",
			func(value *core.MappingNode, target *types.FunctionCode) {
				target.ImageUri = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.s3Bucket",
			func(value *core.MappingNode, target *types.FunctionCode) {
				target.S3Bucket = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.s3Key",
			func(value *core.MappingNode, target *types.FunctionCode) {
				target.S3Key = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.s3ObjectVersion",
			func(value *core.MappingNode, target *types.FunctionCode) {
				target.S3ObjectVersion = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.sourceKMSKeyArn",
			func(value *core.MappingNode, target *types.FunctionCode) {
				target.SourceKMSKeyArn = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.zipFile",
			func(value *core.MappingNode, target *types.FunctionCode) {
				target.ZipFile = []byte(core.StringValue(value))
			},
		),
	}

	for _, valueSetter := range valueSetters {
		valueSetter.Set(value, code)
	}

	input.Code = code
}

func setCreateFunctionDeadLetterConfigTargetARN(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.DeadLetterConfig = &types.DeadLetterConfig{
		TargetArn: aws.String(core.StringValue(value)),
	}
}

func setCreateFunctionDescription(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.Description = aws.String(core.StringValue(value))
}

func setCreateFunctionEnvironmentVariables(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.Environment = &types.Environment{
		Variables: core.StringMapValue(value),
	}
}

func setCreateFunctionEphemeralStorageSize(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.EphemeralStorage = &types.EphemeralStorage{
		Size: aws.Int32(int32(core.IntValue(value))),
	}
}

func setCreateFunctionFileSystemConfig(
	specData *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	fileSystemConfig := &types.FileSystemConfig{}
	setters := []*pluginutils.ValueSetter[*types.FileSystemConfig]{
		pluginutils.NewValueSetter(
			"$.arn",
			setFileSystemConfigARN,
		),
		pluginutils.NewValueSetter(
			"$.localMountPath",
			setFileSystemConfigLocalMountPath,
		),
	}
	for _, setter := range setters {
		setter.Set(specData, fileSystemConfig)
	}
	input.FileSystemConfigs = []types.FileSystemConfig{*fileSystemConfig}
}

func setCreateFunctionName(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.FunctionName = aws.String(core.StringValue(value))
}

func setCreateFunctionHandler(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.Handler = aws.String(core.StringValue(value))
}

func setCreateFunctionImageConfig(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	imageConfig := &types.ImageConfig{}

	valueSetters := []*pluginutils.ValueSetter[*types.ImageConfig]{
		pluginutils.NewValueSetter(
			"$.command",
			setImageConfigCommand,
		),
		pluginutils.NewValueSetter(
			"$.entryPoint",
			setImageConfigEntrypoint,
		),
		pluginutils.NewValueSetter(
			"$.workingDirectory",
			setImageConfigWorkingDirectory,
		),
	}

	for _, valueSetter := range valueSetters {
		valueSetter.Set(value, imageConfig)
	}

	input.ImageConfig = imageConfig
}

func setCreateFunctionCodeSigningConfigARN(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.CodeSigningConfigArn = aws.String(core.StringValue(value))
}

func setCreateFunctionKMSKeyARN(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.KMSKeyArn = aws.String(core.StringValue(value))
}

func setCreateFunctionLayers(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.Layers = core.StringSliceValue(value)
}

func setCreateFunctionLoggingConfig(
	specData *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	loggingConfig := &types.LoggingConfig{}
	setters := []*pluginutils.ValueSetter[*types.LoggingConfig]{
		pluginutils.NewValueSetter(
			"$.applicationLogLevel",
			setLoggingConfigApplicationLogLevel,
		),
		pluginutils.NewValueSetter(
			"$.logFormat",
			setLoggingConfigLogFormat,
		),
		pluginutils.NewValueSetter(
			"$.logGroup",
			setLoggingConfigLogGroup,
		),
		pluginutils.NewValueSetter(
			"$.systemLogLevel",
			setLoggingConfigSystemLogLevel,
		),
	}
	for _, setter := range setters {
		setter.Set(specData, loggingConfig)
	}
	input.LoggingConfig = loggingConfig
}

func setCreateFunctionMemorySize(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.MemorySize = aws.Int32(int32(core.IntValue(value)))
}

func setCreateFunctionPackageType(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.PackageType = types.PackageType(core.StringValue(value))
}

func setCreateFunctionRole(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.Role = aws.String(core.StringValue(value))
}

func setCreateFunctionRuntime(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.Runtime = types.Runtime(core.StringValue(value))
}

func setCreateFunctionSnapStart(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.SnapStart = &types.SnapStart{
		ApplyOn: types.SnapStartApplyOn(core.StringValue(value)),
	}
}

func setCreateFunctionTimeout(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.Timeout = aws.Int32(int32(core.IntValue(value)))
}

func setCreateFunctionTracingConfig(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	input.TracingConfig = &types.TracingConfig{
		Mode: types.TracingMode(core.StringValue(value)),
	}
}

func setCreateFunctionVPCConfig(
	specData *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	vpcConfig := &types.VpcConfig{}
	setters := []*pluginutils.ValueSetter[*types.VpcConfig]{
		pluginutils.NewValueSetter(
			"$.securityGroupIds",
			setVPCConfigSecurityGroupIds,
		),
		pluginutils.NewValueSetter(
			"$.subnetIds",
			setVPCConfigSubnetIds,
		),
		pluginutils.NewValueSetter(
			"$.ipv6AllowedForDualStack",
			setVPCConfigIPv6AllowedForDualStack,
		),
	}
	for _, setter := range setters {
		setter.Set(specData, vpcConfig)
	}
	input.VpcConfig = vpcConfig
}

func setCreateFunctionTags(
	value *core.MappingNode,
	input *lambda.CreateFunctionInput,
) {
	tags := make(map[string]string)
	for _, item := range value.Items {
		key := core.StringValue(item.Fields["key"])
		value := core.StringValue(item.Fields["value"])
		tags[key] = value
	}
	input.Tags = tags
}
