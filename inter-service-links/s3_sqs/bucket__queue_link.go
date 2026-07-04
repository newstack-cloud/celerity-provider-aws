package s3sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3service "github.com/newstack-cloud/bluelink-provider-aws/services/s3/service"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// BucketToQueueLinkDeps is the core dependency set for a link from an S3 bucket to an SQS
// queue (object event notification). Only the S3 service (resource A) is used; the queue
// policy is managed as a Cloud Control intermediary via the resource service.
type BucketToQueueLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	s3service.Service,
	*aws.Config,
	sqsservice.Service,
]

// BucketQueueLink returns a link implementation that configures an S3 bucket to deliver
// object event notifications to an SQS queue. The link grants S3 permission to send
// messages to the queue and merges a queue entry into the bucket's notification
// configuration.
func BucketQueueLink() func(BucketToQueueLinkDeps) provider.Link {
	return func(deps BucketToQueueLinkDeps) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/bucket__queue.md")

		actions := &bucketQueueLinkActions{
			s3ServiceFactory: deps.ResourceAService.ServiceFactory,
			awsConfigStore:   deps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA:                   "aws/s3/bucket",
			ResourceTypeB:                   "aws/sqs/queue",
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceA,
			PlainTextSummary:                "A link that configures an S3 bucket to deliver object event notifications to an SQS queue.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           bucketQueueLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateResourceAFunc:             actions.UpdateResourceA,
			UpdateResourceBFunc:             actions.UpdateResourceB,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type bucketQueueLinkActions struct {
	s3ServiceFactory pluginutils.ServiceFactory[*aws.Config, s3service.Service]
	awsConfigStore   pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *bucketQueueLinkActions) getS3Service(
	ctx context.Context,
	providerContext provider.Context,
) (s3service.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.s3ServiceFactory(awsConfig, providerContext), nil
}
