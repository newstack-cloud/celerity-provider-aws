package testutils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// MockAWSConfigLoader is a mock implementation of the AWSConfigLoader interface
// used for testing.
type MockAWSConfigLoader struct {
	LoadDefaultConfigFunc func(
		ctx context.Context,
		optFns ...func(*config.LoadOptions) error,
	) (aws.Config, error)
}

func (m *MockAWSConfigLoader) LoadDefaultConfig(
	ctx context.Context,
	optFns ...func(*config.LoadOptions) error,
) (aws.Config, error) {
	if m.LoadDefaultConfigFunc != nil {
		return m.LoadDefaultConfigFunc(ctx, optFns...)
	}
	return aws.Config{}, nil
}
