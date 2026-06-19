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

// LayerVersionPermissionResource returns a resource implementation for an AWS Lambda Layer Version Permission.
func LayerVersionPermissionResource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := examples.ReadFile("examples/resources/lambda_layer_version_permission_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/lambda_layer_version_permission_complete.md")

	lambdaLayerVersionPermissionActions := &lambdaLayerVersionPermissionResourceActions{
		lambdaServiceFactory,
		awsConfigStore,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/lambda/layerVersionPermission",
		Label:            "AWS Lambda Layer Version Permission",
		PlainTextSummary: "A resource for managing AWS Lambda layer version permissions.",
		FormattedDescription: "The resource type used to define [Lambda layer version permissions](https://docs.aws.amazon.com/lambda/latest/api/API_AddLayerVersionPermission.html) " +
			"that grant access to layer versions to other AWS accounts, organizations, or all AWS accounts.",
		Schema:         lambdaLayerVersionPermissionResourceSchema(),
		IDField:        "id",
		CommonTerminal: true,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
		},
		GetExternalStateFunc: lambdaLayerVersionPermissionActions.GetExternalState,
		CreateFunc:           lambdaLayerVersionPermissionActions.Create,
		UpdateFunc:           lambdaLayerVersionPermissionActions.Update,
		DestroyFunc:          lambdaLayerVersionPermissionActions.Destroy,
		StabilisedFunc:       lambdaLayerVersionPermissionActions.Stabilised,
	}
}

type lambdaLayerVersionPermissionResourceActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaLayerVersionPermissionResourceActions) getLambdaService(
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
