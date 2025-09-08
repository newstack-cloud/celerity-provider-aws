package sqsservice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the Amazon Simple Queue Service
// used by the SQS resource implementations.
type Service interface {
	// CreateQueue creates a new queue or returns the URL of an existing one.
	CreateQueue(
		ctx context.Context,
		params *sqs.CreateQueueInput,
		optFns ...func(*sqs.Options),
	) (*sqs.CreateQueueOutput, error)

	// GetQueueUrl returns the URL of an existing queue.
	GetQueueUrl(
		ctx context.Context,
		params *sqs.GetQueueUrlInput,
		optFns ...func(*sqs.Options),
	) (*sqs.GetQueueUrlOutput, error)

	// GetQueueAttributes retrieves attributes for the specified queue.
	GetQueueAttributes(
		ctx context.Context,
		params *sqs.GetQueueAttributesInput,
		optFns ...func(*sqs.Options),
	) (*sqs.GetQueueAttributesOutput, error)

	// SetQueueAttributes sets the value of one or more queue attributes.
	SetQueueAttributes(
		ctx context.Context,
		params *sqs.SetQueueAttributesInput,
		optFns ...func(*sqs.Options),
	) (*sqs.SetQueueAttributesOutput, error)

	// DeleteQueue deletes the queue specified by the queue URL.
	DeleteQueue(
		ctx context.Context,
		params *sqs.DeleteQueueInput,
		optFns ...func(*sqs.Options),
	) (*sqs.DeleteQueueOutput, error)

	// TagQueue adds cost allocation tags to the specified queue.
	TagQueue(
		ctx context.Context,
		params *sqs.TagQueueInput,
		optFns ...func(*sqs.Options),
	) (*sqs.TagQueueOutput, error)

	// UntagQueue removes cost allocation tags from the specified queue.
	UntagQueue(
		ctx context.Context,
		params *sqs.UntagQueueInput,
		optFns ...func(*sqs.Options),
	) (*sqs.UntagQueueOutput, error)

	// ListQueueTags lists all cost allocation tags added to the specified queue.
	ListQueueTags(
		ctx context.Context,
		params *sqs.ListQueueTagsInput,
		optFns ...func(*sqs.Options),
	) (*sqs.ListQueueTagsOutput, error)
}

// NewService creates a new instance of the AWS SQS service
// based on the provided AWS configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return sqs.NewFromConfig(*awsConfig)
}
