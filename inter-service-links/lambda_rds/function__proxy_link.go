package lambdards

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// FunctionToProxyLinkDeps is a type alias for the core dependencies for a link from a Lambda
// function to an RDS Proxy. The proxy is a Cloud Control–backed resource, so resource B's
// service is the Cloud Control service; the link only reads the proxy's endpoint, ARN and
// security group from state and does not call it.
type FunctionToProxyLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	cloudcontrolservice.Service,
]

// FunctionProxyLink returns a link implementation for a link from a Lambda function to an RDS
// Proxy. The function is given network access to the proxy (a security-group rule on the
// database port), optionally an IAM rds-db:connect grant for token-based auth, and connection
// environment variables (endpoint/port/database name).
func FunctionProxyLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
) func(FunctionToProxyLinkDeps) provider.Link {
	return func(
		linkServiceDeps FunctionToProxyLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/function__proxy.md")

		actions := &functionProxyLinkActions{
			lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
			iamServiceFactory:    iamServiceFactory,
			ec2ServiceFactory:    ec2ServiceFactory,
			awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/lambda/function",
			ResourceTypeB: "aws/rds/dbProxy",
			// It doesn't matter which resource is created first; the function is configured
			// to reach the proxy once both have been created.
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceNone,
			PlainTextSummary:                "A link that grants a Lambda function network and auth access to an RDS database via an RDS Proxy.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           functionProxyLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type functionProxyLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
	iamServiceFactory    pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	ec2ServiceFactory    pluginutils.ServiceFactory[*aws.Config, ec2service.Service]
}

func (l *functionProxyLinkActions) getLambdaServiceWithRegion(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, string, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, "", err
	}
	return l.lambdaServiceFactory(awsConfig, providerContext), awsConfig.Region, nil
}

func (l *functionProxyLinkActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.iamServiceFactory(awsConfig, providerContext), nil
}

func (l *functionProxyLinkActions) getEC2Service(
	ctx context.Context,
	providerContext provider.Context,
) (ec2service.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.ec2ServiceFactory(awsConfig, providerContext), nil
}
