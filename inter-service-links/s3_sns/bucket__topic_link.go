package s3sns

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	s3service "github.com/newstack-cloud/bluelink-provider-aws/services/s3/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// BucketToTopicLinkDeps is the core dependency set for a link from an S3 bucket to an SNS
// topic (object event notification). Only the S3 service (resource A) is used; the topic
// policy is managed as a Cloud Control intermediary via the resource service, so the B
// service is the Cloud Control service and is unused.
type BucketToTopicLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	s3service.Service,
	*aws.Config,
	cloudcontrolservice.Service,
]

// BucketTopicLink returns a link implementation that configures an S3 bucket to publish
// object events to an SNS topic. The link grants S3 permission to publish to the topic and
// merges a topic entry into the bucket's notification configuration.
func BucketTopicLink() func(BucketToTopicLinkDeps) provider.Link {
	return func(deps BucketToTopicLinkDeps) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/bucket__topic.md")

		actions := &bucketTopicLinkActions{
			s3ServiceFactory: deps.ResourceAService.ServiceFactory,
			awsConfigStore:   deps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA:                   "aws/s3/bucket",
			ResourceTypeB:                   "aws/sns/topic",
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceA,
			PlainTextSummary:                "A link that configures an S3 bucket to publish object event notifications to an SNS topic.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           bucketTopicLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type bucketTopicLinkActions struct {
	s3ServiceFactory pluginutils.ServiceFactory[*aws.Config, s3service.Service]
	awsConfigStore   pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *bucketTopicLinkActions) getS3Service(
	ctx context.Context,
	providerContext provider.Context,
) (s3service.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.s3ServiceFactory(awsConfig, providerContext), nil
}
