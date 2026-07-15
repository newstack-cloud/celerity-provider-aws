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

// LoadDefaultConfig delegates to LoadDefaultConfigFunc when set. Otherwise it resolves
// the given load options in memory the same way the real loader does (e.g. region from
// config.WithRegion), without touching the filesystem, environment or network.
func (m *MockAWSConfigLoader) LoadDefaultConfig(
	ctx context.Context,
	optFns ...func(*config.LoadOptions) error,
) (aws.Config, error) {
	if m.LoadDefaultConfigFunc != nil {
		return m.LoadDefaultConfigFunc(ctx, optFns...)
	}

	opts := &config.LoadOptions{}
	for _, fn := range optFns {
		if err := fn(opts); err != nil {
			return aws.Config{}, err
		}
	}
	return aws.Config{Region: opts.Region}, nil
}
