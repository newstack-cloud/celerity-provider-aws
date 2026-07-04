package s3service

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service represents the subset of the Amazon S3 API used by the S3 link
// implementations. Bucket resources themselves are Cloud Control–backed; this service
// covers the bucket-notification read-modify-write that links perform when wiring a
// bucket to a Lambda function, SQS queue or SNS topic.
type Service interface {
	// GetBucketNotificationConfiguration returns the notification configuration of a bucket.
	GetBucketNotificationConfiguration(
		ctx context.Context,
		params *s3.GetBucketNotificationConfigurationInput,
		optFns ...func(*s3.Options),
	) (*s3.GetBucketNotificationConfigurationOutput, error)

	// PutBucketNotificationConfiguration sets the notification configuration of a bucket.
	PutBucketNotificationConfiguration(
		ctx context.Context,
		params *s3.PutBucketNotificationConfigurationInput,
		optFns ...func(*s3.Options),
	) (*s3.PutBucketNotificationConfigurationOutput, error)
}

// NewService creates a new instance of the AWS S3 service based on the provided AWS
// configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return s3.NewFromConfig(*awsConfig)
}
