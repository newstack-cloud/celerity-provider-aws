package lambda

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/celerity-provider-aws/utils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// arn is the ID field that must be present in order to update the resource.
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(input.Changes)
	arnValue, err := core.GetPathValue(
		"$.arn",
		currentStateSpecData,
		core.MappingNodeMaxTraverseDepth,
	)
	if err != nil {
		return nil, err
	}

	arn := core.StringValue(arnValue)

	updateOperations := []pluginutils.SaveOperation[Service]{
		&functionConfigUpdate{},
		&functionCodeUpdate{},
		&functionCodeSigningConfigUpdate{},
		&functionConcurrencyUpdate{},
		&functionRecursionConfigUpdate{},
		&functionRuntimeManagementConfigUpdate{},
		&tagsUpdate{
			pathRoot: "$.tags",
		},
	}

	hasUpdates, _, err := pluginutils.RunSaveOperations(
		ctx,
		pluginutils.SaveOperationContext{
			ProviderUpstreamID: arn,
			Data:               map[string]any{},
		},
		updateOperations,
		input,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	if hasUpdates {
		getFunctionOutput, err := lambdaService.GetFunction(ctx, &lambda.GetFunctionInput{
			FunctionName: &arn,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get updated function configuration: %w", err)
		}

		computedFields := l.extractComputedFieldsFromFunctionConfig(
			getFunctionOutput.Configuration,
		)
		return &provider.ResourceDeployOutput{
			ComputedFieldValues: computedFields,
		}, nil
	}

	// If no updates were made, return the current computed fields from the current state.
	currentStateComputedFields := l.extractComputedFieldsFromCurrentState(
		currentStateSpecData,
	)
	return &provider.ResourceDeployOutput{
		ComputedFieldValues: currentStateComputedFields,
	}, nil
}

func (l *lambdaFunctionResourceActions) extractComputedFieldsFromFunctionConfig(
	functionConfiguration *types.FunctionConfiguration,
) map[string]*core.MappingNode {
	return extractComputedFieldsFromFunctionConfig(functionConfiguration)
}

func (l *lambdaFunctionResourceActions) extractComputedFieldsFromCurrentState(
	currentStateSpecData *core.MappingNode,
) map[string]*core.MappingNode {
	fields := map[string]*core.MappingNode{}
	if v, ok := pluginutils.GetValueByPath("$.arn", currentStateSpecData); ok {
		fields["spec.arn"] = v
	}

	if v, ok := pluginutils.GetValueByPath(
		"$.snapStartResponseApplyOn",
		currentStateSpecData,
	); ok {
		fields["spec.snapStartResponseApplyOn"] = v
	}

	if v, ok := pluginutils.GetValueByPath(
		"$.snapStartResponseOptimizationStatus",
		currentStateSpecData,
	); ok {
		fields["spec.snapStartResponseOptimizationStatus"] = v
	}

	return fields
}

func changesToUpdateFunctionInput(
	arn string,
	updatedSpecData *core.MappingNode,
	changes *provider.Changes,
) (*lambda.UpdateFunctionConfigurationInput, bool) {
	modifiedFields := pluginutils.MergeFieldChanges(changes.ModifiedFields, changes.NewFields)
	input := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: &arn,
	}

	valueSetters := []*pluginutils.ValueSetter[*lambda.UpdateFunctionConfigurationInput]{
		pluginutils.NewValueSetter(
			"$.deadLetterConfig.targetArn",
			setUpdateFunctionConfigDeadLetterConfigTargetARN,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionConfigurationInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionConfigurationInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.description",
			setUpdateFunctionConfigDescription,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionConfigurationInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionConfigurationInput](
				modifiedFields,
				"spec",
			),
		),
		// We won't check if the environment variables have changed
		// as the change set will have a modified fields entry for every
		// key pair in the environment variables map and not the map as a whole,
		// it will generally be cheaper to replace the entire map than to compare
		// each key/value pair.
		pluginutils.NewValueSetter(
			"$.environment.variables",
			setUpdateFunctionConfigEnvironmentVariables,
		),
		pluginutils.NewValueSetter(
			"$.ephemeralStorage.size",
			setUpdateFunctionConfigEphemeralStorageSize,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionConfigurationInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionConfigurationInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.fileSystemConfig",
			func(value *core.MappingNode, target *lambda.UpdateFunctionConfigurationInput) {
				setUpdateFunctionConfigFileSystemConfig(value, target, changes, "spec.fileSystemConfig")
			},
		),
		pluginutils.NewValueSetter(
			"$.handler",
			setUpdateFunctionConfigHandler,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionConfigurationInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionConfigurationInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.imageConfig",
			func(value *core.MappingNode, target *lambda.UpdateFunctionConfigurationInput) {
				setUpdateFunctionConfigImageConfig(value, target, changes, "spec.imageConfig")
			},
		),
		pluginutils.NewValueSetter(
			"$.kmsKeyArn",
			setUpdateFunctionConfigKMSKeyARN,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionConfigurationInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionConfigurationInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.layers",
			setUpdateFunctionConfigLayers,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionConfigurationInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionConfigurationInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.loggingConfig",
			func(value *core.MappingNode, target *lambda.UpdateFunctionConfigurationInput) {
				setUpdateFunctionConfigLoggingConfig(value, target, changes, "spec.loggingConfig")
			},
		),
		pluginutils.NewValueSetter(
			"$.memorySize",
			setUpdateFunctionConfigMemorySize,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionConfigurationInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionConfigurationInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.role",
			setUpdateFunctionConfigRole,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionConfigurationInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionConfigurationInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.runtime",
			setUpdateFunctionConfigRuntime,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionConfigurationInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionConfigurationInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.snapStart",
			setUpdateFunctionConfigSnapStart,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionConfigurationInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionConfigurationInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.timeout",
			setUpdateFunctionConfigTimeout,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionConfigurationInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionConfigurationInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.tracingConfig",
			setUpdateFunctionConfigTracingConfig,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionConfigurationInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionConfigurationInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.vpcConfig",
			func(value *core.MappingNode, target *lambda.UpdateFunctionConfigurationInput) {
				setUpdateFunctionConfigVPCConfig(value, target, changes, "spec.vpcConfig")
			},
		),
	}

	hasUpdates := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(updatedSpecData, input)
		hasUpdates = hasUpdates || valueSetter.DidSet()
	}

	return input, hasUpdates
}

func setUpdateFunctionConfigDeadLetterConfigTargetARN(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
) {
	input.DeadLetterConfig = &types.DeadLetterConfig{
		TargetArn: aws.String(
			core.StringValue(value),
		),
	}
}

func setUpdateFunctionConfigDescription(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
) {
	input.Description = aws.String(
		core.StringValue(value),
	)
}

func setUpdateFunctionConfigEnvironmentVariables(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
) {
	input.Environment = &types.Environment{
		Variables: core.StringMapValue(value),
	}
}

func setUpdateFunctionConfigEphemeralStorageSize(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
) {
	input.EphemeralStorage = &types.EphemeralStorage{
		Size: aws.Int32(int32(core.IntValue(value))),
	}
}

func setUpdateFunctionConfigFileSystemConfig(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
	changes *provider.Changes,
	pathRoot string,
) {
	modifiedFields := pluginutils.MergeFieldChanges(changes.ModifiedFields, changes.NewFields)
	fileSystemConfig := &types.FileSystemConfig{}

	valueSetters := []*pluginutils.ValueSetter[*types.FileSystemConfig]{
		pluginutils.NewValueSetter(
			"$.arn",
			setFileSystemConfigARN,
			pluginutils.WithValueSetterCheckIfChanged[*types.FileSystemConfig](true),
			pluginutils.WithValueSetterModifiedFields[*types.FileSystemConfig](
				modifiedFields,
				pathRoot,
			),
		),
		pluginutils.NewValueSetter(
			"$.localMountPath",
			setFileSystemConfigLocalMountPath,
			pluginutils.WithValueSetterCheckIfChanged[*types.FileSystemConfig](true),
			pluginutils.WithValueSetterModifiedFields[*types.FileSystemConfig](
				modifiedFields,
				pathRoot,
			),
		),
	}

	for _, valueSetter := range valueSetters {
		valueSetter.Set(value, fileSystemConfig)
	}

	input.FileSystemConfigs = []types.FileSystemConfig{
		*fileSystemConfig,
	}
}

func setUpdateFunctionConfigHandler(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
) {
	input.Handler = aws.String(core.StringValue(value))
}

func setUpdateFunctionConfigImageConfig(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
	changes *provider.Changes,
	pathRoot string,
) {
	modifiedFields := pluginutils.MergeFieldChanges(changes.ModifiedFields, changes.NewFields)
	imageConfig := &types.ImageConfig{}

	valueSetters := []*pluginutils.ValueSetter[*types.ImageConfig]{
		pluginutils.NewValueSetter(
			"$.command",
			setImageConfigCommand,
			pluginutils.WithValueSetterCheckIfChanged[*types.ImageConfig](true),
			pluginutils.WithValueSetterModifiedFields[*types.ImageConfig](
				modifiedFields,
				pathRoot,
			),
		),
		pluginutils.NewValueSetter(
			"$.entryPoint",
			setImageConfigEntrypoint,
			pluginutils.WithValueSetterCheckIfChanged[*types.ImageConfig](true),
			pluginutils.WithValueSetterModifiedFields[*types.ImageConfig](
				modifiedFields,
				pathRoot,
			),
		),
		pluginutils.NewValueSetter(
			"$.workingDirectory",
			setImageConfigWorkingDirectory,
			pluginutils.WithValueSetterCheckIfChanged[*types.ImageConfig](true),
			pluginutils.WithValueSetterModifiedFields[*types.ImageConfig](
				modifiedFields,
				pathRoot,
			),
		),
	}

	for _, valueSetter := range valueSetters {
		valueSetter.Set(value, imageConfig)
	}

	input.ImageConfig = imageConfig
}

func setUpdateFunctionConfigKMSKeyARN(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
) {
	input.KMSKeyArn = aws.String(core.StringValue(value))
}

func setUpdateFunctionConfigLayers(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
) {
	input.Layers = core.StringSliceValue(value)
}

func setUpdateFunctionConfigLoggingConfig(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
	changes *provider.Changes,
	pathRoot string,
) {
	modifiedFields := pluginutils.MergeFieldChanges(changes.ModifiedFields, changes.NewFields)
	loggingConfig := &types.LoggingConfig{}

	valueSetters := []*pluginutils.ValueSetter[*types.LoggingConfig]{
		pluginutils.NewValueSetter(
			"$.applicationLogLevel",
			setLoggingConfigApplicationLogLevel,
			pluginutils.WithValueSetterCheckIfChanged[*types.LoggingConfig](true),
			pluginutils.WithValueSetterModifiedFields[*types.LoggingConfig](
				modifiedFields,
				pathRoot,
			),
		),
		pluginutils.NewValueSetter(
			"$.logFormat",
			setLoggingConfigLogFormat,
			pluginutils.WithValueSetterCheckIfChanged[*types.LoggingConfig](true),
			pluginutils.WithValueSetterModifiedFields[*types.LoggingConfig](
				modifiedFields,
				pathRoot,
			),
		),
		pluginutils.NewValueSetter(
			"$.logGroup",
			setLoggingConfigLogGroup,
			pluginutils.WithValueSetterCheckIfChanged[*types.LoggingConfig](true),
			pluginutils.WithValueSetterModifiedFields[*types.LoggingConfig](
				modifiedFields,
				pathRoot,
			),
		),
		pluginutils.NewValueSetter(
			"$.systemLogLevel",
			setLoggingConfigSystemLogLevel,
			pluginutils.WithValueSetterCheckIfChanged[*types.LoggingConfig](true),
			pluginutils.WithValueSetterModifiedFields[*types.LoggingConfig](
				modifiedFields,
				pathRoot,
			),
		),
	}

	for _, valueSetter := range valueSetters {
		valueSetter.Set(value, loggingConfig)
	}

	input.LoggingConfig = loggingConfig
}

func setUpdateFunctionConfigMemorySize(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
) {
	input.MemorySize = aws.Int32(int32(core.IntValue(value)))
}

func setUpdateFunctionConfigRole(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
) {
	input.Role = aws.String(core.StringValue(value))
}

func setUpdateFunctionConfigRuntime(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
) {
	input.Runtime = types.Runtime(core.StringValue(value))
}

func setUpdateFunctionConfigSnapStart(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
) {
	input.SnapStart = &types.SnapStart{
		ApplyOn: types.SnapStartApplyOn(core.StringValue(value)),
	}
}

func setUpdateFunctionConfigTimeout(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
) {
	input.Timeout = aws.Int32(int32(core.IntValue(value)))
}

func setUpdateFunctionConfigTracingConfig(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
) {
	input.TracingConfig = &types.TracingConfig{
		Mode: types.TracingMode(core.StringValue(value)),
	}
}

func setUpdateFunctionConfigVPCConfig(
	value *core.MappingNode,
	input *lambda.UpdateFunctionConfigurationInput,
	changes *provider.Changes,
	pathRoot string,
) {
	modifiedFields := pluginutils.MergeFieldChanges(changes.ModifiedFields, changes.NewFields)
	vpcConfig := &types.VpcConfig{}

	valueSetters := []*pluginutils.ValueSetter[*types.VpcConfig]{
		pluginutils.NewValueSetter(
			"$.securityGroupIds",
			func(value *core.MappingNode, target *types.VpcConfig) {
				target.SecurityGroupIds = core.StringSliceValue(value)
			},
			pluginutils.WithValueSetterCheckIfChanged[*types.VpcConfig](true),
			pluginutils.WithValueSetterModifiedFields[*types.VpcConfig](
				modifiedFields,
				pathRoot,
			),
		),
		pluginutils.NewValueSetter(
			"$.subnetIds",
			func(value *core.MappingNode, target *types.VpcConfig) {
				target.SubnetIds = core.StringSliceValue(value)
			},
			pluginutils.WithValueSetterCheckIfChanged[*types.VpcConfig](true),
			pluginutils.WithValueSetterModifiedFields[*types.VpcConfig](
				modifiedFields,
				pathRoot,
			),
		),
		pluginutils.NewValueSetter(
			"$.ipv6AllowedForDualStack",
			func(value *core.MappingNode, target *types.VpcConfig) {
				target.Ipv6AllowedForDualStack = aws.Bool(core.BoolValue(value))
			},
			pluginutils.WithValueSetterCheckIfChanged[*types.VpcConfig](true),
			pluginutils.WithValueSetterModifiedFields[*types.VpcConfig](
				modifiedFields,
				pathRoot,
			),
		),
	}

	for _, valueSetter := range valueSetters {
		valueSetter.Set(value, vpcConfig)
	}

	input.VpcConfig = vpcConfig
}

func changesToUpdateFunctionCodeInput(
	arn string,
	updatedSpecData *core.MappingNode,
	changes *provider.Changes,
) (*lambda.UpdateFunctionCodeInput, bool, error) {
	modifiedFields := pluginutils.MergeFieldChanges(changes.ModifiedFields, changes.NewFields)

	input := &lambda.UpdateFunctionCodeInput{
		FunctionName: &arn,
		// Code updates from a change in the blueprint will always be published
		// for the "current" version of the function represented by the function resource,
		// versioning is enabled by using the separate function version resources.
		Publish: true,
	}

	runtime, _ := pluginutils.GetValueByPath(
		"$.runtime",
		updatedSpecData,
	)
	err := prepareZipFormatForInlineCode(
		updatedSpecData,
		core.StringValue(runtime),
	)
	if err != nil {
		return nil, false, err
	}

	valueSetters := []*pluginutils.ValueSetter[*lambda.UpdateFunctionCodeInput]{
		pluginutils.NewValueSetter(
			"$.architecture",
			setUpdateFunctionCodeArchitecture,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionCodeInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionCodeInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.code.imageUri",
			setUpdateFunctionCodeImageUri,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionCodeInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionCodeInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.code.s3Bucket",
			setUpdateFunctionCodeS3Bucket,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionCodeInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionCodeInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.code.s3Key",
			setUpdateFunctionCodeS3Key,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionCodeInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionCodeInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.code.s3ObjectVersion",
			setUpdateFunctionCodeS3ObjectVersion,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionCodeInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionCodeInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.code.sourceKMSKeyArn",
			setUpdateFunctionCodeSourceKMSKeyARN,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionCodeInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionCodeInput](
				modifiedFields,
				"spec",
			),
		),
		pluginutils.NewValueSetter(
			"$.code.zipFile",
			setUpdateFunctionCodeZipFile,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.UpdateFunctionCodeInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.UpdateFunctionCodeInput](
				modifiedFields,
				"spec",
			),
		),
	}

	hasUpdates := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(updatedSpecData, input)
		hasUpdates = hasUpdates || valueSetter.DidSet()
	}

	return input, hasUpdates, nil
}

func prepareZipFormatForInlineCode(
	inputSpecData *core.MappingNode,
	runtime string,
) error {
	zipFile, hasZipFile := pluginutils.GetValueByPath(
		"$.code.zipFile",
		inputSpecData,
	)
	if !hasZipFile {
		return nil
	}

	language := getLanguageFromRuntime(runtime)
	if language == "python" || language == "nodejs" {
		extension := getExtensionFromLanguage(language)
		fileName := fmt.Sprintf("index.%s", extension)
		zipB64Encoded, err := utils.ZipInMemory(fileName, core.StringValue(zipFile))
		if err != nil {
			return err
		}
		inputSpecData.Fields["zipFile"] = core.MappingNodeFromString(zipB64Encoded)

		return nil
	}

	return fmt.Errorf(
		"inline code is only supported for Node.js and Python runtimes, "+
			"the %s runtime can not be used with inline code",
		runtime,
	)
}

func getLanguageFromRuntime(runtime string) string {
	if strings.HasPrefix(runtime, "nodejs") {
		return "nodejs"
	} else if strings.HasPrefix(runtime, "python") {
		return "python"
	}

	return ""
}

func getExtensionFromLanguage(language string) string {
	switch language {
	case "nodejs":
		return "js"
	case "python":
		return "py"
	}
	return ""
}

func setUpdateFunctionCodeArchitecture(
	value *core.MappingNode,
	input *lambda.UpdateFunctionCodeInput,
) {
	input.Architectures = []types.Architecture{
		types.Architecture(core.StringValue(value)),
	}
}

func setUpdateFunctionCodeImageUri(
	value *core.MappingNode,
	input *lambda.UpdateFunctionCodeInput,
) {
	input.ImageUri = aws.String(core.StringValue(value))
}

func setUpdateFunctionCodeS3Bucket(
	value *core.MappingNode,
	input *lambda.UpdateFunctionCodeInput,
) {
	input.S3Bucket = aws.String(core.StringValue(value))
}

func setUpdateFunctionCodeS3Key(
	value *core.MappingNode,
	input *lambda.UpdateFunctionCodeInput,
) {
	input.S3Key = aws.String(core.StringValue(value))
}

func setUpdateFunctionCodeS3ObjectVersion(
	value *core.MappingNode,
	input *lambda.UpdateFunctionCodeInput,
) {
	input.S3ObjectVersion = aws.String(core.StringValue(value))
}

func setUpdateFunctionCodeSourceKMSKeyARN(
	value *core.MappingNode,
	input *lambda.UpdateFunctionCodeInput,
) {
	input.SourceKMSKeyArn = aws.String(core.StringValue(value))
}

func setUpdateFunctionCodeZipFile(
	value *core.MappingNode,
	input *lambda.UpdateFunctionCodeInput,
) {
	input.ZipFile = []byte(core.StringValue(value))
}

func changesToPutFunctionCodeSigningConfigInput(
	arn string,
	updatedSpecData *core.MappingNode,
) (*lambda.PutFunctionCodeSigningConfigInput, bool) {
	codeSigningConfigArn, hasCodeSigningConfigArn := pluginutils.GetValueByPath(
		"$.codeSigningConfig.codeSigningConfigArn",
		updatedSpecData,
	)
	if !hasCodeSigningConfigArn {
		return nil, false
	}

	return &lambda.PutFunctionCodeSigningConfigInput{
		FunctionName:         &arn,
		CodeSigningConfigArn: aws.String(core.StringValue(codeSigningConfigArn)),
	}, true
}

func changesToPutFunctionConcurrencyInput(
	arn string,
	putFunctionConcurrenyData *core.MappingNode,
) (*lambda.PutFunctionConcurrencyInput, bool) {
	reservedConcurrentExecutions, hasReservedConcurrentExecutions := pluginutils.GetValueByPath(
		"$.reservedConcurrentExecutions",
		putFunctionConcurrenyData,
	)
	if !hasReservedConcurrentExecutions {
		return nil, false
	}

	return &lambda.PutFunctionConcurrencyInput{
		FunctionName: &arn,
		ReservedConcurrentExecutions: aws.Int32(
			int32(core.IntValue(reservedConcurrentExecutions)),
		),
	}, true
}

func changesToPutFunctionRecursionConfigInput(
	arn string,
	putFunctionRecursionConfigData *core.MappingNode,
) (*lambda.PutFunctionRecursionConfigInput, bool) {
	recursiveLoop, hasRecursiveLoop := pluginutils.GetValueByPath(
		"$.recursiveLoop",
		putFunctionRecursionConfigData,
	)
	if !hasRecursiveLoop {
		return nil, false
	}

	return &lambda.PutFunctionRecursionConfigInput{
		FunctionName:  &arn,
		RecursiveLoop: types.RecursiveLoop(core.StringValue(recursiveLoop)),
	}, true
}

func changesToPutRuntimeMgmtConfigInput(
	arn string,
	putRuntimeMgmtConfigData *core.MappingNode,
	changes *provider.Changes,
) (*lambda.PutRuntimeManagementConfigInput, bool) {
	modifiedFields := pluginutils.MergeFieldChanges(changes.ModifiedFields, changes.NewFields)

	pathRoot := "spec.runtimeManagementConfig"
	input := &lambda.PutRuntimeManagementConfigInput{
		FunctionName: &arn,
	}

	valueSetters := []*pluginutils.ValueSetter[*lambda.PutRuntimeManagementConfigInput]{
		pluginutils.NewValueSetter(
			"$.runtimeVersionArn",
			setUpdateFunctionConfigRuntimeVersionARN,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.PutRuntimeManagementConfigInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.PutRuntimeManagementConfigInput](
				modifiedFields,
				pathRoot,
			),
		),
		pluginutils.NewValueSetter(
			"$.updateRuntimeOn",
			setUpdateFunctionConfigUpdateRuntimeOn,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.PutRuntimeManagementConfigInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.PutRuntimeManagementConfigInput](
				modifiedFields,
				pathRoot,
			),
		),
	}

	hasUpdates := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(putRuntimeMgmtConfigData, input)
		hasUpdates = hasUpdates || valueSetter.DidSet()
	}

	return input, hasUpdates
}

func setUpdateFunctionConfigRuntimeVersionARN(
	value *core.MappingNode,
	input *lambda.PutRuntimeManagementConfigInput,
) {
	input.RuntimeVersionArn = aws.String(core.StringValue(value))
}

func setUpdateFunctionConfigUpdateRuntimeOn(
	value *core.MappingNode,
	input *lambda.PutRuntimeManagementConfigInput,
) {
	input.UpdateRuntimeOn = types.UpdateRuntimeOn(core.StringValue(value))
}
