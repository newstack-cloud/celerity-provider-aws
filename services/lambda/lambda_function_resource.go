package lambda

import (
	"context"

	providertypes "github.com/newstack-cloud/celerity-provider-aws/types"
	"github.com/newstack-cloud/celerity-provider-aws/utils"

	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/providerv1"
)

// FunctionResource returns a resource implementation for an AWS Lambda Function.
func FunctionResource(
	lambdaServiceFactory providertypes.ServiceFactory[Service],
	awsConfigStore *utils.AWSConfigStore,
) provider.Resource {
	yamlExample, _ := examples.ReadFile("examples/resources/lambda_function_yaml.md")
	jsoncExample, _ := examples.ReadFile("examples/resources/lambda_function_jsonc.md")
	yamlInlineExample, _ := examples.ReadFile("examples/resources/lambda_function_inline_yaml.md")

	lambdaFunctionActions := &lambdaFunctionResourceActions{
		lambdaServiceFactory,
		awsConfigStore,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/lambda/function",
		Label:            "AWS Lambda Function",
		PlainTextSummary: "A resource for managing an AWS Lambda function.",
		FormattedDescription: "The resource type used to define a [Lambda function](https://docs.aws.amazon.com/lambda/latest/api/API_GetFunction.html) " +
			"that is deployed to AWS.",
		Schema:  lambdaFunctionResourceSchema(),
		IDField: "arn",
		// A Lambda function will usually contain application code that will typically
		// use other resources that can be defined in a blueprint such as an S3 bucket,
		// a DynamoDB table or an SNS topic.
		CommonTerminal: false,
		FormattedExamples: []string{
			string(yamlExample),
			string(jsoncExample),
			string(yamlInlineExample),
		},
		ResourceCanLinkTo:    []string{},
		GetExternalStateFunc: lambdaFunctionActions.GetExternalState,
		CreateFunc:           lambdaFunctionActions.CreateFunc,
		UpdateFunc:           lambdaFunctionActions.UpdateFunc,
		DestroyFunc:          lambdaFunctionActions.DestroyFunc,
		StabilisedFunc:       lambdaFunctionActions.StabilisedFunc,
	}
}

type lambdaFunctionResourceActions struct {
	lambdaServiceFactory providertypes.ServiceFactory[Service]
	awsConfigStore       *utils.AWSConfigStore
}

func (l *lambdaFunctionResourceActions) getLambdaService(
	ctx context.Context,
	providerContext provider.Context,
) (Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext)
	if err != nil {
		return nil, err
	}

	return l.lambdaServiceFactory(awsConfig, providerContext), nil
}
