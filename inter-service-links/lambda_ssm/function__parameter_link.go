package lambdassm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// FunctionToParameterLinkDeps is a type alias for the core dependencies for a link from a
// Lambda function to an SSM parameter. The parameter is hand-written (SDK-backed), so
// resource B's service is the SSM service; the link only reads the parameter name and ARN
// from state and does not call it.
type FunctionToParameterLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	ssmservice.Service,
]

// FunctionParameterLink returns a link implementation for a link from a Lambda function to
// an SSM parameter. The Lambda function is granted access to the parameter and, optionally,
// an environment variable referencing the parameter name. When the function is VPC-isolated,
// an SSM interface VPC endpoint is activated on the linked flex VPC.
func FunctionParameterLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
) func(FunctionToParameterLinkDeps) provider.Link {
	return func(
		linkServiceDeps FunctionToParameterLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/function__parameter.md")

		actions := &functionParameterLinkActions{
			lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
			iamServiceFactory:    iamServiceFactory,
			ec2ServiceFactory:    ec2ServiceFactory,
			awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/lambda/function",
			ResourceTypeB: "aws/ssm/parameter",
			// ReconcileLinkNetworking reads the function's live VPC attachment, so the
			// placement link that establishes it has to run first.
			Requires: linkutils.NetworkAttachedRequired(provider.LinkPriorityResourceA),
			// It doesn't matter which resource is created first; the lambda function will
			// be configured to access the parameter once both have been created.
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceNone,
			PlainTextSummary:                "A link that grants a Lambda function access to an SSM parameter.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           functionParameterLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type functionParameterLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
	iamServiceFactory    pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	ec2ServiceFactory    pluginutils.ServiceFactory[*aws.Config, ec2service.Service]
}

func (l *functionParameterLinkActions) getLambdaServiceWithRegion(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, string, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, "", err
	}
	return l.lambdaServiceFactory(awsConfig, providerContext), awsConfig.Region, nil
}

func (l *functionParameterLinkActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.iamServiceFactory(awsConfig, providerContext), nil
}

func (l *functionParameterLinkActions) getEC2Service(
	ctx context.Context,
	providerContext provider.Context,
) (ec2service.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.ec2ServiceFactory(awsConfig, providerContext), nil
}
