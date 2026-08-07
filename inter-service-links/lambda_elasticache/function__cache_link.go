package lambdaelasticache

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// FunctionToCacheLinkDeps is a type alias for the core dependencies for a link from a Lambda
// function to an ElastiCache replication group. The cache is a Cloud Control–backed resource, so
// resource B's service is the Cloud Control service; the link only reads the cache's endpoints and
// security group from state/spec and does not call it.
type FunctionToCacheLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	cloudcontrolservice.Service,
]

// FunctionCacheLink returns a link implementation for a link from a Lambda function to an
// ElastiCache (Redis/Valkey) replication group. The function is given network access to the cache
// (a security-group rule on the cache port) and connection environment variables (endpoint/port).
func FunctionCacheLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
) func(FunctionToCacheLinkDeps) provider.Link {
	return func(
		linkServiceDeps FunctionToCacheLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/function__cache.md")

		actions := &functionCacheLinkActions{
			lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
			iamServiceFactory:    iamServiceFactory,
			ec2ServiceFactory:    ec2ServiceFactory,
			awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/lambda/function",
			ResourceTypeB: "aws/elasticache/replicationGroup",
			// ReconcileLinkNetworking reads the function's live VPC attachment, so the
			// placement link that establishes it has to run first.
			Requires: linkutils.NetworkAttachedRequired(provider.LinkPriorityResourceA),
			// It doesn't matter which resource is created first; the function is configured
			// to reach the cache once both have been created.
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceNone,
			PlainTextSummary:                "A link that grants a Lambda function network access to an ElastiCache replication group.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           functionCacheLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type functionCacheLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
	iamServiceFactory    pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	ec2ServiceFactory    pluginutils.ServiceFactory[*aws.Config, ec2service.Service]
}

func (l *functionCacheLinkActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.iamServiceFactory(awsConfig, providerContext), nil
}

func (l *functionCacheLinkActions) getLambdaServiceWithRegion(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, string, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, "", err
	}
	return l.lambdaServiceFactory(awsConfig, providerContext), awsConfig.Region, nil
}

func (l *functionCacheLinkActions) getEC2Service(
	ctx context.Context,
	providerContext provider.Context,
) (ec2service.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.ec2ServiceFactory(awsConfig, providerContext), nil
}
