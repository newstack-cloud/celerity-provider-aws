//go:build unit

package flex

import (
	"context"
	"fmt"
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
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

// IPv6 egress from a private subnet has to go through an egress-only internet gateway
// rather than the internet gateway.
//
// Both make outbound traffic work, so an end-to-end assertion on connectivity would pass
// either way. The difference is direction: an internet gateway is bidirectional for IPv6,
// so routing ::/0 through it leaves anything in the subnet holding an IPv6 address
// reachable from the internet. Only the shape of the route distinguishes them.
type FlexVPCIPv6RoutingSuite struct {
	suite.Suite
}

// Every default IPv6 route in the VPC, whichever subnet or route table it belongs to.
func defaultIPv6Routes(service *ec2mock.Ec2ServiceMock) []*ec2.CreateRouteInput {
	routes := []*ec2.CreateRouteInput{}
	for _, input := range service.CreateRouteInputs() {
		if aws.ToString(input.DestinationIpv6CidrBlock) == "::/0" {
			routes = append(routes, input)
		}
	}

	return routes
}

// Route tables whose IPv4 default goes to a NAT gateway, which is what distinguishes a
// private subnet's table from a public one's.
func privateSubnetRouteTables(service *ec2mock.Ec2ServiceMock) map[string]bool {
	tables := map[string]bool{}
	for _, route := range service.CreateRouteInputs() {
		if aws.ToString(route.DestinationCidrBlock) == "0.0.0.0/0" &&
			aws.ToString(route.NatGatewayId) != "" {
			tables[aws.ToString(route.RouteTableId)] = true
		}
	}

	return tables
}

func ipv6RoutingVPCSpec(preset string) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"mode":               core.MappingNodeFromString("create"),
			"preset":             core.MappingNodeFromString(preset),
			"name":               core.MappingNodeFromString("TestVPC"),
			"cidrBlock":          core.MappingNodeFromString("10.0.0.0/16"),
			"enableDNSSupport":   core.MappingNodeFromBool(true),
			"enableDNSHostnames": core.MappingNodeFromBool(true),
		},
	}
}

func ipv6RoutingDeployInput(
	providerCtx provider.Context,
	preset string,
) *provider.ResourceDeployInput {
	return &provider.ResourceDeployInput{
		InstanceID: "test-instance-id",
		ResourceID: "test-resource-id",
		Changes: &provider.Changes{
			AppliedResourceInfo: provider.ResourceInfo{
				ResourceID:   "test-resource-id",
				ResourceName: "TestVPC",
				InstanceID:   "test-instance-id",
				ResourceWithResolvedSubs: &provider.ResolvedResource{
					Type: &schema.ResourceTypeWrapper{Value: "aws/ec2/vpc"},
					Spec: ipv6RoutingVPCSpec(preset),
				},
			},
		},
		ProviderContext: providerCtx,
	}
}

func (s *FlexVPCIPv6RoutingSuite) Test_ipv6_routing() {
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
		privateSubnetIPv6RoutesThroughEgressOnlyGatewayTestCase(providerCtx, loader),
		isolatedPresetHasNoDefaultIPv6RouteTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		ipv6RoutingVPCResourceWrapper(),
		&s.Suite,
	)
}

func ipv6RoutingVPCResourceWrapper() func(
	serviceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
	configStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	return func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		taggingFactory := func(
			config *aws.Config,
			ctx provider.Context,
		) resgrouptagservice.Service {
			return &mockResourceGroupTaggingService{}
		}
		return VPCResource(serviceFactory, taggingFactory, configStore)
	}
}

// The standard preset has private subnets that reach the internet, so it provisions the
// egress-only gateway and routes ::/0 through it.
func privateSubnetIPv6RoutesThroughEgressOnlyGatewayTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := standardPresetEc2ServiceMock()

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "private subnet IPv6 egress goes through the egress-only internet gateway",
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
		Input:          ipv6RoutingDeployInput(providerCtx, "standard"),
		ExpectedOutput: standardPresetExpectedOutput(),
		ExtraAssertions: func(
			ctx context.Context,
			testSuite *suite.Suite,
			output *provider.ResourceDeployOutput,
		) {
			// Public subnets route ::/0 through the internet gateway quite correctly,
			// so the assertion has to be about the private ones. A route table whose
			// IPv4 default goes to a NAT gateway belongs to a private subnet.
			privateTables := privateSubnetRouteTables(service)
			testSuite.Require().NotEmpty(
				privateTables,
				"the standard preset has private subnets that reach the internet",
			)

			ipv6ByTable := map[string]*ec2.CreateRouteInput{}
			for _, route := range defaultIPv6Routes(service) {
				ipv6ByTable[aws.ToString(route.RouteTableId)] = route
			}

			for tableID := range privateTables {
				route, hasRoute := ipv6ByTable[tableID]
				testSuite.Require().Truef(
					hasRoute,
					"private subnet route table %s has no default IPv6 route", tableID,
				)
				testSuite.Equal(
					"eigw-12345678",
					aws.ToString(route.EgressOnlyInternetGatewayId),
					"the default IPv6 route must target the egress-only internet gateway",
				)
				testSuite.Empty(
					aws.ToString(route.GatewayId),
					"routing IPv6 through the internet gateway would make the subnet "+
						"inbound-reachable from the internet",
				)
			}
		},
		ExpectError: false,
	}
}

// A VPC with no route to the internet gets no default IPv6 route at all, so the isolated
// preset stays isolated in both address families rather than only in IPv4.
func isolatedPresetHasNoDefaultIPv6RouteTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service] {
	service := isolatedPresetEc2ServiceMock()

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ec2service.Service]{
		Name: "isolated preset creates no default IPv6 route",
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
		Input: ipv6RoutingDeployInput(providerCtx, "isolated"),
		ExpectedOutputMatcher: func(
			actual *provider.ResourceDeployOutput,
		) (plugintestutils.EqualityCheckValues, error) {
			// The VPC's own identity is enough here; the routes are what this case is
			// about and they are asserted below.
			return plugintestutils.EqualityCheckValues{
				Expected: core.MappingNodeFromString("vpc-12345678"),
				Actual:   actual.ComputedFieldValues["spec.vpcId"],
			}, nil
		},
		SaveActionsNotCalled: []string{
			// No route out means nothing to route through.
			"CreateEgressOnlyInternetGateway",
			"CreateInternetGateway",
			"CreateNatGateway",
		},
		ExtraAssertions: func(
			ctx context.Context,
			testSuite *suite.Suite,
			output *provider.ResourceDeployOutput,
		) {
			testSuite.Empty(
				defaultIPv6Routes(service),
				"an isolated VPC must not have a default IPv6 route",
			)
		},
		ExpectError: false,
	}
}

// The standard preset's EC2 mock. The preset has private subnets that reach the
// internet, which is what makes it the case where an egress-only gateway is expected.
func standardPresetEc2ServiceMock() *ec2mock.Ec2ServiceMock {
	return ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(standardPresetDescribeVPCMockOutputs()),
		ec2mock.WithCreateVpcOutput(standardPresetCreateVpcMockOutput()),
		ec2mock.WithModifyVpcAttributeOutput(&ec2.ModifyVpcAttributeOutput{}),
		ec2mock.WithDescribeAvailabilityZonesOutput(
			standardPresetDescribeAvailabilityZonesMockOutput(),
		),
		ec2mock.WithCreateSubnetOutputs(standardPresetCreateSubnetMockOutputs()),
		ec2mock.WithModifySubnetAttributeOutput(&ec2.ModifySubnetAttributeOutput{}),
		ec2mock.WithCreateInternetGatewayOutput(standardPresetCreateInternetGatewayMockOutput()),
		ec2mock.WithAttachInternetGatewayOutput(&ec2.AttachInternetGatewayOutput{}),
		ec2mock.WithCreateEgressOnlyInternetGatewayOutput(
			standardPresetCreateEgressOnlyInternetGatewayMockOutput(),
		),
		ec2mock.WithCreateRouteTableOutputs(standardPresetCreateRouteTableMockOutputs()),
		ec2mock.WithAssociateRouteTableOutput(&ec2.AssociateRouteTableOutput{}),
		ec2mock.WithCreateRouteOutput(&ec2.CreateRouteOutput{}),
		ec2mock.WithAllocateAddressOutputs(standardPresetAllocateAddressMockOutputs()),
		ec2mock.WithCreateNatGatewayOutputs(standardPresetCreateNatGatewayMockOutputs()),
		ec2mock.WithDescribeNatGatewaysOutput(standardPresetDescribeNatGatewaysMockOutput()),
		ec2mock.WithCreateSecurityGroupOutput(standardPresetCreateSecurityGroupMockOutput()),
		ec2mock.WithRevokeSecurityGroupEgressOutput(&ec2.RevokeSecurityGroupEgressOutput{}),
	)
}

// The isolated preset's EC2 mock: three private subnets, and nothing that could carry
// traffic out of the VPC.
func isolatedPresetEc2ServiceMock() *ec2mock.Ec2ServiceMock {
	subnetOutputs := []*ec2.CreateSubnetOutput{}
	routeTableOutputs := []*ec2.CreateRouteTableOutput{}
	for i := 1; i <= 3; i++ {
		subnetOutputs = append(subnetOutputs, &ec2.CreateSubnetOutput{
			Subnet: &types.Subnet{
				SubnetId: aws.String(fmt.Sprintf("subnet-isolated-%d", i)),
				VpcId:    aws.String("vpc-12345678"),
			},
		})
		routeTableOutputs = append(routeTableOutputs, &ec2.CreateRouteTableOutput{
			RouteTable: &types.RouteTable{
				RouteTableId: aws.String(fmt.Sprintf("rtb-isolated-%d", i)),
				VpcId:        aws.String("vpc-12345678"),
			},
		})
	}

	return ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(standardPresetDescribeVPCMockOutputs()),
		ec2mock.WithCreateVpcOutput(standardPresetCreateVpcMockOutput()),
		ec2mock.WithModifyVpcAttributeOutput(&ec2.ModifyVpcAttributeOutput{}),
		ec2mock.WithDescribeAvailabilityZonesOutput(
			standardPresetDescribeAvailabilityZonesMockOutput(),
		),
		ec2mock.WithCreateSubnetOutputs(subnetOutputs),
		ec2mock.WithModifySubnetAttributeOutput(&ec2.ModifySubnetAttributeOutput{}),
		ec2mock.WithCreateRouteTableOutputs(routeTableOutputs),
		ec2mock.WithAssociateRouteTableOutput(&ec2.AssociateRouteTableOutput{}),
		ec2mock.WithCreateRouteOutput(&ec2.CreateRouteOutput{}),
		ec2mock.WithCreateSecurityGroupOutput(standardPresetCreateSecurityGroupMockOutput()),
		ec2mock.WithRevokeSecurityGroupEgressOutput(&ec2.RevokeSecurityGroupEgressOutput{}),
	)
}

func TestFlexVPCIPv6RoutingSuite(t *testing.T) {
	suite.Run(t, new(FlexVPCIPv6RoutingSuite))
}
