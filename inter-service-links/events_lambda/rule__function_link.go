package eventslambda

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// RuleFunctionLink returns a link implementation for a link from an EventBridge
// rule to a Lambda function. The rule -> function wiring itself is modelled inline
// in the rule's targets array. This link only grants the function's resource-based
// permission allowing EventBridge to invoke it.
func RuleFunctionLink() provider.Link {
	description, _ := descriptions.ReadFile("descriptions/rule__function.md")

	actions := &ruleFunctionLinkActions{}

	return &providerv1.LinkDefinition{
		ResourceTypeA:                   "aws/events/rule",
		ResourceTypeB:                   "aws/lambda/function",
		Kind:                            provider.LinkKindSoft,
		PriorityResource:                provider.LinkPriorityResourceNone,
		PlainTextSummary:                "A link that grants an EventBridge rule permission to invoke a Lambda function.",
		FormattedDescription:            string(description),
		StageChangesFunc:                actions.StageChanges,
		UpdateResourceAFunc:             actions.UpdateResourceA,
		UpdateResourceBFunc:             actions.UpdateResourceB,
		UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
	}
}

type ruleFunctionLinkActions struct{}
