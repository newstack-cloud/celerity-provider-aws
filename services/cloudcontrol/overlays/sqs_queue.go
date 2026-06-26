package overlays

import (
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

const sqsQueueType = "aws/sqs/queue"

func init() {
	Register(sqsQueueType, sqsQueueOverlay)
	RegisterBehaviour(sqsQueueType, &Behaviour{
		Name: &NameGeneration{
			Field:    "queueName",
			Generate: sqsQueueNameGenerator,
		},
	})
}

// Appends the .fifo suffix when the resolved spec marks the queue as FIFO, mirroring the
// native resource.
func sqsQueueNameGenerator(input *provider.ResourceDeployInput) (string, error) {
	fifo := false
	if spec := resolvedSpec(input); spec != nil {
		fifo = core.BoolValue(spec.Fields["fifoQueue"])
	}
	return utils.SQSQueueNameGenerator(fifo)(input)
}

func resolvedSpec(input *provider.ResourceDeployInput) *core.MappingNode {
	if input == nil || input.Changes == nil {
		return nil
	}
	resolved := input.Changes.AppliedResourceInfo.ResourceWithResolvedSubs
	if resolved == nil {
		return nil
	}
	return resolved.Spec
}

// Refines the generated AWS::SQS::Queue schema: CloudFormation documents the
// numeric bounds only in prose, so they are restored here. RedrivePolicy /
// RedriveAllowPolicy are intentionally left as the generated JSON-string fields
// (free-form objects): the engine parses/serialises them at the Cloud Control
// boundary, and the DLQ link owns each policy in full (a single relationship), so
// it writes the whole JSON value without needing to work with sub-fields directly.
func sqsQueueOverlay(
	schema *provider.ResourceDefinitionsSchema,
) *provider.ResourceDefinitionsSchema {
	if schema.Attributes == nil {
		return schema
	}

	setNumericBounds(schema.Attributes["delaySeconds"], 0, 900)
	setNumericBounds(schema.Attributes["visibilityTimeout"], 0, 43200)
	setNumericBounds(schema.Attributes["maximumMessageSize"], 1024, 262144)
	setNumericBounds(schema.Attributes["messageRetentionPeriod"], 60, 1209600)
	setNumericBounds(schema.Attributes["receiveMessageWaitTimeSeconds"], 0, 20)

	return schema
}

func setNumericBounds(field *provider.ResourceDefinitionsSchema, minimum, maximum int) {
	if field == nil {
		return
	}
	field.Minimum = core.ScalarFromInt(minimum)
	field.Maximum = core.ScalarFromInt(maximum)
}
