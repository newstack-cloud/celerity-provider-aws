//go:build unit

package linkutils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

func interfaceEndpointActivation() NetworkingActivation {
	return NetworkingActivation{
		Caller: CallerNetworking{
			VPCID:            "vpc-1",
			SubnetIDs:        []string{"subnet-1"},
			SecurityGroupIDs: []string{"sg-caller"},
		},
		Region:     "us-west-2",
		AWSService: "secretsmanager",
	}
}

// Creating an interface endpoint must pair the caller with the endpoint's security
// group, referencing it by ID.
func (s *ReconcileLinkNetworkingSuite) Test_interface_endpoint_references_caller_group_by_id() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{}),
		ec2mock.WithCreateSecurityGroupOutput(&ec2.CreateSecurityGroupOutput{
			GroupId: aws.String("sg-endpoint"),
		}),
		ec2mock.WithAuthorizeSecurityGroupIngressOutput(&ec2.AuthorizeSecurityGroupIngressOutput{}),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
		ec2mock.WithCreateVpcEndpointOutput(&ec2.CreateVpcEndpointOutput{
			VpcEndpoint: &ec2types.VpcEndpoint{VpcEndpointId: aws.String("vpce-1")},
		}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		interfaceEndpointActivation(),
		output,
	)
	s.Require().NoError(err)

	ec2Svc.AssertCalledWith(&s.Suite, "AuthorizeSecurityGroupIngress", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.AuthorizeSecurityGroupIngressInput)
		if !ok || len(in.IpPermissions) != 1 {
			return false
		}
		perm := in.IpPermissions[0]
		return aws.ToString(in.GroupId) == "sg-endpoint" &&
			// The field that made the old call malformed must not be set at all.
			in.SourceSecurityGroupName == nil &&
			len(perm.UserIdGroupPairs) == 1 &&
			aws.ToString(perm.UserIdGroupPairs[0].GroupId) == "sg-caller"
	})
}

func (s *ReconcileLinkNetworkingSuite) Test_interface_endpoint_opens_caller_egress() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{}),
		ec2mock.WithCreateSecurityGroupOutput(&ec2.CreateSecurityGroupOutput{
			GroupId: aws.String("sg-endpoint"),
		}),
		ec2mock.WithAuthorizeSecurityGroupIngressOutput(&ec2.AuthorizeSecurityGroupIngressOutput{}),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
		ec2mock.WithCreateVpcEndpointOutput(&ec2.CreateVpcEndpointOutput{
			VpcEndpoint: &ec2types.VpcEndpoint{VpcEndpointId: aws.String("vpce-1")},
		}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		interfaceEndpointActivation(),
		output,
	)
	s.Require().NoError(err)

	ec2Svc.AssertCalledWith(&s.Suite, "AuthorizeSecurityGroupEgress", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.AuthorizeSecurityGroupEgressInput)
		if !ok || len(in.IpPermissions) != 1 {
			return false
		}
		perm := in.IpPermissions[0]
		return aws.ToString(in.GroupId) == "sg-caller" &&
			len(perm.UserIdGroupPairs) == 1 &&
			aws.ToString(perm.UserIdGroupPairs[0].GroupId) == "sg-endpoint" &&
			aws.ToInt32(perm.FromPort) == 443 &&
			hasLinkRuleTag(in.TagSpecifications)
	})
}

// A gateway endpoint has no security group to pair with, so the caller's egress is
// opened to the service's AWS-managed prefix list instead. Without it the route to the
// gateway exists and the security group drops the traffic.
func (s *ReconcileLinkNetworkingSuite) Test_gateway_endpoint_opens_caller_egress_to_prefix_list() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
			RouteTables: []ec2types.RouteTable{
				{
					RouteTableId: aws.String("rtb-1"),
					Associations: []ec2types.RouteTableAssociation{
						{SubnetId: aws.String("subnet-1")},
					},
				},
			},
		}),
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{}),
		ec2mock.WithCreateVpcEndpointOutput(&ec2.CreateVpcEndpointOutput{
			VpcEndpoint: &ec2types.VpcEndpoint{VpcEndpointId: aws.String("vpce-gw-1")},
		}),
		ec2mock.WithDescribeManagedPrefixListsOutput(gatewayPrefixListOutput()),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	activation := interfaceEndpointActivation()
	activation.AWSService = "s3"
	activation.EndpointType = ec2types.VpcEndpointTypeGateway

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		activation,
		output,
	)
	s.Require().NoError(err)

	ec2Svc.AssertCalledWith(&s.Suite, "AuthorizeSecurityGroupEgress", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.AuthorizeSecurityGroupEgressInput)
		if !ok || len(in.IpPermissions) != 1 {
			return false
		}
		perm := in.IpPermissions[0]
		return aws.ToString(in.GroupId) == "sg-caller" &&
			len(perm.PrefixListIds) == 1 &&
			aws.ToString(perm.PrefixListIds[0].PrefixListId) == "pl-s3" &&
			hasLinkRuleTag(in.TagSpecifications)
	})
}

// A teardown must not provision the very endpoint it is tearing down. With no endpoint
// left to remove the destroy path has to stop rather than fall through to create.
func (s *ReconcileLinkNetworkingSuite) Test_gateway_endpoint_destroy_creates_nothing_when_already_gone() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
			RouteTables: []ec2types.RouteTable{
				{
					RouteTableId: aws.String("rtb-1"),
					Associations: []ec2types.RouteTableAssociation{
						{SubnetId: aws.String("subnet-1")},
					},
				},
			},
		}),
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{}),
		ec2mock.WithDescribeSecurityGroupRulesOutput(&ec2.DescribeSecurityGroupRulesOutput{}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	activation := interfaceEndpointActivation()
	activation.AWSService = "s3"
	activation.EndpointType = ec2types.VpcEndpointTypeGateway

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeDestroy, rs),
		activation,
		output,
	)
	s.Require().NoError(err)
	ec2Svc.AssertNotCalled(&s.Suite, "CreateVpcEndpoint")
	ec2Svc.AssertNotCalled(&s.Suite, "AuthorizeSecurityGroupEgress")
}

func hasLinkRuleTag(specs []ec2types.TagSpecification) bool {
	expected := utils.TagBlueprintLinkIDPrefix + testNetworkingLinkID
	for _, spec := range specs {
		if spec.ResourceType != ec2types.ResourceTypeSecurityGroupRule {
			continue
		}
		for _, tag := range spec.Tags {
			if aws.ToString(tag.Key) == expected {
				return true
			}
		}
	}

	return false
}

// An interface endpoint teardown must not provision the endpoint it is removing.
//
// The destroy branch was guarded on having found an endpoint, so a teardown that found
// none fell straight through to the create path. The gateway path had the same defect
// and was fixed earlier; this one was missed.
func (s *ReconcileLinkNetworkingSuite) Test_interface_endpoint_destroy_creates_nothing_when_already_gone() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{}),
		// No group left for this link either.
		ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeDestroy, rs),
		interfaceEndpointActivation(),
		output,
	)
	s.Require().NoError(err)
	ec2Svc.AssertNotCalled(&s.Suite, "CreateVpcEndpoint")
	ec2Svc.AssertNotCalled(&s.Suite, "CreateSecurityGroup")
}

// On destroy the caller may already have been detached by the placement link, since
// nothing orders the two links against each other.
//
// Returning early then leaves the endpoint and its security group in place, and that
// group's ingress rule referencing the caller's group blocks the caller group, and with
// it the entire VPC, from ever being deleted. This is what failed two end-to-end runs.
func (s *ReconcileLinkNetworkingSuite) Test_destroy_removes_recorded_endpoint_when_caller_already_detached() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{
			VpcEndpoints: []ec2types.VpcEndpoint{
				{
					VpcEndpointId: aws.String("vpce-recorded"),
					VpcId:         aws.String("vpc-1"),
					ServiceName:   aws.String("com.amazonaws.us-west-2.sqs"),
					Tags: []ec2types.Tag{
						utils.CreateTagBlueprintLinkID(testNetworkingLinkID),
					},
				},
			},
		}),
		ec2mock.WithDeleteVpcEndpointsOutput(&ec2.DeleteVpcEndpointsOutput{}),
		ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{}),
	)
	rs := resourceservicemock.Create()

	input := networkingInput(provider.LinkUpdateTypeDestroy, rs)
	input.CurrentLinkState = &state.LinkState{
		Data: map[string]*core.MappingNode{
			"callerVPCEndpoint": core.MappingNodeFields(
				"id", core.MappingNodeFromString("vpce-recorded"),
			),
		},
	}

	// The caller carries no VPC attachment: the placement link has already cleared it.
	detached := NetworkingActivation{Region: "us-west-2", AWSService: "sqs"}

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(context.Background(), ec2Svc, input, detached, output)
	s.Require().NoError(err)

	ec2Svc.AssertCalledWith(&s.Suite, "DeleteVpcEndpoints", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.DeleteVpcEndpointsInput)
		return ok && len(in.VpcEndpointIds) == 1 && in.VpcEndpointIds[0] == "vpce-recorded"
	})
}

// A detached caller with nothing recorded has nothing to clean up, and must not fail
// the teardown.
func (s *ReconcileLinkNetworkingSuite) Test_destroy_is_a_noop_for_a_detached_caller_with_no_recorded_endpoint() {
	ec2Svc := ec2mock.CreateEc2ServiceMock()
	rs := resourceservicemock.Create()

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeDestroy, rs),
		NetworkingActivation{Region: "us-west-2", AWSService: "sqs"},
		output,
	)
	s.Require().NoError(err)
	ec2Svc.AssertNotCalled(&s.Suite, "DeleteVpcEndpoints")
}

// An endpoint that is already gone does not mean its security group is.
//
// Deleting a VPC endpoint returns before AWS releases the interfaces it created, so the
// group delete can yield and be retried after the endpoint no longer exists. Returning
// early on "no endpoint found" left the group behind permanently, and its ingress rule
// then blocked the caller's group and the whole VPC from being deleted.
func (s *ReconcileLinkNetworkingSuite) Test_destroy_removes_the_endpoint_group_even_when_the_endpoint_is_gone() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		// The endpoint has already been removed by an earlier attempt.
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{}),
		ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []ec2types.SecurityGroup{
				{
					GroupId: aws.String("sg-endpoint"),
					Tags: []ec2types.Tag{
						utils.CreateTagBlueprintLinkID(testNetworkingLinkID),
					},
				},
			},
		}),
		ec2mock.WithDescribeNetworkInterfacesOutputs([]*ec2.DescribeNetworkInterfacesOutput{{}}),
		ec2mock.WithDeleteSecurityGroupOutput(&ec2.DeleteSecurityGroupOutput{}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeDestroy, rs),
		interfaceEndpointActivation(),
		output,
	)
	s.Require().NoError(err)

	ec2Svc.AssertCalledWith(&s.Suite, "DeleteSecurityGroup", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.DeleteSecurityGroupInput)
		return ok && aws.ToString(in.GroupId) == "sg-endpoint"
	})
	// Still must not provision an endpoint during a teardown.
	ec2Svc.AssertNotCalled(&s.Suite, "CreateVpcEndpoint")
}
