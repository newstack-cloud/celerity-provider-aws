package snslambda

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// SubscriptionFunctionLink returns a link implementation for an SNS subscription (lambda
// protocol) invoking a Lambda function. The subscription wiring and all of its
// configuration are modelled on the aws/sns/subscription resource; this link only grants
// SNS permission to invoke the function, scoped to the subscription's topic. It is
// reference-implied, it activates when the subscription's endpoint references the
// function.
func SubscriptionFunctionLink() provider.Link {
	description, _ := descriptions.ReadFile("descriptions/subscription__function.md")

	actions := &subscriptionFunctionLinkActions{}

	return &providerv1.LinkDefinition{
		ResourceTypeA:                   "aws/sns/subscription",
		ResourceTypeB:                   "aws/lambda/function",
		Kind:                            provider.LinkKindSoft,
		PriorityResource:                provider.LinkPriorityResourceNone,
		PlainTextSummary:                "A link that grants SNS permission to invoke a Lambda function for a subscription delivering to it.",
		FormattedDescription:            string(description),
		StageChangesFunc:                actions.StageChanges,
		UpdateResourceAFunc:             actions.UpdateResourceA,
		UpdateResourceBFunc:             actions.UpdateResourceB,
		UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
	}
}

type subscriptionFunctionLinkActions struct{}
