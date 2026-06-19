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

// FunctionVersionResource returns a resource implementation for an AWS Lambda Function Version.
func FunctionVersionResource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := examples.ReadFile("examples/resources/lambda_function_version_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/lambda_function_version_complete.md")

	lambdaFunctionVersionActions := &lambdaFunctionVersionResourceActions{
		lambdaServiceFactory,
		awsConfigStore,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/lambda/functionVersion",
		Label:            "AWS Lambda Function Version",
		PlainTextSummary: "A resource for managing an AWS Lambda function version.",
		FormattedDescription: "The resource type used to define a [Lambda function version](https://docs.aws.amazon.com/lambda/latest/api/API_PublishVersion.html) " +
			"that is deployed to AWS.",
		Schema:  lambdaFunctionVersionResourceSchema(),
		IDField: "functionArnWithVersion",
		// A Lambda function version is commonly used as a terminal node because it represents
		// an immutable snapshot of a function's code and configuration. While other resources
		// like event sources or API Gateway can link to the function version through the
		// blueprint framework's linking mechanism, the version itself is typically the end
		// point.
		CommonTerminal: true,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
		},
		GetExternalStateFunc: lambdaFunctionVersionActions.GetExternalState,
		CreateFunc:           lambdaFunctionVersionActions.Create,
		UpdateFunc:           lambdaFunctionVersionActions.Update,
		DestroyFunc:          lambdaFunctionVersionActions.Destroy,
		StabilisedFunc:       lambdaFunctionVersionActions.Stabilised,
	}
}

type lambdaFunctionVersionResourceActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaFunctionVersionResourceActions) getLambdaService(
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
