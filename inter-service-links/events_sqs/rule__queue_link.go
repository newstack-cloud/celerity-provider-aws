package eventssqs

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// RuleQueueLink returns a link implementation for a link from an EventBridge rule
// to an SQS queue. The rule -> queue wiring itself is modelled inline in the rule's
// targets list. This link only grants the queue's resource-based policy permission
// allowing EventBridge to send messages to it.
func RuleQueueLink() provider.Link {
	description, _ := descriptions.ReadFile("descriptions/rule__queue.md")

	actions := &ruleQueueLinkActions{}

	return &providerv1.LinkDefinition{
		ResourceTypeA:                   "aws/events/rule",
		ResourceTypeB:                   "aws/sqs/queue",
		Kind:                            provider.LinkKindSoft,
		PriorityResource:                provider.LinkPriorityResourceNone,
		PlainTextSummary:                "A link that grants an EventBridge rule permission to send messages to an SQS queue.",
		FormattedDescription:            string(description),
		StageChangesFunc:                actions.StageChanges,
		UpdateResourceAFunc:             actions.UpdateResourceA,
		UpdateResourceBFunc:             actions.UpdateResourceB,
		UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
	}
}

type ruleQueueLinkActions struct{}
