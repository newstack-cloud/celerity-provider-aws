package overlays

import "github.com/newstack-cloud/bluelink/libs/blueprint/provider"

const snsSubscriptionType = "aws/sns/subscription"

func init() {
	Register(snsSubscriptionType, snsSubscriptionOverlay)
}

// An SNS subscription wires its destination through the endpoint field. Marking that
// field as a link wiring slot lets a reference to the destination resource (an SQS queue
// or a Lambda function) activate the subscription -> destination link that grants the
// required queue policy / invoke permission, without the author also adding a
// linkSelector.
func snsSubscriptionOverlay(
	schema *provider.ResourceDefinitionsSchema,
) *provider.ResourceDefinitionsSchema {
	if schema.Attributes == nil {
		return schema
	}
	if endpoint := schema.Attributes["endpoint"]; endpoint != nil {
		endpoint.ActivatesLinkOnReference = true
	}
	return schema
}
