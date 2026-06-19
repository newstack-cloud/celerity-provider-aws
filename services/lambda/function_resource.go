package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// FunctionResource returns a resource implementation for an AWS Lambda Function.
func FunctionResource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	yamlExample, _ := examples.ReadFile("examples/resources/lambda_function_yaml.md")
	yamlInlineExample, _ := examples.ReadFile("examples/resources/lambda_function_inline_yaml.md")

	lambdaFunctionActions := &lambdaFunctionResourceActions{
		lambdaServiceFactory:               lambdaServiceFactory,
		resourceGroupTaggingServiceFactory: resourceGroupTaggingServiceFactory,
		awsConfigStore:                     awsConfigStore,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/lambda/function",
		Label:            "AWS Lambda Function",
		PlainTextSummary: "A resource for managing an AWS Lambda function.",
		FormattedDescription: "The resource type used to define a [Lambda function](https://docs.aws.amazon.com/lambda/latest/api/API_GetFunction.html) " +
			"that is deployed to AWS.",
		Schema:         lambdaFunctionResourceSchema(),
		IDField:        "arn",
		TaggingSupport: provider.TaggingSupportFull,
		// A Lambda function will usually contain application code that will typically
		// use other resources that can be defined in a blueprint such as an S3 bucket,
		// a DynamoDB table or an SNS topic.
		CommonTerminal: false,
		FormattedExamples: []string{
			string(yamlExample),
			string(yamlInlineExample),
		},
		GetExternalStateFunc: lambdaFunctionActions.GetExternalState,
		CreateFunc:           lambdaFunctionActions.Create,
		UpdateFunc:           lambdaFunctionActions.Update,
		DestroyFunc:          lambdaFunctionActions.Destroy,
		StabilisedFunc:       lambdaFunctionActions.Stabilised,
	}
}

type lambdaFunctionResourceActions struct {
	lambdaServiceFactory               pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service]
	awsConfigStore                     pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaFunctionResourceActions) getLambdaService(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.lambdaServiceFactory(awsConfig, providerContext), nil
}

func (l *lambdaFunctionResourceActions) getResourceGroupTaggingService(
	ctx context.Context,
	providerContext provider.Context,
) (resgrouptagservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.resourceGroupTaggingServiceFactory(awsConfig, providerContext), nil
}
