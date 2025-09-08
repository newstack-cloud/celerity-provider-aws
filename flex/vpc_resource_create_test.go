//go:build unit

package flex

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	resgrouptagtypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FlexVPCResourceCreateSuite struct {
	suite.Suite
}

func (s *FlexVPCResourceCreateSuite) Test_create() {
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

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		createVPCCreateWithStandardPresetTestCase(providerCtx, loader),
		createVPCCreateWithMissingModeTestCase(providerCtx, loader),
		createVPCCreateWithMissingNameTestCase(providerCtx, loader),
		createVPCCreateWithMissingCIDRBlockTestCase(providerCtx, loader),
		createVPCCreateWithExistingVPCTestCase(providerCtx, loader),
		createVPCCreateWithCreateVPCErrorTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	vpcResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		// Create a mock resource group tagging service factory for testing
		mockResourceGroupTaggingServiceFactory := func(config *aws.Config, ctx provider.Context) resgrouptagservice.Service {
			// Return a mock service that returns empty results
			return &mockResourceGroupTaggingService{}
		}
		return VPCResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		vpcResourceWrapper,
		&s.Suite,
	)
}

func createVPCCreateWithStandardPresetTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	ec2Service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(standardPresetDescribeVPCMockOutputs()),
		ec2mock.WithCreateVpcOutput(standardPresetCreateVpcMockOutput()),
		// Only a single call for modifying attributes to enable
		// DNS host names (DNS support is enabled by default).
		ec2mock.WithModifyVpcAttributeOutput(&ec2.ModifyVpcAttributeOutput{}),
		// Retrieve availability zones for the region.
		ec2mock.WithDescribeAvailabilityZonesOutput(standardPresetDescribeAvailabilityZonesMockOutput()),
		// For the standard preset, 6 subnets are created.
		ec2mock.WithCreateSubnetOutputs(standardPresetCreateSubnetMockOutputs()),
		// Multiple calls will be made to modify subnet attributes to enable IPv6 address
		// assignment on creation and public IP address mapping for public subnets.
		ec2mock.WithModifySubnetAttributeOutput(&ec2.ModifySubnetAttributeOutput{}),
		// A single internet gateway will be created for the VPC.
		ec2mock.WithCreateInternetGatewayOutput(standardPresetCreateInternetGatewayMockOutput()),
		// The internet gateway will be attached to the VPC.
		ec2mock.WithAttachInternetGatewayOutput(&ec2.AttachInternetGatewayOutput{}),
		// A route table will be created for each subnet.
		ec2mock.WithCreateRouteTableOutputs(standardPresetCreateRouteTableMockOutputs()),
		// The route tables will be associated with the subnets.
		ec2mock.WithAssociateRouteTableOutput(&ec2.AssociateRouteTableOutput{}),
		// Routes will be created from public subnets to internet gateways
		// and from private subnets to NAT gateways.
		ec2mock.WithCreateRouteOutput(&ec2.CreateRouteOutput{}),
		// An elastic IP will be allocated for each NAT gateway
		// and a NAT gateway will be created for each private subnet
		// that needs access to the internet.
		ec2mock.WithAllocateAddressOutputs(standardPresetAllocateAddressMockOutputs()),
		ec2mock.WithCreateNatGatewayOutputs(standardPresetCreateNatGatewayMockOutputs()),
		// Describe NAT gateways is used for waiting for the NAT gateways to be available.
		ec2mock.WithDescribeNatGatewaysOutput(standardPresetDescribeNatGatewaysMockOutput()),
		// Creates initial security group.
		ec2mock.WithCreateSecurityGroupOutput(standardPresetCreateSecurityGroupMockOutput()),
		// Revokes egress rules from the security group.
		ec2mock.WithRevokeSecurityGroupEgressOutput(&ec2.RevokeSecurityGroupEgressOutput{}),
		// Standard preset is not a public-only VPC, so an initial NACL with deny-all
		// rules will not be created.
	)

	resourceSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"mode":               core.MappingNodeFromString("create"),
			"preset":             core.MappingNodeFromString("standard"),
			"name":               core.MappingNodeFromString("TestVPC"),
			"cidrBlock":          core.MappingNodeFromString("10.0.0.0/16"),
			"enableDNSSupport":   core.MappingNodeFromBool(true),
			"enableDNSHostnames": core.MappingNodeFromBool(true),
			"tags": core.MappingNodeFields(
				"System",
				core.MappingNodeFromString("Orders"),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "create VPC with standard preset",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return ec2Service
		},
		ServiceMockCalls: &ec2Service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-resource-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-resource-id",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: resourceSpecData,
					},
				},
			},
			ProviderContext: providerCtx,
		},
		// This is quite a complex operation, so we won't track all the individual API calls,
		// instead, we'll just check the outputs are correct, which implies the correct API calls
		// are made. The integrated test suite will provide additional assurances that all the resources
		// are provisioned correctly for a flex VPC.
		ExpectedOutput: standardPresetExpectedOutput(),
		ExpectError:    false,
	}
}

func createVPCCreateWithMissingModeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	ec2Service := ec2mock.CreateEc2ServiceMock()

	resourceSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":      core.MappingNodeFromString("TestVPC"),
			"cidrBlock": core.MappingNodeFromString("10.0.0.0/16"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when mode is missing",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return ec2Service
		},
		ServiceMockCalls: &ec2Service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-resource-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-resource-id",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: resourceSpecData,
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createVPCCreateWithMissingNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	ec2Service := ec2mock.CreateEc2ServiceMock()

	resourceSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"mode":      core.MappingNodeFromString("create"),
			"cidrBlock": core.MappingNodeFromString("10.0.0.0/16"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when name is missing",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return ec2Service
		},
		ServiceMockCalls: &ec2Service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-resource-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-resource-id",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: resourceSpecData,
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createVPCCreateWithMissingCIDRBlockTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	ec2Service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{{Vpcs: []types.Vpc{}}}),
	)

	resourceSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"mode": core.MappingNodeFromString("create"),
			"name": core.MappingNodeFromString("TestVPC"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when cidrBlock is missing",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return ec2Service
		},
		ServiceMockCalls: &ec2Service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-resource-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-resource-id",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: resourceSpecData,
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createVPCCreateWithExistingVPCTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	ec2Service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{
			{
				Vpcs: []types.Vpc{
					{
						VpcId: aws.String("vpc-existing"),
						Tags: []types.Tag{
							{
								Key:   aws.String(TagFlexVPCName),
								Value: aws.String("TestVPC"),
							},
						},
					},
				},
			},
		},
		),
	)

	resourceSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"mode":      core.MappingNodeFromString("create"),
			"name":      core.MappingNodeFromString("TestVPC"),
			"cidrBlock": core.MappingNodeFromString("10.0.0.0/16"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when VPC already exists",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return ec2Service
		},
		ServiceMockCalls: &ec2Service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-resource-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-resource-id",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: resourceSpecData,
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createVPCCreateWithCreateVPCErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	ec2Service := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs([]*ec2.DescribeVpcsOutput{{Vpcs: []types.Vpc{}}}),
		ec2mock.WithCreateVpcError(errors.New("failed to create VPC")),
	)

	resourceSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"mode":      core.MappingNodeFromString("create"),
			"name":      core.MappingNodeFromString("TestVPC"),
			"cidrBlock": core.MappingNodeFromString("10.0.0.0/16"),
			"region":    core.MappingNodeFromString("us-east-1"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "returns error when CreateVpc fails",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) ec2service.Service {
			return ec2Service
		},
		ServiceMockCalls: &ec2Service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-resource-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-resource-id",
					ResourceName: "TestVPC",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/ec2/vpc",
						},
						Spec: resourceSpecData,
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

// Mock resource group tagging service for testing.
type mockResourceGroupTaggingService struct{}

func (m *mockResourceGroupTaggingService) GetResources(ctx context.Context, input *resourcegroupstaggingapi.GetResourcesInput, optFns ...func(*resourcegroupstaggingapi.Options)) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
	return &resourcegroupstaggingapi.GetResourcesOutput{
		ResourceTagMappingList: []resgrouptagtypes.ResourceTagMapping{},
	}, nil
}

func standardPresetDescribeVPCMockOutputs() []*ec2.DescribeVpcsOutput {
	return []*ec2.DescribeVpcsOutput{
		{
			// No VPCs found when checking for existing VPC with the given name.
			Vpcs: []types.Vpc{},
		},
		{
			Vpcs: []types.Vpc{
				// Checking to see the status of the newly created VPC.
				{
					VpcId:     aws.String("vpc-12345678"),
					CidrBlock: aws.String("10.0.0.0/16"),
					State:     types.VpcStateAvailable,
					Tags: []types.Tag{
						{
							Key:   aws.String(TagFlexVPCName),
							Value: aws.String("TestVPC"),
						},
					},
				},
			},
		},
	}
}

func standardPresetCreateVpcMockOutput() *ec2.CreateVpcOutput {
	return &ec2.CreateVpcOutput{
		Vpc: &types.Vpc{
			VpcId:     aws.String("vpc-12345678"),
			CidrBlock: aws.String("10.0.0.0/16"),
			Ipv6CidrBlockAssociationSet: []types.VpcIpv6CidrBlockAssociation{
				{
					Ipv6CidrBlock: aws.String("2001:db8:1234:1a00::/56"),
				},
			},
			State: types.VpcStatePending,
			Tags: []types.Tag{
				{
					Key:   aws.String(TagFlexVPCName),
					Value: aws.String("TestVPC"),
				},
			},
		},
	}
}

func standardPresetDescribeAvailabilityZonesMockOutput() *ec2.DescribeAvailabilityZonesOutput {
	return &ec2.DescribeAvailabilityZonesOutput{
		AvailabilityZones: []types.AvailabilityZone{
			{
				ZoneName: aws.String("us-east-1a"),
			},
			{
				ZoneName: aws.String("us-east-1b"),
			},
			{
				ZoneName: aws.String("us-east-1c"),
			},
		},
	}
}

func standardPresetCreateSubnetMockOutputs() []*ec2.CreateSubnetOutput {
	return []*ec2.CreateSubnetOutput{
		{
			Subnet: &types.Subnet{
				SubnetId:  aws.String("subnet-1"),
				CidrBlock: aws.String("10.0.0.0/19"),
				Ipv6CidrBlockAssociationSet: []types.SubnetIpv6CidrBlockAssociation{
					{
						Ipv6CidrBlock: aws.String("2001:db8:1234:1a00::/64"),
					},
				},
				State:            types.SubnetStateAvailable,
				AvailabilityZone: aws.String("us-east-1a"),
				Tags: []types.Tag{
					{
						Key:   aws.String(TagFlexVPCName),
						Value: aws.String("TestVPC"),
					},
				},
			},
		},
		{
			Subnet: &types.Subnet{
				SubnetId:  aws.String("subnet-2"),
				CidrBlock: aws.String("10.0.32.0/19"),
				Ipv6CidrBlockAssociationSet: []types.SubnetIpv6CidrBlockAssociation{
					{
						Ipv6CidrBlock: aws.String("2001:db8:1234:1a01::/64"),
					},
				},
				State:            types.SubnetStateAvailable,
				AvailabilityZone: aws.String("us-east-1b"),
				Tags: []types.Tag{
					{
						Key:   aws.String(TagFlexVPCName),
						Value: aws.String("TestVPC"),
					},
				},
			},
		},
		{
			Subnet: &types.Subnet{
				SubnetId:  aws.String("subnet-3"),
				CidrBlock: aws.String("10.0.64.0/19"),
				Ipv6CidrBlockAssociationSet: []types.SubnetIpv6CidrBlockAssociation{
					{
						Ipv6CidrBlock: aws.String("2001:db8:1234:1a02::/64"),
					},
				},
				State:            types.SubnetStateAvailable,
				AvailabilityZone: aws.String("us-east-1c"),
				Tags: []types.Tag{
					{
						Key:   aws.String(TagFlexVPCName),
						Value: aws.String("TestVPC"),
					},
				},
			},
		},
		{
			Subnet: &types.Subnet{
				SubnetId:  aws.String("subnet-4"),
				CidrBlock: aws.String("10.0.96.0/19"),
				Ipv6CidrBlockAssociationSet: []types.SubnetIpv6CidrBlockAssociation{
					{
						Ipv6CidrBlock: aws.String("2001:db8:1234:1a03::/64"),
					},
				},
				State:            types.SubnetStateAvailable,
				AvailabilityZone: aws.String("us-east-1a"),
				Tags: []types.Tag{
					{
						Key:   aws.String(TagFlexVPCName),
						Value: aws.String("TestVPC"),
					},
				},
			},
		},
		{
			Subnet: &types.Subnet{
				SubnetId:  aws.String("subnet-5"),
				CidrBlock: aws.String("10.0.128.0/19"),
				Ipv6CidrBlockAssociationSet: []types.SubnetIpv6CidrBlockAssociation{
					{
						Ipv6CidrBlock: aws.String("2001:db8:1234:1a04::/64"),
					},
				},
				State:            types.SubnetStateAvailable,
				AvailabilityZone: aws.String("us-east-1b"),
				Tags: []types.Tag{
					{
						Key:   aws.String(TagFlexVPCName),
						Value: aws.String("TestVPC"),
					},
				},
			},
		},
		{
			Subnet: &types.Subnet{
				SubnetId:  aws.String("subnet-6"),
				CidrBlock: aws.String("10.0.160.0/19"),
				Ipv6CidrBlockAssociationSet: []types.SubnetIpv6CidrBlockAssociation{
					{
						Ipv6CidrBlock: aws.String("2001:db8:1234:1a05::/64"),
					},
				},
				State:            types.SubnetStateAvailable,
				AvailabilityZone: aws.String("us-east-1c"),
				Tags: []types.Tag{
					{
						Key:   aws.String(TagFlexVPCName),
						Value: aws.String("TestVPC"),
					},
				},
			},
		},
	}
}

func standardPresetCreateInternetGatewayMockOutput() *ec2.CreateInternetGatewayOutput {
	return &ec2.CreateInternetGatewayOutput{
		InternetGateway: &types.InternetGateway{
			InternetGatewayId: aws.String("igw-12345678"),
			Tags: []types.Tag{
				{
					Key:   aws.String(TagFlexVPCName),
					Value: aws.String("TestVPC"),
				},
			},
		},
	}
}

func standardPresetCreateRouteTableMockOutputs() []*ec2.CreateRouteTableOutput {
	routeTableOutputs := []*ec2.CreateRouteTableOutput{}

	// 1 route table for each subnet.
	for i := 1; i <= 6; i++ {
		routeTableOutputs = append(routeTableOutputs, &ec2.CreateRouteTableOutput{
			RouteTable: &types.RouteTable{
				RouteTableId: aws.String(fmt.Sprintf("rtb-%d", i)),
				VpcId:        aws.String("vpc-12345678"),
				Tags: []types.Tag{
					{
						Key:   aws.String(TagFlexVPCName),
						Value: aws.String("TestVPC"),
					},
				},
			},
		})
	}

	return routeTableOutputs
}

func standardPresetAllocateAddressMockOutputs() []*ec2.AllocateAddressOutput {
	elasticIPOutputs := []*ec2.AllocateAddressOutput{}

	// 3 elastic IPs are allocated for the NAT gateways
	// used to allow private subnets to access the internet.
	// With the standard preset, there are 3 private subnets.
	for i := 1; i <= 3; i++ {
		elasticIPOutputs = append(elasticIPOutputs, &ec2.AllocateAddressOutput{
			AllocationId: aws.String(fmt.Sprintf("eipalloc-%d", i)),
		})
	}

	return elasticIPOutputs
}

func standardPresetCreateNatGatewayMockOutputs() []*ec2.CreateNatGatewayOutput {
	natGatewayOutputs := []*ec2.CreateNatGatewayOutput{}

	// 3 NAT gateways are created for the private subnets
	// that need access to the internet.
	// With the standard preset, there are 3 private subnets.
	for i := 1; i <= 3; i++ {
		natGatewayOutputs = append(natGatewayOutputs, &ec2.CreateNatGatewayOutput{
			NatGateway: &types.NatGateway{
				NatGatewayId: aws.String(fmt.Sprintf("nat-%d", i)),
				State:        types.NatGatewayStateAvailable,
				Tags: []types.Tag{
					{
						Key:   aws.String(TagFlexVPCName),
						Value: aws.String("TestVPC"),
					},
				},
			},
		})
	}

	return natGatewayOutputs
}

func standardPresetDescribeNatGatewaysMockOutput() *ec2.DescribeNatGatewaysOutput {
	// Any NAT gateway will be available on first check.
	return &ec2.DescribeNatGatewaysOutput{
		NatGateways: []types.NatGateway{
			{
				State: types.NatGatewayStateAvailable,
			},
		},
	}
}

func standardPresetCreateSecurityGroupMockOutput() *ec2.CreateSecurityGroupOutput {
	return &ec2.CreateSecurityGroupOutput{
		GroupId: aws.String("sg-12345678"),
	}
}

func standardPresetExpectedOutput() *provider.ResourceDeployOutput {
	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.vpcId": core.MappingNodeFromString("vpc-12345678"),
			"spec.subnets": core.MappingNodeFields(
				"private-az-1",
				core.MappingNodeFields(
					"id",
					core.MappingNodeFromString("subnet-1"),
					"availabilityZone",
					core.MappingNodeFromString("us-east-1a"),
				),
				"private-az-2",
				core.MappingNodeFields(
					"id",
					core.MappingNodeFromString("subnet-2"),
					"availabilityZone",
					core.MappingNodeFromString("us-east-1b"),
				),
				"private-az-3",
				core.MappingNodeFields(
					"id",
					core.MappingNodeFromString("subnet-3"),
					"availabilityZone",
					core.MappingNodeFromString("us-east-1c"),
				),
				"public-az-1",
				core.MappingNodeFields(
					"id",
					core.MappingNodeFromString("subnet-4"),
					"availabilityZone",
					core.MappingNodeFromString("us-east-1a"),
				),
				"public-az-2",
				core.MappingNodeFields(
					"id",
					core.MappingNodeFromString("subnet-5"),
					"availabilityZone",
					core.MappingNodeFromString("us-east-1b"),
				),
				"public-az-3",
				core.MappingNodeFields(
					"id",
					core.MappingNodeFromString("subnet-6"),
					"availabilityZone",
					core.MappingNodeFromString("us-east-1c"),
				),
			),
			"spec.routeTables": core.MappingNodeItems(
				// Route tables are created for public subnets
				// before private subnets.
				// Private subnets are created before public subnets,
				// so private subnets are 1, 2, 3 and public subnets are 4, 5, 6.
				core.MappingNodeFields(
					"id",
					core.MappingNodeFromString("rtb-1"),
					"subnetIds",
					core.MappingNodeItems(
						core.MappingNodeFromString("subnet-4"),
					),
				),
				core.MappingNodeFields(
					"id",
					core.MappingNodeFromString("rtb-2"),
					"subnetIds",
					core.MappingNodeItems(
						core.MappingNodeFromString("subnet-5"),
					),
				),
				core.MappingNodeFields(
					"id",
					core.MappingNodeFromString("rtb-3"),
					"subnetIds",
					core.MappingNodeItems(
						core.MappingNodeFromString("subnet-6"),
					),
				),
				core.MappingNodeFields(
					"id",
					core.MappingNodeFromString("rtb-4"),
					"subnetIds",
					core.MappingNodeItems(
						core.MappingNodeFromString("subnet-1"),
					),
				),
				core.MappingNodeFields(
					"id",
					core.MappingNodeFromString("rtb-5"),
					"subnetIds",
					core.MappingNodeItems(
						core.MappingNodeFromString("subnet-2"),
					),
				),
				core.MappingNodeFields(
					"id",
					core.MappingNodeFromString("rtb-6"),
					"subnetIds",
					core.MappingNodeItems(
						core.MappingNodeFromString("subnet-3"),
					),
				),
			),
			"spec.securityGroups": core.MappingNodeItems(
				core.MappingNodeFromString("sg-12345678"),
			),
			"spec.networkAcls": core.MappingNodeItems(),
			"spec.gateways": core.MappingNodeFields(
				"internetGatewayId",
				core.MappingNodeFromString("igw-12345678"),
				"natGateways",
				core.MappingNodeItems(
					core.MappingNodeFields(
						"id",
						core.MappingNodeFromString("nat-1"),
						"elasticIpId",
						core.MappingNodeFromString("eipalloc-1"),
						"inPublicSubnetId",
						core.MappingNodeFromString("subnet-4"),
						"forPrivateSubnetId",
						core.MappingNodeFromString("subnet-1"),
					),
					core.MappingNodeFields(
						"id",
						core.MappingNodeFromString("nat-2"),
						"elasticIpId",
						core.MappingNodeFromString("eipalloc-2"),
						"inPublicSubnetId",
						core.MappingNodeFromString("subnet-5"),
						"forPrivateSubnetId",
						core.MappingNodeFromString("subnet-2"),
					),
					core.MappingNodeFields(
						"id",
						core.MappingNodeFromString("nat-3"),
						"elasticIpId",
						core.MappingNodeFromString("eipalloc-3"),
						"inPublicSubnetId",
						core.MappingNodeFromString("subnet-6"),
						"forPrivateSubnetId",
						core.MappingNodeFromString("subnet-3"),
					),
				),
			),
		},
	}
}

func TestFlexVPCResourceCreate(t *testing.T) {
	suite.Run(t, new(FlexVPCResourceCreateSuite))
}
