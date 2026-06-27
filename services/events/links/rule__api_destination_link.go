package eventslinks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	eventsservice "github.com/newstack-cloud/bluelink-provider-aws/services/events/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// RuleToAPIDestinationLinkDeps is a type alias for the core dependencies for a
// link from an EventBridge rule to an EventBridge API destination.
type RuleToAPIDestinationLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	eventsservice.Service,
	*aws.Config,
	eventsservice.Service,
]

// RuleAPIDestinationLink returns a link implementation for a link from an
// EventBridge rule to an EventBridge API destination. It "activates" the existing
// IAM role referenced by the rule's matching target entry (targets[].roleArn) by
// attaching an inline policy granting events:InvokeApiDestination on the API
// destination's ARN.
func RuleAPIDestinationLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
) func(RuleToAPIDestinationLinkDeps) provider.Link {
	return func(linkServiceDeps RuleToAPIDestinationLinkDeps) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/rule__api_destination.md")

		actions := &ruleAPIDestinationLinkActions{
			iamServiceFactory: iamServiceFactory,
			awsConfigStore:    linkServiceDeps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA:    "aws/events/rule",
			ResourceTypeB:    "aws/events/apiDestination",
			Kind:             provider.LinkKindSoft,
			PriorityResource: provider.LinkPriorityResourceNone,
			PlainTextSummary: "A link that grants an EventBridge rule permission to " +
				"invoke an API destination by populating the target's IAM role.",
			FormattedDescription:            string(description),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type ruleAPIDestinationLinkActions struct {
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	awsConfigStore    pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *ruleAPIDestinationLinkActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}

	return l.iamServiceFactory(awsConfig, providerContext), nil
}
