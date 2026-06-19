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

// CodeSigningConfigResource returns a resource implementation for an AWS Lambda Code Signing Configuration.
func CodeSigningConfigResource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := examples.ReadFile("examples/resources/lambda_code_signing_config_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/lambda_code_signing_config_complete.md")

	lambdaCodeSigningConfigActions := &lambdaCodeSigningConfigResourceActions{
		lambdaServiceFactory:               lambdaServiceFactory,
		resourceGroupTaggingServiceFactory: resourceGroupTaggingServiceFactory,
		awsConfigStore:                     awsConfigStore,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/lambda/codeSigningConfig",
		Label:            "AWS Lambda Code Signing Configuration",
		PlainTextSummary: "A resource for managing an AWS Lambda code signing configuration.",
		FormattedDescription: "The resource type used to define a [Lambda code signing configuration](https://docs.aws.amazon.com/lambda/latest/dg/configuration-codesigning.html) " +
			"that is deployed to AWS.",
		Schema:         lambdaCodeSigningConfigResourceSchema(),
		IDField:        "codeSigningConfigArn",
		TaggingSupport: provider.TaggingSupportFull,
		// A code signing configuration is not a terminal resource as it can be used by Lambda functions
		CommonTerminal: false,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
		},
		GetExternalStateFunc: lambdaCodeSigningConfigActions.GetExternalState,
		CreateFunc:           lambdaCodeSigningConfigActions.Create,
		UpdateFunc:           lambdaCodeSigningConfigActions.Update,
		DestroyFunc:          lambdaCodeSigningConfigActions.Destroy,
		StabilisedFunc:       lambdaCodeSigningConfigActions.Stabilised,
	}
}

type lambdaCodeSigningConfigResourceActions struct {
	lambdaServiceFactory               pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service]
	awsConfigStore                     pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaCodeSigningConfigResourceActions) getLambdaService(
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

func (l *lambdaCodeSigningConfigResourceActions) getResourceGroupTaggingService(
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
