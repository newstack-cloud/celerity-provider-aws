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

// EventInvokeConfigResource returns a resource implementation for an AWS Lambda Event Invoke Config.
func EventInvokeConfigResource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	yamlExample, _ := examples.ReadFile("examples/resources/lambda_event_invoke_config_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/lambda_event_invoke_config_complete.md")

	lambdaEventInvokeConfigActions := &lambdaEventInvokeConfigResourceActions{
		lambdaServiceFactory,
		awsConfigStore,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/lambda/eventInvokeConfig",
		Label:            "AWS Lambda Event Invoke Config",
		PlainTextSummary: "A resource for managing AWS Lambda function event invoke configuration.",
		FormattedDescription: "The resource type used to define a [Lambda Event Invoke Config](https://docs.aws.amazon.com/lambda/latest/api/API_PutFunctionEventInvokeConfig.html) " +
			"that configures options for asynchronous invocation on a Lambda function, version, or alias.",
		Schema:  lambdaEventInvokeConfigResourceSchema(),
		IDField: "functionArn",
		// An Event Invoke Config is typically used to configure error handling and destinations
		// for asynchronous function invocations.
		CommonTerminal: false,
		FormattedExamples: []string{
			string(yamlExample),
			string(completeExample),
		},
		GetExternalStateFunc: lambdaEventInvokeConfigActions.GetExternalState,
		CreateFunc:           lambdaEventInvokeConfigActions.Create,
		UpdateFunc:           lambdaEventInvokeConfigActions.Update,
		DestroyFunc:          lambdaEventInvokeConfigActions.Destroy,
		StabilisedFunc:       lambdaEventInvokeConfigActions.Stabilised,
	}
}

type lambdaEventInvokeConfigResourceActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaEventInvokeConfigResourceActions) getLambdaService(
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
