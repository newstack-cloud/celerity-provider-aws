//go:build unit

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
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FlexVPCResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *FlexVPCResourceGetExternalStateSuite) Test_get_external_state() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"sessionID": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service]{
		createVPCStateTestCase(providerCtx, loader),
		createGetVPCErrorTestCase(providerCtx, loader),
		createGetVPCAttributesErrorTestCase(providerCtx, loader),
		createGetSubnetsErrorTestCase(providerCtx, loader),
		createGetRouteTablesErrorTestCase(providerCtx, loader),
		createGetSecurityGroupsErrorTestCase(providerCtx, loader),
		createGetNetworkACLsErrorTestCase(providerCtx, loader),
		createGetInternetGatewaysErrorTestCase(providerCtx, loader),
		createGetNATGatewaysErrorTestCase(providerCtx, loader),
		createMissingNameFieldTestCase(providerCtx, loader),
		createVPCWithPresetTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		func(serviceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service], configStore pluginutils.ServiceConfigStore[*aws.Config]) provider.Resource {
			return VPCResource(serviceFactory, nil, configStore)
		},
		&s.Suite,
	)
}

func createVPCStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service] {
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name": core.MappingNodeFromString("test-vpc"),
			"mode": core.MappingNodeFromString("create"),
		},
	}

	expectedResourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":               core.MappingNodeFromString("test-vpc"),
			"mode":               core.MappingNodeFromString("create"),
			"preset":             nil,
			"cidrBlock":          core.MappingNodeFromString("10.0.0.0/16"),
			"enableDNSSupport":   core.MappingNodeFromBool(true),
			"enableDNSHostnames": core.MappingNodeFromBool(true),
			"region":             core.MappingNodeFromString("us-west-2"),
			"tags":               nil,
			"vpcId":              core.MappingNodeFromString("vpc-12345678"),
			"subnets": core.MappingNodeFields(
				"private-subnet-1", core.MappingNodeFields(
					"id", core.MappingNodeFromString("subnet-12345678"),
					"availabilityZone", core.MappingNodeFromString("us-west-2a"),
					"subnetType", core.MappingNodeFromString("private"),
				),
			),
			"privateSubnetIds": core.MappingNodeItems(
				core.MappingNodeFromString("subnet-12345678"),
			),
			"publicSubnetIds": core.MappingNodeItems(),
			"routeTables": core.MappingNodeItems(
				core.MappingNodeFields(
					"id", core.MappingNodeFromString("rtb-12345678"),
					"subnetIds", core.MappingNodeItems(core.MappingNodeFromString("subnet-12345678")),
				),
			),
			"securityGroupIds": core.MappingNodeItems(
				core.MappingNodeFromString("sg-12345678"),
			),
			"securityGroupIdsByName": {Fields: map[string]*core.MappingNode{}},
			"networkAcls": core.MappingNodeItems(
				core.MappingNodeFields(
					"id", core.MappingNodeFromString("acl-12345678"),
					"subnetIds", core.MappingNodeItems(core.MappingNodeFromString("subnet-12345678")),
				),
			),
			"gateways": core.MappingNodeFields(
				"internetGatewayId", core.MappingNodeFromString("igw-12345678"),
				"natGateways", core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("nat-12345678"),
						"elasticIpId", core.MappingNodeFromString("eip-12345678"),
						"inPublicSubnetId", core.MappingNodeFromString("subnet-public"),
						"forPrivateSubnetId", core.MappingNodeFromString("subnet-12345678"),
					),
				),
			),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service]{
		Name: "successfully gets VPC with all components",
		ServiceFactory: ec2mock.CreateEc2ServiceMockFactory(
			ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
				{
					Vpcs: []types.Vpc{
						{
							VpcId:     aws.String("vpc-12345678"),
							CidrBlock: aws.String("10.0.0.0/16"),
							Tags: []types.Tag{
								{
									Key:   aws.String("flex-vpc-name"),
									Value: aws.String("test-vpc"),
								},
							},
						},
					},
				},
			},
			),
			ec2mock.WithDescribeVpcAttributeOutput(&ec2.DescribeVpcAttributeOutput{
				EnableDnsSupport: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
				EnableDnsHostnames: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
			}),
			ec2mock.WithDescribeSubnetsOutput(&ec2.DescribeSubnetsOutput{
				Subnets: []types.Subnet{
					{
						SubnetId:         aws.String("subnet-12345678"),
						AvailabilityZone: aws.String("us-west-2a"),
						Tags: []types.Tag{
							{
								Key:   aws.String("flex-vpc-name"),
								Value: aws.String("test-vpc"),
							},
							{
								Key:   aws.String(TagFlexVPCSubnetName),
								Value: aws.String("private-subnet-1"),
							},
							{
								Key:   aws.String(TagFlexVPCSubnetType),
								Value: aws.String("private"),
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
				RouteTables: []types.RouteTable{
					{
						RouteTableId: aws.String("rtb-12345678"),
						Associations: []types.RouteTableAssociation{
							{
								SubnetId: aws.String("subnet-12345678"),
							},
						},
						Routes: []types.Route{
							{
								NatGatewayId: aws.String("nat-12345678"),
							},
						},
						Tags: []types.Tag{
							{
								Key:   aws.String("flex-vpc-name"),
								Value: aws.String("test-vpc"),
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []types.SecurityGroup{
					{
						GroupId: aws.String("sg-12345678"),
						Tags: []types.Tag{
							{
								Key:   aws.String("flex-vpc-name"),
								Value: aws.String("test-vpc"),
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeNetworkAclsOutput(&ec2.DescribeNetworkAclsOutput{
				NetworkAcls: []types.NetworkAcl{
					{
						NetworkAclId: aws.String("acl-12345678"),
						Associations: []types.NetworkAclAssociation{
							{
								SubnetId: aws.String("subnet-12345678"),
							},
						},
						Tags: []types.Tag{
							{
								Key:   aws.String("flex-vpc-name"),
								Value: aws.String("test-vpc"),
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeInternetGatewaysOutput(&ec2.DescribeInternetGatewaysOutput{
				InternetGateways: []types.InternetGateway{
					{
						InternetGatewayId: aws.String("igw-12345678"),
						Tags: []types.Tag{
							{
								Key:   aws.String("flex-vpc-name"),
								Value: aws.String("test-vpc"),
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeNatGatewaysOutput(&ec2.DescribeNatGatewaysOutput{
				NatGateways: []types.NatGateway{
					{
						NatGatewayId: aws.String("nat-12345678"),
						SubnetId:     aws.String("subnet-public"),
						NatGatewayAddresses: []types.NatGatewayAddress{
							{
								AllocationId: aws.String("eip-12345678"),
							},
						},
						Tags: []types.Tag{
							{
								Key:   aws.String("flex-vpc-name"),
								Value: aws.String("test-vpc"),
							},
						},
					},
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-vpc",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: expectedResourceSpecState,
		},
	}
}

func createGetVPCErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service] {
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name": core.MappingNodeFromString("test-vpc"),
			"mode": core.MappingNodeFromString("basic"),
		},
	}

	// When VPC is not found by flex VPC name tag, the fallback lookup is attempted.
	// If tagging is not enabled (default in tests), fallback returns nil and we return empty state.
	// This is correct behavior - resource doesn't exist yet.
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service]{
		Name: "returns empty state when VPC not found",
		ServiceFactory: ec2mock.CreateEc2ServiceMockFactory(
			ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
				{
					Vpcs: []types.Vpc{},
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-vpc",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{},
			},
		},
		ExpectError: false,
	}
}

func createGetVPCAttributesErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service] {
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name": core.MappingNodeFromString("test-vpc"),
			"mode": core.MappingNodeFromString("basic"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when VPC attributes cannot be retrieved",
		ServiceFactory: ec2mock.CreateEc2ServiceMockFactory(
			ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
				{
					Vpcs: []types.Vpc{
						{
							VpcId:     aws.String("vpc-12345678"),
							CidrBlock: aws.String("10.0.0.0/16"),
							Tags: []types.Tag{
								{
									Key:   aws.String("flex-vpc-name"),
									Value: aws.String("test-vpc"),
								},
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeVpcAttributeError(errors.New("failed to get VPC attributes")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-vpc",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createGetSubnetsErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service] {
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name": core.MappingNodeFromString("test-vpc"),
			"mode": core.MappingNodeFromString("basic"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when subnets cannot be retrieved",
		ServiceFactory: ec2mock.CreateEc2ServiceMockFactory(
			ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
				{
					Vpcs: []types.Vpc{
						{
							VpcId:     aws.String("vpc-12345678"),
							CidrBlock: aws.String("10.0.0.0/16"),
							Tags: []types.Tag{
								{
									Key:   aws.String("flex-vpc-name"),
									Value: aws.String("test-vpc"),
								},
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeVpcAttributeOutput(&ec2.DescribeVpcAttributeOutput{
				EnableDnsSupport: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
				EnableDnsHostnames: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
			}),
			ec2mock.WithDescribeSubnetsError(errors.New("failed to get subnets")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-vpc",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createGetRouteTablesErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service] {
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name": core.MappingNodeFromString("test-vpc"),
			"mode": core.MappingNodeFromString("basic"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when route tables cannot be retrieved",
		ServiceFactory: ec2mock.CreateEc2ServiceMockFactory(
			ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
				{
					Vpcs: []types.Vpc{
						{
							VpcId:     aws.String("vpc-12345678"),
							CidrBlock: aws.String("10.0.0.0/16"),
							Tags: []types.Tag{
								{
									Key:   aws.String("flex-vpc-name"),
									Value: aws.String("test-vpc"),
								},
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeVpcAttributeOutput(&ec2.DescribeVpcAttributeOutput{
				EnableDnsSupport: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
				EnableDnsHostnames: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
			}),
			ec2mock.WithDescribeSubnetsOutput(&ec2.DescribeSubnetsOutput{
				Subnets: []types.Subnet{},
			}),
			ec2mock.WithDescribeRouteTablesError(errors.New("failed to get route tables")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-vpc",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createGetSecurityGroupsErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service] {
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name": core.MappingNodeFromString("test-vpc"),
			"mode": core.MappingNodeFromString("basic"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when security groups cannot be retrieved",
		ServiceFactory: ec2mock.CreateEc2ServiceMockFactory(
			ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
				{
					Vpcs: []types.Vpc{
						{
							VpcId:     aws.String("vpc-12345678"),
							CidrBlock: aws.String("10.0.0.0/16"),
							Tags: []types.Tag{
								{
									Key:   aws.String("flex-vpc-name"),
									Value: aws.String("test-vpc"),
								},
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeVpcAttributeOutput(&ec2.DescribeVpcAttributeOutput{
				EnableDnsSupport: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
				EnableDnsHostnames: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
			}),
			ec2mock.WithDescribeSubnetsOutput(&ec2.DescribeSubnetsOutput{
				Subnets: []types.Subnet{},
			}),
			ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
				RouteTables: []types.RouteTable{},
			}),
			ec2mock.WithDescribeSecurityGroupsError(errors.New("failed to get security groups")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-vpc",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createGetNetworkACLsErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service] {
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name": core.MappingNodeFromString("test-vpc"),
			"mode": core.MappingNodeFromString("basic"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when network ACLs cannot be retrieved",
		ServiceFactory: ec2mock.CreateEc2ServiceMockFactory(
			ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
				{
					Vpcs: []types.Vpc{
						{
							VpcId:     aws.String("vpc-12345678"),
							CidrBlock: aws.String("10.0.0.0/16"),
							Tags: []types.Tag{
								{
									Key:   aws.String("flex-vpc-name"),
									Value: aws.String("test-vpc"),
								},
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeVpcAttributeOutput(&ec2.DescribeVpcAttributeOutput{
				EnableDnsSupport: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
				EnableDnsHostnames: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
			}),
			ec2mock.WithDescribeSubnetsOutput(&ec2.DescribeSubnetsOutput{
				Subnets: []types.Subnet{},
			}),
			ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
				RouteTables: []types.RouteTable{},
			}),
			ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []types.SecurityGroup{},
			}),
			ec2mock.WithDescribeNetworkAclsError(errors.New("failed to get network ACLs")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-vpc",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createGetInternetGatewaysErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service] {
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name": core.MappingNodeFromString("test-vpc"),
			"mode": core.MappingNodeFromString("basic"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when internet gateways cannot be retrieved",
		ServiceFactory: ec2mock.CreateEc2ServiceMockFactory(
			ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
				{
					Vpcs: []types.Vpc{
						{
							VpcId:     aws.String("vpc-12345678"),
							CidrBlock: aws.String("10.0.0.0/16"),
							Tags: []types.Tag{
								{
									Key:   aws.String("flex-vpc-name"),
									Value: aws.String("test-vpc"),
								},
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeVpcAttributeOutput(&ec2.DescribeVpcAttributeOutput{
				EnableDnsSupport: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
				EnableDnsHostnames: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
			}),
			ec2mock.WithDescribeSubnetsOutput(&ec2.DescribeSubnetsOutput{
				Subnets: []types.Subnet{},
			}),
			ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
				RouteTables: []types.RouteTable{},
			}),
			ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []types.SecurityGroup{},
			}),
			ec2mock.WithDescribeNetworkAclsOutput(&ec2.DescribeNetworkAclsOutput{
				NetworkAcls: []types.NetworkAcl{},
			}),
			ec2mock.WithDescribeInternetGatewaysError(errors.New("failed to get internet gateways")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-vpc",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createGetNATGatewaysErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service] {
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name": core.MappingNodeFromString("test-vpc"),
			"mode": core.MappingNodeFromString("basic"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when NAT gateways cannot be retrieved",
		ServiceFactory: ec2mock.CreateEc2ServiceMockFactory(
			ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
				{
					Vpcs: []types.Vpc{
						{
							VpcId:     aws.String("vpc-12345678"),
							CidrBlock: aws.String("10.0.0.0/16"),
							Tags: []types.Tag{
								{
									Key:   aws.String("flex-vpc-name"),
									Value: aws.String("test-vpc"),
								},
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeVpcAttributeOutput(&ec2.DescribeVpcAttributeOutput{
				EnableDnsSupport: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
				EnableDnsHostnames: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
			}),
			ec2mock.WithDescribeSubnetsOutput(&ec2.DescribeSubnetsOutput{
				Subnets: []types.Subnet{},
			}),
			ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
				RouteTables: []types.RouteTable{},
			}),
			ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []types.SecurityGroup{},
			}),
			ec2mock.WithDescribeNetworkAclsOutput(&ec2.DescribeNetworkAclsOutput{
				NetworkAcls: []types.NetworkAcl{},
			}),
			ec2mock.WithDescribeInternetGatewaysOutput(&ec2.DescribeInternetGatewaysOutput{
				InternetGateways: []types.InternetGateway{
					{
						InternetGatewayId: aws.String("igw-12345678"),
						Tags: []types.Tag{
							{
								Key:   aws.String("flex-vpc-name"),
								Value: aws.String("test-vpc"),
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeNatGatewaysError(errors.New("failed to get NAT gateways")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-vpc",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createMissingNameFieldTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service] {
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"mode": core.MappingNodeFromString("basic"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service]{
		Name:           "returns error when name field is missing",
		ServiceFactory: ec2mock.CreateEc2ServiceMockFactory(),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-vpc",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createVPCWithPresetTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service] {
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":   core.MappingNodeFromString("test-vpc"),
			"mode":   core.MappingNodeFromString("basic"),
			"preset": core.MappingNodeFromString("production"),
		},
	}

	// When no items are found, the code returns &core.MappingNode{} which has nil Items,
	// not an empty slice from core.MappingNodeItems()
	expectedResourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":                   core.MappingNodeFromString("test-vpc"),
			"mode":                   core.MappingNodeFromString("basic"),
			"preset":                 core.MappingNodeFromString("production"),
			"cidrBlock":              core.MappingNodeFromString("10.0.0.0/16"),
			"enableDNSSupport":       core.MappingNodeFromBool(true),
			"enableDNSHostnames":     core.MappingNodeFromBool(true),
			"region":                 core.MappingNodeFromString("us-west-2"),
			"tags":                   nil,
			"vpcId":                  core.MappingNodeFromString("vpc-12345678"),
			"subnets":                core.MappingNodeFields(),
			"privateSubnetIds":       core.MappingNodeItems(),
			"publicSubnetIds":        core.MappingNodeItems(),
			"routeTables":            {},
			"securityGroupIds":       {},
			"securityGroupIdsByName": {Fields: map[string]*core.MappingNode{}},
			"networkAcls":            {},
			"gateways": core.MappingNodeFields(
				"internetGatewayId", core.MappingNodeFromString("igw-12345678"),
				"natGateways", &core.MappingNode{},
			),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ec2service.Service]{
		Name: "successfully gets VPC with preset configuration",
		ServiceFactory: ec2mock.CreateEc2ServiceMockFactory(
			ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
				{
					Vpcs: []types.Vpc{
						{
							VpcId:     aws.String("vpc-12345678"),
							CidrBlock: aws.String("10.0.0.0/16"),
							Tags: []types.Tag{
								{
									Key:   aws.String("flex-vpc-name"),
									Value: aws.String("test-vpc"),
								},
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeVpcAttributeOutput(&ec2.DescribeVpcAttributeOutput{
				EnableDnsSupport: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
				EnableDnsHostnames: &types.AttributeBooleanValue{
					Value: aws.Bool(true),
				},
			}),
			ec2mock.WithDescribeSubnetsOutput(&ec2.DescribeSubnetsOutput{
				Subnets: []types.Subnet{},
			}),
			ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
				RouteTables: []types.RouteTable{},
			}),
			ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []types.SecurityGroup{},
			}),
			ec2mock.WithDescribeNetworkAclsOutput(&ec2.DescribeNetworkAclsOutput{
				NetworkAcls: []types.NetworkAcl{},
			}),
			ec2mock.WithDescribeInternetGatewaysOutput(&ec2.DescribeInternetGatewaysOutput{
				InternetGateways: []types.InternetGateway{
					{
						InternetGatewayId: aws.String("igw-12345678"),
						Tags: []types.Tag{
							{
								Key:   aws.String("flex-vpc-name"),
								Value: aws.String("test-vpc"),
							},
						},
					},
				},
			}),
			ec2mock.WithDescribeNatGatewaysOutput(&ec2.DescribeNatGatewaysOutput{
				NatGateways: []types.NatGateway{},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-vpc",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: expectedResourceSpecState,
		},
	}
}

func TestFlexVPCResourceGetExternalState(t *testing.T) {
	suite.Run(t, new(FlexVPCResourceGetExternalStateSuite))
}
