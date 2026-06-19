package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// FunctionUrlResource returns a resource implementation for an AWS Lambda Function URL.
func FunctionUrlResource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	yamlExample, _ := examples.ReadFile("examples/resources/lambda_function_url_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/lambda_function_url_complete.md")

	lambdaFunctionUrlActions := &lambdaFunctionUrlResourceActions{
		lambdaServiceFactory,
		awsConfigStore,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/lambda/functionUrl",
		Label:            "AWS Lambda Function URL",
		PlainTextSummary: "A resource for managing an AWS Lambda function URL.",
		FormattedDescription: "The resource type used to define a [Lambda function URL](https://docs.aws.amazon.com/lambda/latest/api/API_CreateFunctionUrl.html) " +
			"that is deployed to AWS. A function URL is a dedicated HTTP(S) endpoint that you can use to invoke your function.",
		Schema:  lambdaFunctionUrlResourceSchema(),
		IDField: "functionUrl",
		// A Lambda function URL is commonly used as a terminal node because it provides
		// a stable HTTP endpoint to invoke a function. Other resources like API Gateway
		// or event sources typically reference the function URL.
		CommonTerminal: true,
		FormattedExamples: []string{
			string(yamlExample),
			string(completeExample),
		},
		GetExternalStateFunc: lambdaFunctionUrlActions.GetExternalState,
		CreateFunc:           lambdaFunctionUrlActions.Create,
		UpdateFunc:           lambdaFunctionUrlActions.Update,
		DestroyFunc:          lambdaFunctionUrlActions.Destroy,
		StabilisedFunc:       lambdaFunctionUrlActions.Stabilised,
	}
}

type lambdaFunctionUrlResourceActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaFunctionUrlResourceActions) getLambdaService(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	return l.lambdaServiceFactory(awsConfig, providerContext), nil
}
