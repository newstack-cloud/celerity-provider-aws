package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/celerity-provider-aws/utils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (l *lambdaFunctionResourceActions) UpdateFunc(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// arn is the ID field that must be present in order to update the resource.
	currentStateSpecData := utils.GetCurrentResourceStateSpecData(input.Changes)
	arn, err := core.GetPathValue(
		"$.arn",
		currentStateSpecData,
		core.MappingNodeMaxTraverseDepth,
	)
	if err != nil {
		return nil, err
	}

	resolvedResourceSpecData := utils.GetResolvedResourceSpecData(input.Changes)
	return l.updateFunctionResource(
		ctx,
		core.StringValue(arn),
		lambdaService,
		resolvedResourceSpecData,
	)
}

func (l *lambdaFunctionResourceActions) updateFunctionResource(
	ctx context.Context,
	arn string,
	lambdaService Service,
	updatedSpecData *core.MappingNode,
) (*provider.ResourceDeployOutput, error) {
	updateFunctionInput, hasFuncConfigUpdates, err := changesToUpdateFunctionInput(
		arn,
		updatedSpecData,
	)
	if err != nil {
		return nil, err
	}

	if hasFuncConfigUpdates {
		_, err := lambdaService.UpdateFunctionConfiguration(ctx, updateFunctionInput)
		if err != nil {
			return nil, err
		}
	}
	// Collect modified fields
	// Collect deleted fields
	// Collect new fields

	return nil, nil
}

func changesToUpdateFunctionInput(
	arn string,
	updatedSpecData *core.MappingNode,
) (*lambda.UpdateFunctionConfigurationInput, bool, error) {
	input := &lambda.UpdateFunctionConfigurationInput{
		FunctionName: &arn,
	}

	valueSetters := []utils.SpecValueSetter[*lambda.UpdateFunctionConfigurationInput]{
		{
			PathInSpec:   "$.deadLetterConfig.targetArn",
			SetValueFunc: setUpdateFunctionConfigDeadLetterConfigTargetARN,
		},
		{
			PathInSpec:   "$.description",
			SetValueFunc: setUpdateFunctionConfigDescription,
		},
		{
			PathInSpec:   "$.environment.variables",
			SetValueFunc: setUpdateFunctionConfigEnvironmentVariables,
		},
		{
			PathInSpec:   "$.ephemeralStorage.size",
			SetValueFunc: setUpdateFunctionConfigEphemeralStorageSize,
		},
		{
			PathInSpec:   "$.fileSystemConfig",
			SetValueFunc: setUpdateFunctionConfigFileSystemConfig,
		},
		{
			PathInSpec:   "$.handler",
			SetValueFunc: setUpdateFunctionConfigHandler,
		},
		{
			PathInSpec:   "$.imageConfig",
			SetValueFunc: setUpdateFunctionConfigImageConfig,
		},
	}

	hasUpdates := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(updatedSpecData, input)
		hasUpdates = hasUpdates || valueSetter.DidSet()
	}

	return input, hasUpdates, nil
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
		Variables: mappingNodeToLambdaEnvVars(value),
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
) {
	fileSystemConfig := &types.FileSystemConfig{}

	valueSetters := []utils.SpecValueSetter[*types.FileSystemConfig]{
		{
			PathInSpec: "$.arn",
			SetValueFunc: func(value *core.MappingNode, target *types.FileSystemConfig) {
				target.Arn = aws.String(core.StringValue(value))
			},
		},
		{
			PathInSpec: "$.localMountPath",
			SetValueFunc: func(value *core.MappingNode, target *types.FileSystemConfig) {
				target.LocalMountPath = aws.String(core.StringValue(value))
			},
		},
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
) {
	imageConfig := &types.ImageConfig{}

	valueSetters := []utils.SpecValueSetter[*types.ImageConfig]{
		{
			PathInSpec: "$.command",
			SetValueFunc: func(value *core.MappingNode, target *types.ImageConfig) {
				target.Command = core.StringSliceValue(value)
			},
		},
		{
			PathInSpec: "$.entryPoint",
			SetValueFunc: func(value *core.MappingNode, target *types.ImageConfig) {
				target.EntryPoint = core.StringSliceValue(value)
			},
		},
		{
			PathInSpec: "$.workingDirectory",
			SetValueFunc: func(value *core.MappingNode, target *types.ImageConfig) {
				target.WorkingDirectory = aws.String(core.StringValue(value))
			},
		},
	}

	for _, valueSetter := range valueSetters {
		valueSetter.Set(value, imageConfig)
	}

	input.ImageConfig = imageConfig
}

func mappingNodeToLambdaEnvVars(envVars *core.MappingNode) map[string]string {
	envVarsMap := make(map[string]string)
	for key, value := range envVars.Fields {
		envVarsMap[key] = core.StringValue(value)
	}
	return envVarsMap
}
