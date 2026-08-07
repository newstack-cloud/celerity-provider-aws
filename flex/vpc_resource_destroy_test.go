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
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FlexVPCResourceDestroySuite struct {
	suite.Suite
}

func (s *FlexVPCResourceDestroySuite) Test_destroy() {
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

	testCases := []plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		createSuccessfulDestroyTestCase(providerCtx, loader),
		createReferenceModeDestroyTestCase(providerCtx, loader),
		createMissingVPCIDTestCase(providerCtx, loader),
		createPartialCreateDestroyTestCase(providerCtx, loader),
		createPartialCreateNothingDeployedTestCase(providerCtx, loader),
		createMissingModeTestCase(providerCtx, loader),
		createNetworkACLDisassociationErrorTestCase(providerCtx, loader),
		createNetworkACLDeletionErrorTestCase(providerCtx, loader),
		createSecurityGroupDeletionErrorTestCase(providerCtx, loader),
		createRouteTableDisassociationErrorTestCase(providerCtx, loader),
		createRouteTableDeletionErrorTestCase(providerCtx, loader),
		createNATGatewayDeletionErrorTestCase(providerCtx, loader),
		createElasticIPReleaseErrorTestCase(providerCtx, loader),
		createInternetGatewayDetachmentErrorTestCase(providerCtx, loader),
		createInternetGatewayDeletionErrorTestCase(providerCtx, loader),
		createSubnetDeletionErrorTestCase(providerCtx, loader),
		createVPCDeletionErrorTestCase(providerCtx, loader),
		createDefaultNetworkACLNotFoundTestCase(providerCtx, loader),
		createNetworkACLNotFoundTestCase(providerCtx, loader),
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

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		vpcResourceWrapper,
		&s.Suite,
	)
}

func createSuccessfulDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeNetworkAclsOutput(&ec2.DescribeNetworkAclsOutput{
			NetworkAcls: []types.NetworkAcl{
				{
					NetworkAclId: aws.String("acl-default"),
					IsDefault:    aws.Bool(true),
					VpcId:        aws.String("vpc-12345678"),
					Associations: []types.NetworkAclAssociation{
						{
							NetworkAclAssociationId: aws.String("aclassoc-default"),
							SubnetId:                aws.String("subnet-12345678"),
						},
					},
				},
			},
		}),
		ec2mock.WithReplaceNetworkAclAssociationOutput(&ec2.ReplaceNetworkAclAssociationOutput{
			NewAssociationId: aws.String("aclassoc-new"),
		}),
		ec2mock.WithDeleteNetworkAclOutput(&ec2.DeleteNetworkAclOutput{}),
		ec2mock.WithDeleteSecurityGroupOutput(&ec2.DeleteSecurityGroupOutput{}),
		ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
			RouteTables: []types.RouteTable{
				{
					RouteTableId: aws.String("rtb-12345678"),
					Associations: []types.RouteTableAssociation{
						{
							RouteTableAssociationId: aws.String("rtbassoc-12345678"),
							SubnetId:                aws.String("subnet-12345678"),
						},
					},
				},
			},
		}),
		ec2mock.WithDisassociateRouteTableOutput(&ec2.DisassociateRouteTableOutput{}),
		ec2mock.WithDeleteRouteTableOutput(&ec2.DeleteRouteTableOutput{}),
		ec2mock.WithDeleteNatGatewayOutput(&ec2.DeleteNatGatewayOutput{}),
		// The elastic IP stays attached until its NAT gateway has finished
		// deleting, so teardown waits on the gateway before releasing the address.
		ec2mock.WithDescribeNatGatewaysOutput(&ec2.DescribeNatGatewaysOutput{
			NatGateways: []types.NatGateway{
				{
					NatGatewayId: aws.String("nat-12345678"),
					State:        types.NatGatewayStateDeleted,
				},
			},
		}),
		ec2mock.WithReleaseAddressOutput(&ec2.ReleaseAddressOutput{}),
		ec2mock.WithDetachInternetGatewayOutput(&ec2.DetachInternetGatewayOutput{}),
		ec2mock.WithDeleteInternetGatewayOutput(&ec2.DeleteInternetGatewayOutput{}),
		ec2mock.WithDeleteSubnetOutput(&ec2.DeleteSubnetOutput{}),
		ec2mock.WithDeleteVpcOutput(&ec2.DeleteVpcOutput{}),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
				"networkAcls": core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("acl-12345678"),
					),
				),
				"securityGroupIds": core.MappingNodeItems(
					core.MappingNodeFromString("sg-12345678"),
				),
				"routeTables": core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("rtb-12345678"),
					),
				),
				"gateways": core.MappingNodeFields(
					"internetGatewayId", core.MappingNodeFromString("igw-12345678"),
					"natGateways", core.MappingNodeItems(
						core.MappingNodeFields(
							"id", core.MappingNodeFromString("nat-12345678"),
							"elasticIpId", core.MappingNodeFromString("eip-12345678"),
						),
					),
				),
				// Subnets are held in state as a map keyed by the preset's subnet
				// name, matching what the create path writes.
				"subnets": core.MappingNodeFields(
					"public-az-1", core.MappingNodeFields(
						"id", core.MappingNodeFromString("subnet-12345678"),
					),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "successfully destroys VPC with all components",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"DescribeNetworkAcls": []any{
				&ec2.DescribeNetworkAclsInput{
					Filters: []types.Filter{
						{
							Name:   aws.String("default"),
							Values: []string{"true"},
						},
						{
							Name:   aws.String("vpc-id"),
							Values: []string{"vpc-12345678"},
						},
					},
				},
				&ec2.DescribeNetworkAclsInput{
					Filters: []types.Filter{
						{
							Name:   aws.String("network-acl-id"),
							Values: []string{"acl-12345678"},
						},
					},
				},
			},
			"ReplaceNetworkAclAssociation": &ec2.ReplaceNetworkAclAssociationInput{
				AssociationId: aws.String("aclassoc-default"),
				NetworkAclId:  aws.String("acl-default"),
			},
			"DeleteNetworkAcl": &ec2.DeleteNetworkAclInput{
				NetworkAclId: aws.String("acl-12345678"),
			},
			"DeleteSecurityGroup": &ec2.DeleteSecurityGroupInput{
				GroupId: aws.String("sg-12345678"),
			},
			"DescribeRouteTables": &ec2.DescribeRouteTablesInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("vpc-id"),
						Values: []string{"vpc-12345678"},
					},
				},
			},
			"DisassociateRouteTable": &ec2.DisassociateRouteTableInput{
				AssociationId: aws.String("rtbassoc-12345678"),
			},
			"DeleteRouteTable": &ec2.DeleteRouteTableInput{
				RouteTableId: aws.String("rtb-12345678"),
			},
			"DeleteNatGateway": &ec2.DeleteNatGatewayInput{
				NatGatewayId: aws.String("nat-12345678"),
			},
			"ReleaseAddress": &ec2.ReleaseAddressInput{
				AllocationId: aws.String("eip-12345678"),
			},
			"DetachInternetGateway": &ec2.DetachInternetGatewayInput{
				InternetGatewayId: aws.String("igw-12345678"),
				VpcId:             aws.String("vpc-12345678"),
			},
			"DeleteInternetGateway": &ec2.DeleteInternetGatewayInput{
				InternetGatewayId: aws.String("igw-12345678"),
			},
			"DeleteSubnet": &ec2.DeleteSubnetInput{
				SubnetId: aws.String("subnet-12345678"),
			},
			"DeleteVpc": &ec2.DeleteVpcInput{
				VpcId: aws.String("vpc-12345678"),
			},
		},
	}
}

func createReferenceModeDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	// Create a simple mock - the DestroyActionsNotCalled field will verify no destructive calls are made
	service := ec2mock.CreateEc2ServiceMock()

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("reference"),
				// Include all VPC components to ensure they are NOT deleted in reference mode
				"networkAcls": core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("acl-12345678"),
					),
				),
				"securityGroupIds": core.MappingNodeItems(
					core.MappingNodeFromString("sg-12345678"),
				),
				"routeTables": core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("rtb-12345678"),
					),
				),
				"gateways": core.MappingNodeFields(
					"internetGatewayId", core.MappingNodeFromString("igw-12345678"),
					"natGateways", core.MappingNodeItems(
						core.MappingNodeFields(
							"id", core.MappingNodeFromString("nat-12345678"),
							"elasticIpId", core.MappingNodeFromString("eip-12345678"),
						),
					),
				),
				// Subnets are held in state as a map keyed by the preset's subnet
				// name, matching what the create path writes.
				"subnets": core.MappingNodeFields(
					"public-az-1", core.MappingNodeFields(
						"id", core.MappingNodeFromString("subnet-12345678"),
					),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "skips destruction when mode is reference - verifies no destructive AWS EC2 API calls are made",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: false,
		// Reference mode should return early without making any AWS API calls
		DestroyActionsCalled: map[string]any{},
		DestroyActionsNotCalled: []string{
			"DescribeNetworkAcls",
			"ReplaceNetworkAclAssociation",
			"DeleteNetworkAcl",
			"DeleteSecurityGroup",
			"DescribeRouteTables",
			"DisassociateRouteTable",
			"DeleteRouteTable",
			"DeleteNatGateway",
			"ReleaseAddress",
			"DetachInternetGateway",
			"DeleteInternetGateway",
			"DeleteSubnet",
			"DeleteVpc",
		},
	}
}

func createMissingVPCIDTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"mode": core.MappingNodeFromString("create"),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when neither vpcId nor name is available to locate the VPC",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return ec2mock.CreateEc2ServiceMock()
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError:          true,
		DestroyActionsCalled: map[string]any{},
	}
}

func createMissingModeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when mode is missing",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return ec2mock.CreateEc2ServiceMock()
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError:          true,
		DestroyActionsCalled: map[string]any{},
	}
}

func createNetworkACLDisassociationErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeNetworkAclsError(errors.New("failed to describe network ACLs")),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
				"networkAcls": core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("acl-12345678"),
					),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when network ACL disassociation fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DescribeNetworkAcls": &ec2.DescribeNetworkAclsInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("default"),
						Values: []string{"true"},
					},
					{
						Name:   aws.String("vpc-id"),
						Values: []string{"vpc-12345678"},
					},
				},
			},
		},
	}
}

func createNetworkACLDeletionErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeNetworkAclsOutput(&ec2.DescribeNetworkAclsOutput{
			NetworkAcls: []types.NetworkAcl{
				{
					NetworkAclId: aws.String("acl-default"),
					IsDefault:    aws.Bool(true),
					VpcId:        aws.String("vpc-12345678"),
					Associations: []types.NetworkAclAssociation{
						{
							NetworkAclAssociationId: aws.String("aclassoc-default"),
							SubnetId:                aws.String("subnet-12345678"),
						},
					},
				},
			},
		}),
		ec2mock.WithReplaceNetworkAclAssociationOutput(&ec2.ReplaceNetworkAclAssociationOutput{
			NewAssociationId: aws.String("aclassoc-new"),
		}),
		ec2mock.WithDeleteNetworkAclError(errors.New("failed to delete network ACL")),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
				"networkAcls": core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("acl-12345678"),
					),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when network ACL deletion fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DescribeNetworkAcls": []any{
				&ec2.DescribeNetworkAclsInput{
					Filters: []types.Filter{
						{
							Name:   aws.String("default"),
							Values: []string{"true"},
						},
						{
							Name:   aws.String("vpc-id"),
							Values: []string{"vpc-12345678"},
						},
					},
				},
				&ec2.DescribeNetworkAclsInput{
					Filters: []types.Filter{
						{
							Name:   aws.String("network-acl-id"),
							Values: []string{"acl-12345678"},
						},
					},
				},
			},
			"ReplaceNetworkAclAssociation": &ec2.ReplaceNetworkAclAssociationInput{
				AssociationId: aws.String("aclassoc-default"),
				NetworkAclId:  aws.String("acl-default"),
			},
			"DeleteNetworkAcl": &ec2.DeleteNetworkAclInput{
				NetworkAclId: aws.String("acl-12345678"),
			},
		},
	}
}

func createSecurityGroupDeletionErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDeleteSecurityGroupError(errors.New("failed to delete security group")),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
				"securityGroupIds": core.MappingNodeItems(
					core.MappingNodeFromString("sg-12345678"),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when security group deletion fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteSecurityGroup": &ec2.DeleteSecurityGroupInput{
				GroupId: aws.String("sg-12345678"),
			},
		},
	}
}

func createRouteTableDisassociationErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeRouteTablesError(errors.New("failed to describe route tables")),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
				"routeTables": core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("rtb-12345678"),
					),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when route table disassociation fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DescribeRouteTables": &ec2.DescribeRouteTablesInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("vpc-id"),
						Values: []string{"vpc-12345678"},
					},
				},
			},
		},
	}
}

func createRouteTableDeletionErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
			RouteTables: []types.RouteTable{},
		}),
		ec2mock.WithDeleteRouteTableError(errors.New("failed to delete route table")),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
				"routeTables": core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("rtb-12345678"),
					),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when route table deletion fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DescribeRouteTables": &ec2.DescribeRouteTablesInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("vpc-id"),
						Values: []string{"vpc-12345678"},
					},
				},
			},
			"DeleteRouteTable": &ec2.DeleteRouteTableInput{
				RouteTableId: aws.String("rtb-12345678"),
			},
		},
	}
}

func createNATGatewayDeletionErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDeleteNatGatewayError(errors.New("failed to delete NAT gateway")),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
				"gateways": core.MappingNodeFields(
					"natGateways", core.MappingNodeItems(
						core.MappingNodeFields(
							"id", core.MappingNodeFromString("nat-12345678"),
						),
					),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when NAT gateway deletion fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteNatGateway": &ec2.DeleteNatGatewayInput{
				NatGatewayId: aws.String("nat-12345678"),
			},
		},
	}
}

func createElasticIPReleaseErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithReleaseAddressError(errors.New("failed to release elastic IP")),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
				"gateways": core.MappingNodeFields(
					"natGateways", core.MappingNodeItems(
						core.MappingNodeFields(
							"elasticIpId", core.MappingNodeFromString("eip-12345678"),
						),
					),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when elastic IP release fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"ReleaseAddress": &ec2.ReleaseAddressInput{
				AllocationId: aws.String("eip-12345678"),
			},
		},
	}
}

func createInternetGatewayDetachmentErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDetachInternetGatewayError(errors.New("failed to detach internet gateway")),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
				"gateways": core.MappingNodeFields(
					"internetGatewayId", core.MappingNodeFromString("igw-12345678"),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when internet gateway detachment fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DetachInternetGateway": &ec2.DetachInternetGatewayInput{
				InternetGatewayId: aws.String("igw-12345678"),
				VpcId:             aws.String("vpc-12345678"),
			},
		},
	}
}

func createInternetGatewayDeletionErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDetachInternetGatewayOutput(&ec2.DetachInternetGatewayOutput{}),
		ec2mock.WithDeleteInternetGatewayError(errors.New("failed to delete internet gateway")),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
				"gateways": core.MappingNodeFields(
					"internetGatewayId", core.MappingNodeFromString("igw-12345678"),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when internet gateway deletion fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DetachInternetGateway": &ec2.DetachInternetGatewayInput{
				InternetGatewayId: aws.String("igw-12345678"),
				VpcId:             aws.String("vpc-12345678"),
			},
			"DeleteInternetGateway": &ec2.DeleteInternetGatewayInput{
				InternetGatewayId: aws.String("igw-12345678"),
			},
		},
	}
}

func createSubnetDeletionErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDeleteSubnetError(errors.New("failed to delete subnet")),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
				// Subnets are held in state as a map keyed by the preset's subnet
				// name, matching what the create path writes.
				"subnets": core.MappingNodeFields(
					"public-az-1", core.MappingNodeFields(
						"id", core.MappingNodeFromString("subnet-12345678"),
					),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when subnet deletion fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteSubnet": &ec2.DeleteSubnetInput{
				SubnetId: aws.String("subnet-12345678"),
			},
		},
	}
}

func createVPCDeletionErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDeleteVpcError(errors.New("failed to delete VPC")),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when VPC deletion fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteVpc": &ec2.DeleteVpcInput{
				VpcId: aws.String("vpc-12345678"),
			},
		},
	}
}

func createDefaultNetworkACLNotFoundTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeNetworkAclsOutput(&ec2.DescribeNetworkAclsOutput{
			NetworkAcls: []types.NetworkAcl{},
		}),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
				"networkAcls": core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("acl-12345678"),
					),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when default network ACL not found",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DescribeNetworkAcls": &ec2.DescribeNetworkAclsInput{
				Filters: []types.Filter{
					{
						Name:   aws.String("default"),
						Values: []string{"true"},
					},
					{
						Name:   aws.String("vpc-id"),
						Values: []string{"vpc-12345678"},
					},
				},
			},
		},
	}
}

func createNetworkACLNotFoundTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeNetworkAclsOutput(&ec2.DescribeNetworkAclsOutput{
			NetworkAcls: []types.NetworkAcl{
				{
					NetworkAclId: aws.String("acl-default"),
					IsDefault:    aws.Bool(true),
					VpcId:        aws.String("vpc-12345678"),
				},
			},
		}),
		ec2mock.WithDescribeNetworkAclsError(errors.New("network ACL not found")),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"vpcId": core.MappingNodeFromString("vpc-12345678"),
				"mode":  core.MappingNodeFromString("create"),
				"networkAcls": core.MappingNodeItems(
					core.MappingNodeFields(
						"id", core.MappingNodeFromString("acl-12345678"),
					),
				),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when network ACL not found",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DescribeNetworkAcls": []any{
				&ec2.DescribeNetworkAclsInput{
					Filters: []types.Filter{
						{
							Name:   aws.String("default"),
							Values: []string{"true"},
						},
						{
							Name:   aws.String("vpc-id"),
							Values: []string{"vpc-12345678"},
						},
					},
				},
				&ec2.DescribeNetworkAclsInput{
					Filters: []types.Filter{
						{
							Name:   aws.String("network-acl-id"),
							Values: []string{"acl-12345678"},
						},
					},
				},
			},
		},
	}
}

func TestFlexVPCResourceDestroy(t *testing.T) {
	suite.Run(t, new(FlexVPCResourceDestroySuite))
}

// A create that fails partway records no computed fields, so the teardown steps
// have nothing to work from and the resources are left in the account. The VPC is
// located by its name tag instead and torn down from what discovery finds.
func createPartialCreateDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
			{
				Vpcs: []types.Vpc{
					{VpcId: aws.String("vpc-partial"), CidrBlock: aws.String("10.0.0.0/16")},
				},
			},
		}),
		ec2mock.WithDescribeVpcAttributeOutput(&ec2.DescribeVpcAttributeOutput{}),
		ec2mock.WithDescribeSubnetsOutput(&ec2.DescribeSubnetsOutput{}),
		ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{}),
		ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{}),
		ec2mock.WithDescribeNetworkAclsOutput(&ec2.DescribeNetworkAclsOutput{}),
		ec2mock.WithDescribeInternetGatewaysOutput(&ec2.DescribeInternetGatewaysOutput{
			InternetGateways: []types.InternetGateway{
				{InternetGatewayId: aws.String("igw-partial")},
			},
		}),
		ec2mock.WithDescribeNatGatewaysOutput(&ec2.DescribeNatGatewaysOutput{}),
		ec2mock.WithDeleteVpcOutput(&ec2.DeleteVpcOutput{}),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"mode": core.MappingNodeFromString("create"),
				"name": core.MappingNodeFromString("TestVPC"),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "locates the VPC by name when a partial create left no vpcId in state",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"DeleteVpc": &ec2.DeleteVpcInput{VpcId: aws.String("vpc-partial")},
		},
	}
}

// When the VPC cannot be found, nothing reached the target environment and the
// destroy is a no-op rather than an error.
func createPartialCreateNothingDeployedTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service] {
	service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{{Vpcs: []types.Vpc{}}}),
	)

	resourceState := &state.ResourceState{
		SpecData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"mode": core.MappingNodeFromString("create"),
				"name": core.MappingNodeFromString("TestVPC"),
			},
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ec2service.Service]{
		Name: "is a no-op when no VPC was created in the target environment",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState:   resourceState,
		},
		ExpectError:             false,
		DestroyActionsNotCalled: []string{"DeleteVpc", "DeleteSubnet", "DeleteSecurityGroup"},
	}
}
