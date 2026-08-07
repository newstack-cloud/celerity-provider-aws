package flexlambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// VPCToFunctionLinkDeps is a type alias for the core dependencies for a link from a
// flex VPC to a Lambda function. Resource A is the flex VPC (backed by the EC2 service)
// and resource B is the Lambda function.
type VPCToFunctionLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	ec2service.Service,
	*aws.Config,
	lambdaservice.Service,
]

// VPCFunctionLink returns a link implementation that places a Lambda function within a
// flex VPC. The function's VpcConfig is set from the VPC's computed subnets (selected by
// the aws.flexvpc.lambda.subnetType annotation) and Bluelink-managed security group.
//
// The function's execution role is granted the network interface permissions Lambda
// requires of a VPC-attached function, so the IAM service is needed alongside the two
// resource services.
func VPCFunctionLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
) func(VPCToFunctionLinkDeps) provider.Link {
	return func(
		linkServiceDeps VPCToFunctionLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/vpc__function.md")

		actions := &vpcFunctionLinkActions{
			lambdaServiceFactory: linkServiceDeps.ResourceBService.ServiceFactory,
			awsConfigStore:       linkServiceDeps.ResourceBService.ConfigStore,
			ec2ServiceFactory:    linkServiceDeps.ResourceAService.ServiceFactory,
			ec2ConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
			iamServiceFactory:    iamServiceFactory,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/flex/vpc",
			ResourceTypeB: "aws/lambda/function",
			// The VPC must exist (and have stabilised its subnets and security group)
			// before the function can be placed into it.
			Kind:             provider.LinkKindSoft,
			PriorityResource: provider.LinkPriorityResourceA,
			// Every link that reads the function's live VPC attachment waits for this
			// one, and is torn down before it.
			Provides:                        linkutils.NetworkAttachedProvided(provider.LinkPriorityResourceB),
			PlainTextSummary:                "A link that places a Lambda function within a flex VPC's subnets.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           vpcFunctionLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type vpcFunctionLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
	// The link mints a security group for each function it places, so it needs the
	// EC2 service that backs resource A as well as the Lambda service for resource B.
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service]
	ec2ConfigStore    pluginutils.ServiceConfigStore[*aws.Config]
	// Attaching a function to a VPC requires its execution role to be able to manage
	// network interfaces, which this link grants before it sets the attachment.
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
}

func (l *vpcFunctionLinkActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}

	return l.iamServiceFactory(awsConfig, providerContext), nil
}

func (l *vpcFunctionLinkActions) getEC2Service(
	ctx context.Context,
	providerContext provider.Context,
) (ec2service.Service, error) {
	awsConfig, err := l.ec2ConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}

	return l.ec2ServiceFactory(awsConfig, providerContext), nil
}

func (l *vpcFunctionLinkActions) getLambdaService(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}

	return l.lambdaServiceFactory(awsConfig, providerContext), nil
}
