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

// AliasResource returns a resource implementation for an AWS Lambda Alias.
func AliasResource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	yamlExample, _ := examples.ReadFile("examples/resources/lambda_alias_basic.md")
	trafficRoutingExample, _ := examples.ReadFile("examples/resources/lambda_alias_traffic_routing.md")
	provisionedConcurrencyExample, _ := examples.ReadFile("examples/resources/lambda_alias_provisioned_concurrency.md")
	completeExample, _ := examples.ReadFile("examples/resources/lambda_alias_complete.md")

	lambdaAliasActions := &lambdaAliasResourceActions{
		lambdaServiceFactory,
		awsConfigStore,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/lambda/alias",
		Label:            "AWS Lambda Alias",
		PlainTextSummary: "A resource for managing an AWS Lambda function alias.",
		FormattedDescription: "The resource type used to define a [Lambda function alias](https://docs.aws.amazon.com/lambda/latest/api/API_CreateAlias.html) " +
			"that is deployed to AWS. An alias is a pointer to a specific Lambda function version and can be used to " +
			"invoke the function with a stable identifier.",
		Schema:  lambdaAliasResourceSchema(),
		IDField: "aliasArn",
		// A Lambda alias is commonly used as a terminal node because it provides
		// a stable endpoint to invoke a specific function version. Other resources
		// like API Gateway or event sources typically reference the alias rather
		// than the function version directly.
		CommonTerminal: true,
		FormattedExamples: []string{
			string(yamlExample),
			string(trafficRoutingExample),
			string(provisionedConcurrencyExample),
			string(completeExample),
		},
		GetExternalStateFunc: lambdaAliasActions.GetExternalState,
		CreateFunc:           lambdaAliasActions.Create,
		UpdateFunc:           lambdaAliasActions.Update,
		DestroyFunc:          lambdaAliasActions.Destroy,
		StabilisedFunc:       lambdaAliasActions.Stabilised,
	}
}

type lambdaAliasResourceActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaAliasResourceActions) getLambdaService(
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
