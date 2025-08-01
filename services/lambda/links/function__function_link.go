package lambdalinks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// FunctionToFunctionLinkDeps is a type alias for the core
// dependencies for a link from a lambda function to another lambda function.
type FunctionToFunctionLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
]

// FunctionFunctionLink returns a link implementation for
// a link from a lambda function to another lambda function.
// The first lambda function will be configured with permissions
// and environment variables to be able to invoke the second lambda function.
func FunctionFunctionLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
) func(FunctionToFunctionLinkDeps) provider.Link {
	return func(
		linkServiceDeps FunctionToFunctionLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/function__function.md")

		actions := &lambdaFunctionFunctionLinkActions{
			lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
			iamServiceFactory:    iamServiceFactory,
			awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
			ec2ServiceFactory:    ec2ServiceFactory,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/lambda/function",
			ResourceTypeB: "aws/lambda/function",
			// It doesn't matter which lambda function is created first,
			// the caller function will be configured to be able to invoke
			// the callee function once both have been created.
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceNone,
			PlainTextSummary:                "A link that configures a lambda function to be able to invoke another lambda function.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           lambdaFunctionFunctionLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type lambdaFunctionFunctionLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
	iamServiceFactory    pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	ec2ServiceFactory    pluginutils.ServiceFactory[*aws.Config, ec2service.Service]
}

func (l *lambdaFunctionFunctionLinkActions) getLambdaServiceWithRegion(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, string, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, "", err
	}

	return l.lambdaServiceFactory(awsConfig, providerContext), awsConfig.Region, nil
}

func (l *lambdaFunctionFunctionLinkActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.iamServiceFactory(awsConfig, providerContext), nil
}

func (l *lambdaFunctionFunctionLinkActions) getEC2Service(
	ctx context.Context,
	providerContext provider.Context,
) (ec2service.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.ec2ServiceFactory(awsConfig, providerContext), nil
}
