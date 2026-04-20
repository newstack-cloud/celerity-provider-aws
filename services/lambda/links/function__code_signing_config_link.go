package lambdalinks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// FunctionCodeSigningConfigLink returns a link implementation for
// a link from a lambda function to a code signing config.
func FunctionCodeSigningConfigLink(
	linkServiceDeps pluginutils.LinkServiceDeps[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	],
) provider.Link {
	description, _ := descriptions.ReadFile("descriptions/function__code_signing_config.md")

	actions := &lambdaFunctionCodeSigningConfigLinkActions{
		lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
		awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
	}

	return &providerv1.LinkDefinition{
		ResourceTypeA: "aws/lambda/function",
		ResourceTypeB: "aws/lambda/codeSigningConfig",
		CardinalityA: provider.LinkCardinality{
			// A lambda function can only have one signing config.
			Max: 1,
		},
		Kind:                            provider.LinkKindHard,
		PriorityResource:                provider.LinkPriorityResourceB,
		PlainTextSummary:                "A link from a lambda function to a code signing config.",
		FormattedDescription:            string(description),
		AnnotationDefinitions:           lambdaFunctionCodeSigningConfigLinkAnnotations(),
		StageChangesFunc:                actions.StageChanges,
		UpdateResourceAFunc:             actions.UpdateResourceA,
		UpdateResourceBFunc:             actions.UpdateResourceB,
		UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
	}
}

type lambdaFunctionCodeSigningConfigLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaFunctionCodeSigningConfigLinkActions) getLambdaService(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.lambdaServiceFactory(awsConfig, providerContext), nil
}
