//go:build unit

package lambdasqs

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
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
	lqFunctionARN  = "arn:aws:lambda:us-west-2:123456789012:function:submit-order"
	lqRoleARN      = "arn:aws:iam::123456789012:role/submit-order-role"
	lqRoleName     = "submit-order-role"
	lqRoleResource = "submitOrderRole"
	lqQueueURL     = "https://sqs.us-west-2.amazonaws.com/123456789012/orders"
	lqQueueARN     = "arn:aws:sqs:us-west-2:123456789012:orders"
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

func functionQueueLinkFactory(iamSvc iamservice.Service) func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
) provider.Link {
	build := FunctionQueueLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iamSvc },
		ec2mock.CreateEc2ServiceMockFactory(),
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
	) provider.Link {
		return build(FunctionToQueueLinkDeps(deps))
	}
}

func noopCloudControlServiceFactory(_ *aws.Config, _ provider.Context) cloudcontrolservice.Service {
	return nil
}

func lqFunctionInfo(annotations map[string]*core.MappingNode) *provider.ResourceInfo {
	info := &provider.ResourceInfo{
		ResourceName: "submitOrderFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("arn", core.MappingNodeFromString(lqFunctionARN)),
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

func lqQueueInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersQueue",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"queueUrl", core.MappingNodeFromString(lqQueueURL),
				"arn", core.MappingNodeFromString(lqQueueARN),
			),
		},
	}
}

func lqRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: lqRoleResource,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(lqRoleName),
			"arn", core.MappingNodeFromString(lqRoleARN),
		),
	}
}
