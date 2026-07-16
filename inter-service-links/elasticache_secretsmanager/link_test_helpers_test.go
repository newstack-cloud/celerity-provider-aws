//go:build unit

package elasticachesecretsmanager

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	elasticacheservice "github.com/newstack-cloud/bluelink-provider-aws/services/elasticache/service"
	secretsmanagerservice "github.com/newstack-cloud/bluelink-provider-aws/services/secretsmanager/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

const (
	rgReplicationGroupID = "sessions"
	rgSecretARN          = "arn:aws:secretsmanager:us-west-2:123456789012:secret:redis-auth-AbCdEf"
	rgAuthTokenValue     = "super-secret-redis-auth-token-value"
)

func testLinkContext() provider.LinkContext {
	return plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {"region": core.ScalarFromString("us-west-2")},
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)
}

func testConfigStore(loader *testutils.MockAWSConfigLoader) pluginutils.ServiceConfigStore[*aws.Config] {
	return utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
}

// replicationGroupSecretLinkFactory binds the hand-written ElastiCache and Secrets Manager
// service mocks into the link constructor, so the test harness can build the link with
// Cloud Control resource services (which this link never calls).
func replicationGroupSecretLinkFactory(
	elastiCacheSvc elasticacheservice.Service,
	secretsManagerSvc secretsmanagerservice.Service,
) func(
	pluginutils.LinkServiceDeps[*aws.Config, cloudcontrolservice.Service, *aws.Config, cloudcontrolservice.Service],
) provider.Link {
	build := ReplicationGroupSecretLink(
		func(c *aws.Config, pc provider.Context) elasticacheservice.Service { return elastiCacheSvc },
		func(c *aws.Config, pc provider.Context) secretsmanagerservice.Service { return secretsManagerSvc },
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, cloudcontrolservice.Service, *aws.Config, cloudcontrolservice.Service],
	) provider.Link {
		return build(ReplicationGroupToSecretLinkDeps(deps))
	}
}

func noopCloudControlServiceFactory(_ *aws.Config, _ provider.Context) cloudcontrolservice.Service {
	return nil
}

// replicationGroupInfo builds the replication group (resource A) fixture with its computed
// replicationGroupId and optional annotations.
func replicationGroupInfo(annotations map[string]*core.MappingNode) *provider.ResourceInfo {
	info := &provider.ResourceInfo{
		ResourceName: "sessionCache",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"replicationGroupId", core.MappingNodeFromString(rgReplicationGroupID),
				"transitEncryptionEnabled", core.MappingNodeFromBool(true),
			),
		},
	}
	if annotations != nil {
		info.ResourceWithResolvedSubs = &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{Fields: annotations},
			},
		}
	}
	return info
}

// secretInfo builds the Secrets Manager secret (resource B) fixture. The secret's primary
// identifier ("id") is its ARN.
func secretInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "redisAuthSecret",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"id", core.MappingNodeFromString(rgSecretARN),
			),
		},
	}
}
