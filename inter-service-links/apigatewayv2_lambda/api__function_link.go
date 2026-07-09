package apigatewayv2lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// APIToFunctionLinkDeps is the core dependency set for a link from an API Gateway v2 API to a
// Lambda function. Both resources are Cloud Control–backed; the link reads their state and manages
// the integration, route and Lambda permission as Cloud Control intermediaries. It uses the Lambda
// service (resource B) and an IAM service only for the optional WebSocket execute-api:ManageConnections
// grant on the handler's execution role.
type APIToFunctionLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
]

// APIFunctionLink returns a link implementation that routes API Gateway v2 requests to a Lambda
// function. The link manages an AWS_PROXY integration, a route to that integration, and the Lambda
// permission that lets API Gateway invoke the function. For WebSocket APIs it also (opt-in) grants
// the handler's execution role execute-api:ManageConnections so it can push messages to clients.
func APIFunctionLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
) func(APIToFunctionLinkDeps) provider.Link {
	return func(deps APIToFunctionLinkDeps) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/api__function.md")

		actions := &apiFunctionLinkActions{
			awsConfigStore:       deps.ResourceAService.ConfigStore,
			lambdaServiceFactory: deps.ResourceBService.ServiceFactory,
			iamServiceFactory:    iamServiceFactory,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA:                   "aws/apigatewayv2/api",
			ResourceTypeB:                   "aws/lambda/function",
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceA,
			PlainTextSummary:                "A link that routes API Gateway v2 requests to a Lambda function via a managed integration and route.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           apiFunctionLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type apiFunctionLinkActions struct {
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	iamServiceFactory    pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
}

func (l *apiFunctionLinkActions) getRegion(
	ctx context.Context,
	providerContext provider.Context,
) (string, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return "", err
	}
	return awsConfig.Region, nil
}

func (l *apiFunctionLinkActions) getLambdaService(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.lambdaServiceFactory(awsConfig, providerContext), nil
}

func (l *apiFunctionLinkActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.iamServiceFactory(awsConfig, providerContext), nil
}
