package s3sns

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func bucketTopicLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/sns/topic::aws.s3.sns.event.<index>": {
			Name:  "aws.s3.sns.event.<index>",
			Label: "Notification Event",
			Type:  core.ScalarTypeString,
			Description: "An S3 event type that publishes to the topic. `<index>` is the zero-based " +
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
		"aws/sns/topic::aws.s3.sns.filterPrefix": {
			Name:        "aws.s3.sns.filterPrefix",
			Label:       "Filter Prefix",
			Type:        core.ScalarTypeString,
			Description: "Restricts notifications to objects whose key begins with this prefix.",
			Required:    false,
		},
		"aws/sns/topic::aws.s3.sns.filterSuffix": {
			Name:        "aws.s3.sns.filterSuffix",
			Label:       "Filter Suffix",
			Type:        core.ScalarTypeString,
			Description: "Restricts notifications to objects whose key ends with this suffix.",
			Required:    false,
		},
	}
}
