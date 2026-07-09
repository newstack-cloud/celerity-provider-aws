package apigatewayv2lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// AuthorizerToFunctionLinkDeps is the core dependency set for a link from an API Gateway v2 Lambda
// (REQUEST) authorizer to its backing Lambda function. Both resources are Cloud Control–backed; the
// link reads their state and manages the invoke permission as a Cloud Control intermediary, so it
// uses no AWS service directly (only the shared config store for region).
type AuthorizerToFunctionLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
]

// AuthorizerFunctionLink returns a link implementation that grants API Gateway permission to invoke
// a Lambda (REQUEST) authorizer's backing function. JWT authorizers have no backing function and do
// not use this link.
func AuthorizerFunctionLink() func(AuthorizerToFunctionLinkDeps) provider.Link {
	return func(deps AuthorizerToFunctionLinkDeps) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/authorizer__function.md")

		actions := &authorizerFunctionLinkActions{
			awsConfigStore: deps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA:                   "aws/apigatewayv2/authorizer",
			ResourceTypeB:                   "aws/lambda/function",
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceA,
			PlainTextSummary:                "A link that grants API Gateway permission to invoke a Lambda authorizer's backing function.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           map[string]*provider.LinkAnnotationDefinition{},
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type authorizerFunctionLinkActions struct {
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *authorizerFunctionLinkActions) getRegion(
	ctx context.Context,
	providerContext provider.Context,
) (string, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return "", err
	}
	return awsConfig.Region, nil
}
