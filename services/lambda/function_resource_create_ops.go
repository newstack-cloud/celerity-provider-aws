package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type functionCreate struct {
	input *lambda.CreateFunctionInput
}

func (u *functionCreate) Name() string {
	return "create function"
}

func (u *functionCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToCreateFunctionInput(
		specData,
	)
	if err != nil {
		return false, saveOpCtx, err
	}

	// Merge Bluelink system tags with user-defined tags
	deployInput, ok := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
	if ok && deployInput != nil {
		input.Tags = mergeBluelinkTagsWithLambdaTags(deployInput, input.Tags)
	}

	u.input = input
	return hasValues, saveOpCtx, nil
}

func (u *functionCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	createFunctionOutput, err := lambdaService.CreateFunction(ctx, u.input)
	if err != nil {
		return saveOpCtx, err
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(
		createFunctionOutput.FunctionArn,
	)
	newSaveOpCtx.Data["createFunctionOutput"] = createFunctionOutput
	newSaveOpCtx.Data["functionARN"] = aws.ToString(createFunctionOutput.FunctionArn)

	return newSaveOpCtx, err
}
