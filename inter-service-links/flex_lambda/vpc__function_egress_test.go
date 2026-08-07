//go:build unit

package flexlambda

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func vpcStateWithGateways(internetGatewayID string, natGatewayIDs ...string) *core.MappingNode {
	natGateways := &core.MappingNode{Items: []*core.MappingNode{}}
	for _, id := range natGatewayIDs {
		natGateways.Items = append(
			natGateways.Items,
			core.MappingNodeFields("id", core.MappingNodeFromString(id)),
		)
	}

	gateways := core.MappingNodeFields("natGateways", natGateways)
	if internetGatewayID != "" {
		gateways.Fields["internetGatewayId"] = core.MappingNodeFromString(internetGatewayID)
	}

	return core.MappingNodeFields("gateways", gateways)
}

func functionWithEgressAnnotation(value string) *provider.ResourceInfo {
	info := functionResourceInfoB()
	info.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					egressAnnotationKey: core.MappingNodeFromString(value),
				},
			},
		},
	}
	return info
}

// Left unset, outbound access follows the VPC's topology. Placing a function in a VPC
// must not silently remove access it had before, so unset resolves to the best the
// topology can deliver rather than to none.
func TestEgressDefaultsToWhatTheTopologyProvides(t *testing.T) {
	cases := []struct {
		name       string
		vpcState   *core.MappingNode
		subnetType string
		expected   egressReach
	}{
		{
			name:       "private subnet with a NAT gateway reaches both address families",
			vpcState:   vpcStateWithGateways("igw-1", "nat-1"),
			subnetType: "private",
			expected:   egressFull,
		},
		{
			name:       "private subnet with no NAT gateway is the isolated preset and reaches nothing",
			vpcState:   vpcStateWithGateways(""),
			subnetType: "private",
			expected:   egressNone,
		},
		{
			name:       "public subnet reaches IPv6 only, since a function has no public IPv4 address",
			vpcState:   vpcStateWithGateways("igw-1"),
			subnetType: "public",
			expected:   egressIPv6Only,
		},
		{
			name:       "public subnet with no internet gateway reaches nothing",
			vpcState:   vpcStateWithGateways(""),
			subnetType: "public",
			expected:   egressNone,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			plan, err := resolveEgressPlan(
				functionResourceInfoB(),
				testCase.vpcState,
				testCase.subnetType,
			)
			require.NoError(t, err)
			require.Equal(t, testCase.expected, plan.reach)
			require.Empty(t, plan.cidrs)
		})
	}
}

// The annotation only ever narrows, so "none" closes egress even where the topology
// would provide it.
func TestEgressAnnotationNoneClosesOutboundAccess(t *testing.T) {
	plan, err := resolveEgressPlan(
		functionWithEgressAnnotation("none"),
		vpcStateWithGateways("igw-1", "nat-1"),
		"private",
	)
	require.NoError(t, err)
	require.Equal(t, egressNone, plan.reach)
	require.Nil(t, egressPermissions(plan))
}

// Asking for something the VPC cannot deliver is reported rather than turned into a
// rule that grants nothing and fails at runtime as a timeout.
func TestEgressAnnotationRejectedWhenTopologyCannotDeliverIt(t *testing.T) {
	for _, value := range []string{"internet", "10.0.0.0/8"} {
		t.Run(value, func(t *testing.T) {
			_, err := resolveEgressPlan(
				functionWithEgressAnnotation(value),
				vpcStateWithGateways(""),
				"private",
			)
			require.Error(t, err)
			require.Contains(t, err.Error(), "no outbound path")
		})
	}
}

func TestEgressAnnotationRejectsUnrecognisedValues(t *testing.T) {
	_, err := resolveEgressPlan(
		functionWithEgressAnnotation("everywhere"),
		vpcStateWithGateways("igw-1", "nat-1"),
		"private",
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "comma-separated list of CIDR ranges")
}

func TestEgressPermissionsForFullInternet(t *testing.T) {
	permissions := egressPermissions(&egressPlan{reach: egressFull})
	require.Len(t, permissions, 1)
	require.Equal(t, "-1", aws.ToString(permissions[0].IpProtocol))
	require.Len(t, permissions[0].IpRanges, 1)
	require.Equal(t, "0.0.0.0/0", aws.ToString(permissions[0].IpRanges[0].CidrIp))
	require.Len(t, permissions[0].Ipv6Ranges, 1)
	require.Equal(t, "::/0", aws.ToString(permissions[0].Ipv6Ranges[0].CidrIpv6))
}

// On a public subnet the function has no public IPv4 address, so an IPv4 rule would
// be a grant that never carries traffic. Only IPv6 is opened.
func TestEgressPermissionsForIPv6OnlyOmitIPv4(t *testing.T) {
	permissions := egressPermissions(&egressPlan{reach: egressIPv6Only})
	require.Len(t, permissions, 1)
	require.Empty(t, permissions[0].IpRanges)
	require.Len(t, permissions[0].Ipv6Ranges, 1)
	require.Equal(t, "::/0", aws.ToString(permissions[0].Ipv6Ranges[0].CidrIpv6))
}

func TestEgressPermissionsForDeclaredCIDRs(t *testing.T) {
	plan, err := resolveEgressPlan(
		functionWithEgressAnnotation("10.0.0.0/8, 2001:db8::/32"),
		vpcStateWithGateways("igw-1", "nat-1"),
		"private",
	)
	require.NoError(t, err)
	require.Equal(t, []string{"10.0.0.0/8", "2001:db8::/32"}, plan.cidrs)

	permissions := egressPermissions(plan)
	require.Len(t, permissions, 1)
	require.Equal(t, "10.0.0.0/8", aws.ToString(permissions[0].IpRanges[0].CidrIp))
	require.Equal(t, "2001:db8::/32", aws.ToString(permissions[0].Ipv6Ranges[0].CidrIpv6))
}

// A declared IPv4 range on a subnet that can only reach IPv6 is dropped rather than
// authorised, and if that leaves nothing then no rule is written at all.
func TestEgressDeclaredIPv4DroppedWhenOnlyIPv6IsReachable(t *testing.T) {
	plan, err := resolveEgressPlan(
		functionWithEgressAnnotation("10.0.0.0/8"),
		vpcStateWithGateways("igw-1"),
		"public",
	)
	require.NoError(t, err)
	require.Equal(t, egressIPv6Only, plan.reach)
	require.Nil(t, egressPermissions(plan))
}

// The egress plan has to reach the security group, not just be computed correctly.
//
// Everything above tests resolveEgressPlan and egressPermissions directly, because the
// input space is combinatorial and a full placement per combination would cost far more
// than it proves. That leaves one thing those tests cannot show: that the plan is
// applied at all. Without this, removing the authorise call from the placement link
// would keep every egress test passing and silently leave placed functions unable to
// reach anything.
//
// One combination is enough. Its job is to prove the wire is connected, not to cover the
// matrix a second time.
func TestPlacementAuthorisesTheResolvedEgressOnTheWorkloadGroup(t *testing.T) {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{}),
		ec2mock.WithCreateSecurityGroupOutput(&ec2.CreateSecurityGroupOutput{
			GroupId: aws.String("sg-workload"),
		}),
		ec2mock.WithRevokeSecurityGroupEgressOutput(&ec2.RevokeSecurityGroupEgressOutput{}),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(vfGetFunctionOutput()),
		lambdamock.WithUpdateFunctionConfigurationOutput(nil),
	)
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	link := placementLinkWithEC2(iamSvc, lambdaSvc, ec2Svc)

	// A private subnet behind a NAT gateway, which is the topology that resolves to
	// full egress.
	input := &provider.LinkUpdateResourceInput{
		LinkUpdateType:    provider.LinkUpdateTypeCreate,
		ResourceInfo:      functionResourceInfoB(),
		OtherResourceInfo: flexVPCResourceInfoAWithGateways("igw-1", "nat-1"),
		LinkContext:       testLinkContext(),
		LinkID:            vfLinkID,
		ResourceService:   vfRoleService(),
	}

	_, err := link.UpdateResourceB(context.Background(), input)
	require.NoError(t, err)

	ec2Svc.AssertCalledWith(
		testSuite(t),
		"AuthorizeSecurityGroupEgress",
		0,
		plugintestutils.Any,
		func(arg any) bool {
			in, ok := arg.(*ec2.AuthorizeSecurityGroupEgressInput)
			if !ok || aws.ToString(in.GroupId) != "sg-workload" {
				return false
			}
			return len(in.IpPermissions) > 0
		},
	)
}

// The placement link built through its exported constructor, with the EC2 service
// supplied so the test can inspect what it authorised.
func placementLinkWithEC2(
	iamSvc iamservice.Service,
	lambdaSvc lambdaservice.Service,
	ec2Svc ec2service.Service,
) provider.Link {
	loader := &testutils.MockAWSConfigLoader{}
	build := vpcFunctionLinkFactory(iamSvc)

	return build(pluginutils.LinkServiceDeps[
		*aws.Config, ec2service.Service, *aws.Config, lambdaservice.Service,
	]{
		ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, ec2service.Service]{
			ServiceFactory: func(c *aws.Config, pc provider.Context) ec2service.Service {
				return ec2Svc
			},
			ConfigStore: testConfigStore(loader),
		},
		ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, lambdaservice.Service]{
			ServiceFactory: func(c *aws.Config, pc provider.Context) lambdaservice.Service {
				return lambdaSvc
			},
			ConfigStore: testConfigStore(loader),
		},
	})
}

// The shared VPC fixture declares no gateways, so it resolves to no egress at all. This
// adds the topology that gives a private subnet a route out.
func flexVPCResourceInfoAWithGateways(
	internetGatewayID string,
	natGatewayIDs ...string,
) *provider.ResourceInfo {
	info := flexVPCResourceInfoA()
	gateways := vpcStateWithGateways(internetGatewayID, natGatewayIDs...)
	info.CurrentResourceState.SpecData.Fields["gateways"] = gateways.Fields["gateways"]

	return info
}

// The mock's call assertions take a testify suite; these tests are plain functions.
func testSuite(t *testing.T) *suite.Suite {
	s := new(suite.Suite)
	s.SetT(t)
	return s
}
