package lambdasqs

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

// FunctionToQueueLinkDeps is a type alias for the core dependencies for a link from a Lambda
// function to an SQS queue. The queue is a Cloud Control–backed resource, so resource B's service
// is the Cloud Control service; the link only reads the queue URL and ARN from state.
type FunctionToQueueLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	cloudcontrolservice.Service,
]

// FunctionQueueLink returns a link implementation for a link from a Lambda function to an SQS
// queue. The Lambda function is granted permission to send to (and optionally receive from) the
// queue and, optionally, an environment variable referencing the queue URL. This is the producer
// side of the queue; the consumer side (an SQS event source mapping) is the aws/sqs/queue ->
// aws/lambda/function link.
func FunctionQueueLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
) func(FunctionToQueueLinkDeps) provider.Link {
	return func(
		linkServiceDeps FunctionToQueueLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/function__queue.md")

		actions := &functionQueueLinkActions{
			lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
			iamServiceFactory:    iamServiceFactory,
			ec2ServiceFactory:    ec2ServiceFactory,
			awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/lambda/function",
			ResourceTypeB: "aws/sqs/queue",
			// ReconcileLinkNetworking reads the function's live VPC attachment, so the
			// placement link that establishes it has to run first.
			Requires:                        linkutils.NetworkAttachedRequired(provider.LinkPriorityResourceA),
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceNone,
			PlainTextSummary:                "A link that grants a Lambda function permission to send messages to an SQS queue.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           functionQueueLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type functionQueueLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
	iamServiceFactory    pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	ec2ServiceFactory    pluginutils.ServiceFactory[*aws.Config, ec2service.Service]
}

func (l *functionQueueLinkActions) getLambdaServiceWithRegion(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, string, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, "", err
	}
	return l.lambdaServiceFactory(awsConfig, providerContext), awsConfig.Region, nil
}

func (l *functionQueueLinkActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.iamServiceFactory(awsConfig, providerContext), nil
}

func (l *functionQueueLinkActions) getEC2Service(
	ctx context.Context,
	providerContext provider.Context,
) (ec2service.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.ec2ServiceFactory(awsConfig, providerContext), nil
}
