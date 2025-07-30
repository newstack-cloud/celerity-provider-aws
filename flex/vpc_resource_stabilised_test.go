package flex

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FlexVPCResourceStabilisedSuite struct {
	suite.Suite
}

func (s *FlexVPCResourceStabilisedSuite) Test_stabilised() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service]{
		createVPCStabilisedTestCase(providerCtx, loader),
		createVPCNotStabilisedTestCase(providerCtx, loader),
		createVPCWithNATGatewaysStabilisedTestCase(providerCtx, loader),
		createVPCWithNATGatewaysNotStabilisedTestCase(providerCtx, loader),
		createVPCWithNATGatewayInPendingStateTestCase(providerCtx, loader),
		createVPCWithNATGatewayInFailedStateTestCase(providerCtx, loader),
		createVPCWithMultipleNATGatewaysTestCase(providerCtx, loader),
		createVPCWithMissingVPCIDTestCase(providerCtx, loader),
		createVPCWithDescribeVpcsErrorTestCase(providerCtx, loader),
		createVPCWithDescribeNatGatewaysErrorTestCase(providerCtx, loader),
		createVPCWithNoNATGatewaysTestCase(providerCtx, loader),
		createVPCWithEmptyNATGatewaysArrayTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	vpcResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		// Create a mock resource group tagging service factory for testing
		mockResourceGroupTaggingServiceFactory := func(config *aws.Config, ctx provider.Context) resgrouptagservice.Service {
			return nil // Mock service for testing
		}
		return VPCResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		vpcResourceWrapper,
		&s.Suite,
	)
}

func createVPCStabilisedTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
			{
				Vpcs: []types.Vpc{
					{
						VpcId: aws.String("vpc-12345678"),
						State: types.VpcStateAvailable,
					},
				},
			},
		}),
	)

	resourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service]{
		Name: "VPC is stabilised when available",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpec,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
		ExpectError: false,
	}
}

func createVPCNotStabilisedTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
			{
				Vpcs: []types.Vpc{
					{
						VpcId: aws.String("vpc-12345678"),
						State: types.VpcStatePending,
					},
				},
			},
		}),
	)

	resourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service]{
		Name: "VPC is not stabilised when in pending state",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpec,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: false,
		},
		ExpectError: false,
	}
}

func createVPCWithNATGatewaysStabilisedTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
			{
				Vpcs: []types.Vpc{
					{
						VpcId: aws.String("vpc-12345678"),
						State: types.VpcStateAvailable,
					},
				},
			},
		}),
		ec2mock.WithDescribeNatGatewaysOutput(&ec2.DescribeNatGatewaysOutput{
			NatGateways: []types.NatGateway{
				{
					NatGatewayId: aws.String("nat-12345678"),
					State:        types.NatGatewayStateAvailable,
				},
			},
		}),
	)

	resourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"gateways": core.MappingNodeFields(
				"natGateways", core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("nat-12345678"),
					),
				),
			),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service]{
		Name: "VPC with NAT gateways is stabilised when both VPC and NAT gateway are available",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpec,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
		ExpectError: false,
	}
}

func createVPCWithNATGatewaysNotStabilisedTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
			{
				Vpcs: []types.Vpc{
					{
						VpcId: aws.String("vpc-12345678"),
						State: types.VpcStateAvailable,
					},
				},
			},
		}),
		ec2mock.WithDescribeNatGatewaysOutput(&ec2.DescribeNatGatewaysOutput{
			NatGateways: []types.NatGateway{
				{
					NatGatewayId: aws.String("nat-12345678"),
					State:        types.NatGatewayStatePending,
				},
			},
		}),
	)

	resourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"gateways": core.MappingNodeFields(
				"natGateways", core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("nat-12345678"),
					),
				),
			),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service]{
		Name: "VPC with NAT gateways is not stabilised when NAT gateway is in pending state",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpec,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: false,
		},
		ExpectError: false,
	}
}

func createVPCWithNATGatewayInPendingStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
			{
				Vpcs: []types.Vpc{
					{
						VpcId: aws.String("vpc-12345678"),
						State: types.VpcStateAvailable,
					},
				},
			},
		}),
		ec2mock.WithDescribeNatGatewaysOutput(&ec2.DescribeNatGatewaysOutput{
			NatGateways: []types.NatGateway{
				{
					NatGatewayId: aws.String("nat-12345678"),
					State:        types.NatGatewayStatePending,
				},
			},
		}),
	)

	resourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"gateways": core.MappingNodeFields(
				"natGateways", core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("nat-12345678"),
					),
				),
			),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service]{
		Name: "VPC is not stabilised when NAT gateway is in pending state",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpec,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: false,
		},
		ExpectError: false,
	}
}

func createVPCWithNATGatewayInFailedStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
			{
				Vpcs: []types.Vpc{
					{
						VpcId: aws.String("vpc-12345678"),
						State: types.VpcStateAvailable,
					},
				},
			},
		}),
		ec2mock.WithDescribeNatGatewaysOutput(&ec2.DescribeNatGatewaysOutput{
			NatGateways: []types.NatGateway{
				{
					NatGatewayId: aws.String("nat-12345678"),
					State:        types.NatGatewayStateFailed,
				},
			},
		}),
	)

	resourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"gateways": core.MappingNodeFields(
				"natGateways", core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("nat-12345678"),
					),
				),
			),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service]{
		Name: "VPC is not stabilised when NAT gateway is in failed state",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpec,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: false,
		},
		ExpectError: false,
	}
}

func createVPCWithMultipleNATGatewaysTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
			{
				Vpcs: []types.Vpc{
					{
						VpcId: aws.String("vpc-12345678"),
						State: types.VpcStateAvailable,
					},
				},
			},
		}),
		ec2mock.WithDescribeNatGatewaysOutput(&ec2.DescribeNatGatewaysOutput{
			NatGateways: []types.NatGateway{
				{
					NatGatewayId: aws.String("nat-12345678"),
					State:        types.NatGatewayStateAvailable,
				},
				{
					NatGatewayId: aws.String("nat-87654321"),
					State:        types.NatGatewayStateAvailable,
				},
			},
		}),
	)

	resourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"gateways": core.MappingNodeFields(
				"natGateways", core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("nat-12345678"),
					),
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("nat-87654321"),
					),
				),
			),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service]{
		Name: "VPC with multiple NAT gateways is stabilised when all NAT gateways are available",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpec,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
		ExpectError: false,
	}
}

func createVPCWithMissingVPCIDTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock()

	resourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			// Missing vpcId field
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when vpcId is missing",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpec,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createVPCWithDescribeVpcsErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsError(errors.New("failed to describe VPCs")),
	)

	resourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when DescribeVpcs fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpec,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createVPCWithDescribeNatGatewaysErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
			{
				Vpcs: []types.Vpc{
					{
						VpcId: aws.String("vpc-12345678"),
						State: types.VpcStateAvailable,
					},
				},
			},
		}),
		ec2mock.WithDescribeNatGatewaysError(errors.New("failed to describe NAT gateways")),
	)

	resourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"gateways": core.MappingNodeFields(
				"natGateways", core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("nat-12345678"),
					),
				),
			),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when DescribeNatGateways fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpec,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createVPCWithNoNATGatewaysTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
			{
				Vpcs: []types.Vpc{
					{
						VpcId: aws.String("vpc-12345678"),
						State: types.VpcStateAvailable,
					},
				},
			},
		}),
	)

	resourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			// No gateways field
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service]{
		Name: "VPC without NAT gateways is stabilised when VPC is available",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpec,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
		ExpectError: false,
	}
}

func createVPCWithEmptyNATGatewaysArrayTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
			{
				Vpcs: []types.Vpc{
					{
						VpcId: aws.String("vpc-12345678"),
						State: types.VpcStateAvailable,
					},
				},
			},
		}),
	)

	resourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"vpcId": core.MappingNodeFromString("vpc-12345678"),
			"gateways": core.MappingNodeFields(
				"natGateways", core.MappingNodeItems(), // Empty array
			),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ec2service.Service]{
		Name: "VPC with empty NAT gateways array is stabilised when VPC is available",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpec,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
		ExpectError: false,
	}
}

func TestFlexVPCResourceStabilised(t *testing.T) {
	suite.Run(t, new(FlexVPCResourceStabilisedSuite))
}
