package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/two-hundred/celerity/libs/blueprint/core"
	"github.com/two-hundred/celerity/libs/blueprint/provider"
)

func (l *lambdaFunctionResourceActions) StabilisedFunc(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	functionARN := core.StringValue(
		input.ResourceSpec.Fields["arn"],
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

	lastUpdateStatus := functionOutput.Configuration.LastUpdateStatus
	hasStabilised := lastUpdateStatus == types.LastUpdateStatusSuccessful
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: hasStabilised,
	}, nil
}
