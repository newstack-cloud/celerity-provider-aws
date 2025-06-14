package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (l *lambdaFunctionResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	functionARN := core.StringValue(
		input.ResourceState.SpecData.Fields["arn"],
	)
	_, err = lambdaService.DeleteFunction(
		ctx,
		&lambda.DeleteFunctionInput{
			FunctionName: &functionARN,
		},
	)

	return err
}
