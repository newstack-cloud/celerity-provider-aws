package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

type functionConfigUpdate struct {
	input *lambda.UpdateFunctionConfigurationInput
}

func (u *functionConfigUpdate) Name() string {
	return "function configuration"
}

func (u *functionConfigUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasUpdates := changesToUpdateFunctionInput(
		saveOpCtx.ProviderUpstreamID,
		specData,
		changes,
	)
	u.input = input
	return hasUpdates, saveOpCtx, nil
}

func (u *functionConfigUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := lambdaService.UpdateFunctionConfiguration(ctx, u.input)
	return saveOpCtx, err
}

type functionCodeUpdate struct {
	input *lambda.UpdateFunctionCodeInput
}

func (u *functionCodeUpdate) Name() string {
	return "function code"
}

func (u *functionCodeUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasUpdates, err := changesToUpdateFunctionCodeInput(
		saveOpCtx.ProviderUpstreamID,
		specData,
		changes,
	)
	if err != nil {
		return false, saveOpCtx, err
	}
	u.input = input
	return hasUpdates, saveOpCtx, nil
}

func (u *functionCodeUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := lambdaService.UpdateFunctionCode(ctx, u.input)
	return saveOpCtx, err
}

type functionCodeSigningConfigUpdate struct {
	input *lambda.PutFunctionCodeSigningConfigInput
}

func (u *functionCodeSigningConfigUpdate) Name() string {
	return "code signing config"
}

func (u *functionCodeSigningConfigUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	_ *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasUpdates := changesToPutFunctionCodeSigningConfigInput(
		saveOpCtx.ProviderUpstreamID,
		specData,
	)
	u.input = input
	return hasUpdates, saveOpCtx, nil
}

func (u *functionCodeSigningConfigUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := lambdaService.PutFunctionCodeSigningConfig(ctx, u.input)
	return saveOpCtx, err
}

type functionConcurrencyUpdate struct {
	input *lambda.PutFunctionConcurrencyInput
}

func (u *functionConcurrencyUpdate) Name() string {
	return "function concurrency"
}

func (u *functionConcurrencyUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	_ *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasUpdates := changesToPutFunctionConcurrencyInput(
		saveOpCtx.ProviderUpstreamID,
		specData,
	)
	u.input = input
	return hasUpdates, saveOpCtx, nil
}

func (u *functionConcurrencyUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := lambdaService.PutFunctionConcurrency(ctx, u.input)
	return saveOpCtx, err
}

type functionRecursionConfigUpdate struct {
	input *lambda.PutFunctionRecursionConfigInput
}

func (u *functionRecursionConfigUpdate) Name() string {
	return "function recursion config"
}

func (u *functionRecursionConfigUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	_ *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasUpdates := changesToPutFunctionRecursionConfigInput(
		saveOpCtx.ProviderUpstreamID,
		specData,
	)
	u.input = input
	return hasUpdates, saveOpCtx, nil
}

func (u *functionRecursionConfigUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := lambdaService.PutFunctionRecursionConfig(ctx, u.input)
	return saveOpCtx, err
}

type functionRuntimeManagementConfigUpdate struct {
	input *lambda.PutRuntimeManagementConfigInput
}

func (u *functionRuntimeManagementConfigUpdate) Name() string {
	return "runtime management config"
}

func (u *functionRuntimeManagementConfigUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	runtimeMgmtConfigData, _ := pluginutils.GetValueByPath(
		"$.runtimeManagementConfig",
		specData,
	)
	input, hasUpdates := changesToPutRuntimeMgmtConfigInput(
		saveOpCtx.ProviderUpstreamID,
		runtimeMgmtConfigData,
		changes,
	)
	u.input = input
	return hasUpdates, saveOpCtx, nil
}

func (u *functionRuntimeManagementConfigUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := lambdaService.PutRuntimeManagementConfig(ctx, u.input)
	return saveOpCtx, err
}
