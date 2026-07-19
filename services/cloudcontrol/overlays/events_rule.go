package overlays

import (
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

const eventsRuleType = "aws/events/rule"

func init() {
	Register(eventsRuleType, eventsRuleOverlay)
	RegisterBehaviour(eventsRuleType, &Behaviour{
		// The Cloud Control AWS::Events::Rule handler NPEs when Name is omitted
		// instead of auto-naming the rule, so a name is generated client-side
		// before the create request. Rule names are capped at 64 characters.
		Name: &NameGeneration{
			Field:    "name",
			Generate: utils.DefaultUniqueNameGenerator(64),
		},
	})
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
