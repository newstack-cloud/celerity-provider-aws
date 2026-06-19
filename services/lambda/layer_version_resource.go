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

// LayerVersionResource returns a resource implementation for an AWS Lambda Layer Version.
func LayerVersionResource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := examples.ReadFile("examples/resources/lambda_layer_version_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/lambda_layer_version_complete.md")

	lambdaLayerVersionActions := &lambdaLayerVersionResourceActions{
		lambdaServiceFactory,
		awsConfigStore,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/lambda/layerVersion",
		Label:            "AWS Lambda Layer Version",
		PlainTextSummary: "A resource for managing an AWS Lambda layer version.",
		FormattedDescription: "The resource type used to define a [Lambda layer version](https://docs.aws.amazon.com/lambda/latest/api/API_PublishLayerVersion.html) " +
			"that is deployed to AWS. Each time you publish a layer with the same name, a new version is created.",
		Schema:  lambdaLayerVersionResourceSchema(),
		IDField: "layerVersionArn",
		// A Lambda layer version is commonly used as a terminal node because it represents
		// an immutable snapshot of layer code and configuration. While functions can reference
		// the layer version through the blueprint framework's linking mechanism, the version
		// itself is typically the end point.
		CommonTerminal: true,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
		},
		GetExternalStateFunc: lambdaLayerVersionActions.GetExternalState,
		CreateFunc:           lambdaLayerVersionActions.Create,
		UpdateFunc:           lambdaLayerVersionActions.Update,
		DestroyFunc:          lambdaLayerVersionActions.Destroy,
		StabilisedFunc:       lambdaLayerVersionActions.Stabilised,
	}
}

type lambdaLayerVersionResourceActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaLayerVersionResourceActions) getLambdaService(
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
