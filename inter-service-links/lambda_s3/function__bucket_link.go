package lambdas3

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

// FunctionToBucketLinkDeps is a type alias for the core dependencies for a link from a
// Lambda function to an S3 bucket. The bucket is a Cloud Control–backed resource, so
// resource B's service is the Cloud Control service; the link only reads the bucket name
// from state and does not call it.
type FunctionToBucketLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	cloudcontrolservice.Service,
]

// FunctionBucketLink returns a link implementation for a link from a Lambda function to
// an S3 bucket. The Lambda function is granted read/write access to the bucket and,
// optionally, an environment variable referencing the bucket name. When the function is
// VPC-isolated, an S3 gateway VPC endpoint is activated on the linked flex VPC.
func FunctionBucketLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
) func(FunctionToBucketLinkDeps) provider.Link {
	return func(
		linkServiceDeps FunctionToBucketLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/function__bucket.md")

		actions := &functionBucketLinkActions{
			lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
			iamServiceFactory:    iamServiceFactory,
			ec2ServiceFactory:    ec2ServiceFactory,
			awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/lambda/function",
			ResourceTypeB: "aws/s3/bucket",
			// ReconcileLinkNetworking reads the function's live VPC attachment, so the
			// placement link that establishes it has to run first.
			Requires: linkutils.NetworkAttachedRequired(provider.LinkPriorityResourceA),
			// It doesn't matter which resource is created first; the lambda function
			// will be configured to access the S3 bucket once both have been created.
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceNone,
			PlainTextSummary:                "A link that grants a Lambda function read/write access to an S3 bucket.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           functionBucketLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type functionBucketLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
	iamServiceFactory    pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	ec2ServiceFactory    pluginutils.ServiceFactory[*aws.Config, ec2service.Service]
}

func (l *functionBucketLinkActions) getLambdaServiceWithRegion(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, string, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, "", err
	}
	return l.lambdaServiceFactory(awsConfig, providerContext), awsConfig.Region, nil
}

func (l *functionBucketLinkActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.iamServiceFactory(awsConfig, providerContext), nil
}

func (l *functionBucketLinkActions) getEC2Service(
	ctx context.Context,
	providerContext provider.Context,
) (ec2service.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.ec2ServiceFactory(awsConfig, providerContext), nil
}
