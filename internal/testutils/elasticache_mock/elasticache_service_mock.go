package elasticachemock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticacheservice "github.com/newstack-cloud/bluelink-provider-aws/services/elasticache/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

type elasticacheServiceMock struct {
	plugintestutils.MockCalls

	modifyReplicationGroupOutput *elasticache.ModifyReplicationGroupOutput
	modifyReplicationGroupError  error
}

// CreateElastiCacheServiceMock creates a new instance of the ElastiCache service mock
// with the provided options.
func CreateElastiCacheServiceMock(options ...ElastiCacheServiceMockOption) *elasticacheServiceMock {
	mock := &elasticacheServiceMock{}
	for _, option := range options {
		option(mock)
	}
	return mock
}

// CreateElastiCacheServiceMockFactory returns a service factory bound to a single mock instance.
func CreateElastiCacheServiceMockFactory(
	options ...ElastiCacheServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) elasticacheservice.Service {
	mock := CreateElastiCacheServiceMock(options...)
	return func(awsConfig *aws.Config, providerContext provider.Context) elasticacheservice.Service {
		return mock
	}
}

// ElastiCacheServiceMockOption is a function type for configuring the ElastiCache service mock.
type ElastiCacheServiceMockOption func(*elasticacheServiceMock)

func WithModifyReplicationGroupOutput(output *elasticache.ModifyReplicationGroupOutput) ElastiCacheServiceMockOption {
	return func(mock *elasticacheServiceMock) { mock.modifyReplicationGroupOutput = output }
}

func WithModifyReplicationGroupError(err error) ElastiCacheServiceMockOption {
	return func(mock *elasticacheServiceMock) { mock.modifyReplicationGroupError = err }
}

func (m *elasticacheServiceMock) ModifyReplicationGroup(
	ctx context.Context,
	params *elasticache.ModifyReplicationGroupInput,
	optFns ...func(*elasticache.Options),
) (*elasticache.ModifyReplicationGroupOutput, error) {
	m.RegisterCall(ctx, params)
	if m.modifyReplicationGroupError != nil {
		return nil, m.modifyReplicationGroupError
	}
	return m.modifyReplicationGroupOutput, nil
}
