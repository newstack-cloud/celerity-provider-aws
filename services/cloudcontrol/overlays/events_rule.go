package overlays

import "github.com/newstack-cloud/bluelink/libs/blueprint/provider"

const eventsRuleType = "aws/events/rule"

func init() {
	Register(eventsRuleType, eventsRuleOverlay)
}

// An EventBridge rule wires each target inline through targets[].arn. Marking that
// field as a link wiring slot lets a reference to a target resource (Lambda, SQS,
// API destination) activate the rule -> target link that grants the invoke
// permission, without the author also having to add a linkSelector.
func eventsRuleOverlay(
	schema *provider.ResourceDefinitionsSchema,
) *provider.ResourceDefinitionsSchema {
	if targetArn := targetArnSchema(schema); targetArn != nil {
		targetArn.ActivatesLinkOnReference = true
	}
	return schema
}

func targetArnSchema(
	schema *provider.ResourceDefinitionsSchema,
) *provider.ResourceDefinitionsSchema {
	if schema.Attributes == nil {
		return nil
	}
	targets := schema.Attributes["targets"]
	if targets == nil || targets.Items == nil || targets.Items.Attributes == nil {
		return nil
	}
	return targets.Items.Attributes["arn"]
}
