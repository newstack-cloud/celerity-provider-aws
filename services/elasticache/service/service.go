package elasticacheservice

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the AWS ElastiCache
// service used by the hand-written ElastiCache link implementations. ElastiCache
// replication groups themselves are Cloud Control–backed; this interface covers the
// out-of-band modifications (such as applying a Redis AUTH token) that Cloud Control
// cannot express because authToken is a write-only field with no computed output.
type Service interface {
	// ModifyReplicationGroup modifies the settings of an existing replication group,
	// including applying or rotating a Redis AUTH token via AuthToken and
	// AuthTokenUpdateStrategy.
	ModifyReplicationGroup(ctx context.Context, params *elasticache.ModifyReplicationGroupInput, optFns ...func(*elasticache.Options)) (*elasticache.ModifyReplicationGroupOutput, error)
}

// NewService creates a new instance of the AWS ElastiCache service
// based on the provided AWS configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return elasticache.NewFromConfig(
		*awsConfig,
		elasticache.WithEndpointResolverV2(
			&elasticacheEndpointResolverV2{
				providerContext,
			},
		),
	)
}

type elasticacheEndpointResolverV2 struct {
	providerContext provider.Context
}

func (i *elasticacheEndpointResolverV2) ResolveEndpoint(
	ctx context.Context,
	params elasticache.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	elasticacheAliases := utils.Services["elasticache"]
	elasticacheEndpoint, hasElastiCacheEndpoint := utils.GetEndpointFromProviderConfig(
		i.providerContext,
		"elasticache",
		elasticacheAliases,
	)
	if hasElastiCacheEndpoint && !core.IsScalarNil(elasticacheEndpoint) {
		u, err := url.Parse(core.StringValueFromScalar(elasticacheEndpoint))
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	return elasticache.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
