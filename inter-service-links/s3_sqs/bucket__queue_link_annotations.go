package s3sqs

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func bucketQueueLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/sqs/queue::aws.s3.sqs.event.<index>": {
			Name:  "aws.s3.sqs.event.<index>",
			Label: "Notification Event",
			Type:  core.ScalarTypeString,
			Description: "An S3 event type that delivers a message to the queue. `<index>` is the zero-based " +
				"position in the event list; add multiple events with successive indices (0, 1, ...). " +
				"Defaults to `s3:ObjectCreated:*` when none are specified.",
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("s3:ObjectCreated:*"),
				core.ScalarFromString("s3:ObjectCreated:Put"),
				core.ScalarFromString("s3:ObjectCreated:Post"),
				core.ScalarFromString("s3:ObjectCreated:Copy"),
				core.ScalarFromString("s3:ObjectCreated:CompleteMultipartUpload"),
				core.ScalarFromString("s3:ObjectRemoved:*"),
				core.ScalarFromString("s3:ObjectRemoved:Delete"),
				core.ScalarFromString("s3:ObjectRemoved:DeleteMarkerCreated"),
				core.ScalarFromString("s3:ObjectRestore:*"),
				core.ScalarFromString("s3:ObjectRestore:Post"),
				core.ScalarFromString("s3:ObjectRestore:Completed"),
				core.ScalarFromString("s3:ReducedRedundancyLostObject"),
				core.ScalarFromString("s3:Replication:*"),
			},
			Required: false,
		},
		"aws/sqs/queue::aws.s3.sqs.filterPrefix": {
			Name:        "aws.s3.sqs.filterPrefix",
			Label:       "Filter Prefix",
			Type:        core.ScalarTypeString,
			Description: "Restricts notifications to objects whose key begins with this prefix.",
			Required:    false,
		},
		"aws/sqs/queue::aws.s3.sqs.filterSuffix": {
			Name:        "aws.s3.sqs.filterSuffix",
			Label:       "Filter Suffix",
			Type:        core.ScalarTypeString,
			Description: "Restricts notifications to objects whose key ends with this suffix.",
			Required:    false,
		},
	}
}
