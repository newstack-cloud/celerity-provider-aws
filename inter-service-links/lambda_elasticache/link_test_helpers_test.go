//go:build unit

package lambdaelasticache

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

const (
	lcFunctionARN  = "arn:aws:lambda:us-west-2:123456789012:function:api"
	lcRoleARN      = "arn:aws:iam::123456789012:role/api-role"
	lcRoleName     = "api-role"
	lcRoleResource = "apiFunctionRole"
	lcExecRole     = "apiFunctionExecutionRole"
	lcConnectSID   = "ElastiCacheConnectsessionCache"
	lcPrefix       = "SESSION_CACHE"

	testCacheEndpoint       = "sessions.abc.use1.cache.amazonaws.com"
	testCacheConfigEndpoint = "sessions-cfg.abc.use1.cache.amazonaws.com"
	testCacheRGId           = "sessions"
	testCacheSGID           = "sg-cache"
)

func lcRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: lcRoleResource,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(lcRoleName),
			"arn", core.MappingNodeFromString(lcRoleARN),
		),
	}
}

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

func functionCacheLinkFactory(
	iamSvc iamservice.Service,
	ec2Factory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
) func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
) provider.Link {
	build := FunctionCacheLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iamSvc },
		ec2Factory,
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
	) provider.Link {
		return build(FunctionToCacheLinkDeps(deps))
	}
}

func noopEC2ServiceFactory() pluginutils.ServiceFactory[*aws.Config, ec2service.Service] {
	return ec2mock.CreateEc2ServiceMockFactory()
}

func noopCloudControlServiceFactory(_ *aws.Config, _ provider.Context) cloudcontrolservice.Service {
	return nil
}

func lcFunctionInfo(annotations map[string]*core.MappingNode) *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "apiFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("arn", core.MappingNodeFromString(lcFunctionARN)),
		},
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{Fields: annotations},
			},
		},
	}
}

// Builds a cache resource fixture: endpoints are readable in external state, while
// securityGroupIds is write-only and only present in the resolved blueprint spec.
func lcCacheInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "sessionCache",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"replicationGroupId", core.MappingNodeFromString(testCacheRGId),
				"primaryEndPoint", core.MappingNodeFields(
					"address", core.MappingNodeFromString(testCacheEndpoint),
					"port", core.MappingNodeFromString("6379"),
				),
			),
		},
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Spec: core.MappingNodeFields(
				"securityGroupIds", &core.MappingNode{Items: []*core.MappingNode{
					core.MappingNodeFromString(testCacheSGID),
				}},
			),
		},
	}
}

func lcGetFunctionOutput(vars map[string]string, vpcConfig *lambdatypes.VpcConfigResponse) *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &lambdatypes.FunctionConfiguration{
			FunctionArn: aws.String(lcFunctionARN),
			Role:        aws.String(lcRoleARN),
			Environment: &lambdatypes.EnvironmentResponse{Variables: vars},
			VpcConfig:   vpcConfig,
		},
	}
}
