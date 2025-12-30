package sqslinks

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func sqsQueueQueueLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/sqs/queue::aws.sqs.redrive.maxReceiveCount": {
			Name:  "aws.sqs.redrive.maxReceiveCount",
			Label: "Max Receive Count",
			Type:  core.ScalarTypeInteger,
			Description: "The number of times a message is delivered to the source queue before being moved to the dead-letter queue. " +
				"Valid values: 1 to 1000. Default is 10.",
			DefaultValue: core.ScalarFromInt(10),
			Required:     false,
			Examples: []*core.ScalarValue{
				core.ScalarFromInt(3),
				core.ScalarFromInt(10),
				core.ScalarFromInt(50),
			},
		},
	}
}
