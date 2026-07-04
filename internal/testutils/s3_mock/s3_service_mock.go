package s3mock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3service "github.com/newstack-cloud/bluelink-provider-aws/services/s3/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

type s3ServiceMock struct {
	plugintestutils.MockCalls

	getBucketNotificationConfigurationOutput *s3.GetBucketNotificationConfigurationOutput
	getBucketNotificationConfigurationError  error

	putBucketNotificationConfigurationOutput *s3.PutBucketNotificationConfigurationOutput
	putBucketNotificationConfigurationError  error
}

// CreateS3ServiceMock creates a new instance of the S3 service mock with the provided options.
func CreateS3ServiceMock(options ...S3ServiceMockOption) *s3ServiceMock {
	mock := &s3ServiceMock{}
	for _, option := range options {
		option(mock)
	}
	return mock
}

// CreateS3ServiceMockFactory returns a service factory bound to a single mock instance.
func CreateS3ServiceMockFactory(
	options ...S3ServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) s3service.Service {
	mock := CreateS3ServiceMock(options...)
	return func(awsConfig *aws.Config, providerContext provider.Context) s3service.Service {
		return mock
	}
}

// S3ServiceMockOption is a function type for configuring the S3 service mock.
type S3ServiceMockOption func(*s3ServiceMock)

func WithGetBucketNotificationConfigurationOutput(output *s3.GetBucketNotificationConfigurationOutput) S3ServiceMockOption {
	return func(mock *s3ServiceMock) {
		mock.getBucketNotificationConfigurationOutput = output
	}
}

func WithGetBucketNotificationConfigurationError(err error) S3ServiceMockOption {
	return func(mock *s3ServiceMock) {
		mock.getBucketNotificationConfigurationError = err
	}
}

func WithPutBucketNotificationConfigurationOutput(output *s3.PutBucketNotificationConfigurationOutput) S3ServiceMockOption {
	return func(mock *s3ServiceMock) {
		mock.putBucketNotificationConfigurationOutput = output
	}
}

func WithPutBucketNotificationConfigurationError(err error) S3ServiceMockOption {
	return func(mock *s3ServiceMock) {
		mock.putBucketNotificationConfigurationError = err
	}
}

func (m *s3ServiceMock) GetBucketNotificationConfiguration(
	ctx context.Context,
	params *s3.GetBucketNotificationConfigurationInput,
	optFns ...func(*s3.Options),
) (*s3.GetBucketNotificationConfigurationOutput, error) {
	m.RegisterCall(ctx, params)
	if m.getBucketNotificationConfigurationError != nil {
		return nil, m.getBucketNotificationConfigurationError
	}
	return m.getBucketNotificationConfigurationOutput, nil
}

func (m *s3ServiceMock) PutBucketNotificationConfiguration(
	ctx context.Context,
	params *s3.PutBucketNotificationConfigurationInput,
	optFns ...func(*s3.Options),
) (*s3.PutBucketNotificationConfigurationOutput, error) {
	m.RegisterCall(ctx, params)
	if m.putBucketNotificationConfigurationError != nil {
		return nil, m.putBucketNotificationConfigurationError
	}
	return m.putBucketNotificationConfigurationOutput, nil
}
