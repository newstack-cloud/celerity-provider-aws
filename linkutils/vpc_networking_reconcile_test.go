//go:build unit

package linkutils

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink-provider-aws/flex"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type ReconcileLinkNetworkingSuite struct {
	suite.Suite
}

func testNetworkingLinkContext() provider.LinkContext {
	return plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {"region": core.ScalarFromString("us-west-2")},
		},
		map[string]*core.ScalarValue{},
	)
}

func flexVPCStateForNetworking() *state.ResourceState {
	return &state.ResourceState{
		Name: "appVpc",
		SpecData: core.MappingNodeFields(
			"name", core.MappingNodeFromString("orders-vpc"),
			"enableDNSSupport", core.MappingNodeFromBool(true),
			"enableDNSHostnames", core.MappingNodeFromBool(true),
			// The group the VPC provisions for the target. A link pairs against one of
			// these rather than against whatever group the target happens to list first.
			"securityGroupIdsByName", core.MappingNodeFields(
				"db", core.MappingNodeFromString("sg-db"),
			),
		),
	}
}

// The link ID every networking test runs under; rules are tagged with it and revoked
// by it.
const testNetworkingLinkID = "link-1"

func networkingInput(
	updateType provider.LinkUpdateType,
	resourceService provider.ResourceService,
) *provider.LinkUpdateIntermediaryResourcesInput {
	return &provider.LinkUpdateIntermediaryResourcesInput{
		LinkID:         testNetworkingLinkID,
		InstanceName:   "test-instance",
		LinkUpdateType: updateType,
		LinkContext:    testNetworkingLinkContext(),
		ResourceAInfo: &provider.ResourceInfo{
			ResourceName: "caller",
			InstanceID:   "instance-1",
		},
		ResourceService: resourceService,
	}
}

func vpcAttachedActivation() NetworkingActivation {
	return NetworkingActivation{
		Caller: CallerNetworking{
			VPCID:            "vpc-1",
			SubnetIDs:        []string{"subnet-1"},
			SecurityGroupIDs: []string{"sg-caller"},
		},
		Region:       "us-west-2",
		AWSService:   "sns",
		EndpointType: ec2types.VpcEndpointTypeInterface,
	}
}

// When no endpoint for the service exists yet, the helper provisions a security group,
// authorizes the caller's group to reach it, and creates the interface VPC endpoint,
// recording the endpoint ID in link data.
func (s *ReconcileLinkNetworkingSuite) Test_creates_interface_endpoint_when_absent() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{}),
		ec2mock.WithCreateSecurityGroupOutput(&ec2.CreateSecurityGroupOutput{GroupId: aws.String("sg-endpoint")}),
		ec2mock.WithAuthorizeSecurityGroupIngressOutput(&ec2.AuthorizeSecurityGroupIngressOutput{}),
		ec2mock.WithCreateVpcEndpointOutput(&ec2.CreateVpcEndpointOutput{
			VpcEndpoint: &ec2types.VpcEndpoint{VpcEndpointId: aws.String("vpce-1")},
		}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	result, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		vpcAttachedActivation(),
		output,
	)

	s.Require().NoError(err)
	ec2Svc.AssertCalled(&s.Suite, "DescribeVpcEndpoints")
	ec2Svc.AssertCalled(&s.Suite, "CreateSecurityGroup")
	ec2Svc.AssertCalled(&s.Suite, "AuthorizeSecurityGroupIngress")
	ec2Svc.AssertCalled(&s.Suite, "CreateVpcEndpoint")

	endpoint := result.LinkData.Fields["callerVPCEndpoint"]
	s.Require().NotNil(endpoint)
	s.Equal("vpce-1", core.StringValue(endpoint.Fields["id"]))
}

// On destroy, when the current link is the endpoint's only referencer, the endpoint and
// its security group are deleted.
func (s *ReconcileLinkNetworkingSuite) Test_removes_endpoint_on_destroy_when_last_referencer() {
	linkTag := utils.CreateTagBlueprintLinkID("link-1")
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{
			VpcEndpoints: []ec2types.VpcEndpoint{
				{VpcEndpointId: aws.String("vpce-1"), Tags: []ec2types.Tag{linkTag}},
			},
		}),
		ec2mock.WithDeleteVpcEndpointsOutput(&ec2.DeleteVpcEndpointsOutput{}),
		ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []ec2types.SecurityGroup{
				{GroupId: aws.String("sg-endpoint"), Tags: []ec2types.Tag{linkTag}},
			},
		}),
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
		vpcAttachedActivation(),
		output,
	)

	s.Require().NoError(err)
	ec2Svc.AssertCalled(&s.Suite, "DeleteVpcEndpoints")
	ec2Svc.AssertCalled(&s.Suite, "DeleteSecurityGroup")
}

// An error resolving the flex VPC from state is surfaced to the caller.
func (s *ReconcileLinkNetworkingSuite) Test_propagates_flex_vpc_lookup_error() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupError(errors.New("vpc not found")),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		vpcAttachedActivation(),
		output,
	)

	s.Require().Error(err)
	s.Contains(err.Error(), "vpc not found")
}

// The flex VPC resource is identified in state by its name, not by the AWS VPC ID the
// caller's vpcConfig carries, so the name has to be recovered from the VPC's tag first.
// This is a regression test for an e2e test run where looking it up by VPC ID
// matched nothing and the nil result was dereferenced, so every
// activation for a genuinely placed caller panicked.
func (s *ReconcileLinkNetworkingSuite) Test_looks_up_the_flex_vpc_by_its_bluelink_name() {
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
		vpcAttachedActivation(),
		output,
	)
	s.Require().NoError(err)

	s.Equal("orders-vpc", rs.LastLookupExternalID())
}

// A caller placed in a VPC that holds no matching flex VPC resource is reported, not
// dereferenced.
func (s *ReconcileLinkNetworkingSuite) Test_errors_when_the_flex_vpc_is_absent_from_state() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
	)
	rs := resourceservicemock.Create()

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		vpcAttachedActivation(),
		output,
	)

	s.Require().Error(err)
	s.Contains(err.Error(), "orders-vpc")
}

func gatewayActivation() NetworkingActivation {
	return NetworkingActivation{
		Caller: CallerNetworking{
			VPCID:            "vpc-1",
			SubnetIDs:        []string{"subnet-1"},
			SecurityGroupIDs: []string{"sg-caller"},
		},
		Region:       "us-west-2",
		AWSService:   "s3",
		EndpointType: ec2types.VpcEndpointTypeGateway,
	}
}

// A gateway endpoint (S3/DynamoDB) attaches to the caller's route tables and has no
// security group: when absent, the helper resolves the route tables and creates the
// gateway endpoint without provisioning a security group.
func (s *ReconcileLinkNetworkingSuite) Test_creates_gateway_endpoint_when_absent() {
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

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	result, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		gatewayActivation(),
		output,
	)

	s.Require().NoError(err)
	ec2Svc.AssertCalled(&s.Suite, "DescribeRouteTables")
	ec2Svc.AssertCalled(&s.Suite, "CreateVpcEndpoint")
	ec2Svc.AssertNotCalled(&s.Suite, "CreateSecurityGroup")

	endpoint := result.LinkData.Fields["callerVPCEndpoint"]
	s.Require().NotNil(endpoint)
	s.Equal("vpce-gw-1", core.StringValue(endpoint.Fields["id"]))
}

// On destroy, when the current link is the gateway endpoint's only referencer, the
// endpoint is deleted and no security group cleanup is attempted (gateway endpoints have
// none).
func (s *ReconcileLinkNetworkingSuite) Test_removes_gateway_endpoint_on_destroy_when_last_referencer() {
	linkTag := utils.CreateTagBlueprintLinkID("link-1")
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
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{
			VpcEndpoints: []ec2types.VpcEndpoint{
				{VpcEndpointId: aws.String("vpce-gw-1"), Tags: []ec2types.Tag{linkTag}},
			},
		}),
		ec2mock.WithDeleteVpcEndpointsOutput(&ec2.DeleteVpcEndpointsOutput{}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeDestroy, rs),
		gatewayActivation(),
		output,
	)

	s.Require().NoError(err)
	ec2Svc.AssertCalled(&s.Suite, "DeleteVpcEndpoints")
	ec2Svc.AssertNotCalled(&s.Suite, "DeleteSecurityGroup")
}

// When a gateway endpoint already exists but is not yet associated with the caller's
// route table or tagged for this link, the helper tags it (reference counting) and adds
// the missing route table association rather than creating a new endpoint.
func (s *ReconcileLinkNetworkingSuite) Test_updates_existing_gateway_endpoint() {
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
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{
			VpcEndpoints: []ec2types.VpcEndpoint{
				// Existing endpoint with no link tag and not yet on rtb-1.
				{VpcEndpointId: aws.String("vpce-gw-1"), RouteTableIds: []string{"rtb-other"}},
			},
		}),
		ec2mock.WithCreateTagsOutput(&ec2.CreateTagsOutput{}),
		ec2mock.WithModifyVpcEndpointOutput(&ec2.ModifyVpcEndpointOutput{}),
		ec2mock.WithDescribeManagedPrefixListsOutput(gatewayPrefixListOutput()),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeUpdate, rs),
		gatewayActivation(),
		output,
	)

	s.Require().NoError(err)
	ec2Svc.AssertCalled(&s.Suite, "CreateTags")
	ec2Svc.AssertCalled(&s.Suite, "ModifyVpcEndpoint")
	ec2Svc.AssertNotCalled(&s.Suite, "CreateVpcEndpoint")
}

func TestReconcileLinkNetworkingSuite(t *testing.T) {
	suite.Run(t, new(ReconcileLinkNetworkingSuite))
}

// The AWS-managed prefix list a gateway endpoint's service publishes, which is what the
// caller's egress rule targets: a gateway endpoint has no security group to pair with.
func gatewayPrefixListOutput() *ec2.DescribeManagedPrefixListsOutput {
	return &ec2.DescribeManagedPrefixListsOutput{
		PrefixLists: []ec2types.ManagedPrefixList{
			{
				PrefixListId:   aws.String("pl-s3"),
				PrefixListName: aws.String("com.amazonaws.us-west-2.s3"),
			},
		},
	}
}

// Every networking activation resolves the flex VPC's Bluelink name from the AWS VPC's
// tag before it can find the resource in state, so each mock has to answer it.
func flexVPCDescribeOutput() []*ec2.DescribeVpcsOutput {
	return []*ec2.DescribeVpcsOutput{
		{
			Vpcs: []ec2types.Vpc{
				{
					VpcId: aws.String("vpc-1"),
					Tags: []ec2types.Tag{
						{
							Key:   aws.String(flex.TagFlexVPCName),
							Value: aws.String("orders-vpc"),
						},
					},
				},
			},
		},
	}
}

// The caller's egress rule points at the endpoint's security group, and AWS refuses to
// delete a group that another group's rules reference. Leaving it in place made the
// delete fail with DependencyViolation, which is classified as transient and retried
// until the whole deployment times out.
//
// It survived unnoticed for as long as nothing ordered this link against the placement
// link, which deletes the caller's group outright and took the rule with it. Ordering
// access links before the placement link they depend on removed that accident, so the
// teardown has to revoke what it authorised.
func (s *ReconcileLinkNetworkingSuite) Test_revokes_the_callers_rules_before_deleting_the_endpoint_group() {
	linkTag := utils.CreateTagBlueprintLinkID("link-1")
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{
			VpcEndpoints: []ec2types.VpcEndpoint{
				{VpcEndpointId: aws.String("vpce-1"), Tags: []ec2types.Tag{linkTag}},
			},
		}),
		ec2mock.WithDeleteVpcEndpointsOutput(&ec2.DeleteVpcEndpointsOutput{}),
		ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []ec2types.SecurityGroup{
				{GroupId: aws.String("sg-endpoint"), Tags: []ec2types.Tag{linkTag}},
			},
		}),
		ec2mock.WithDescribeSecurityGroupRulesOutput(&ec2.DescribeSecurityGroupRulesOutput{
			SecurityGroupRules: []ec2types.SecurityGroupRule{
				{
					SecurityGroupRuleId: aws.String("sgr-caller-egress"),
					IsEgress:            aws.Bool(true),
				},
			},
		}),
		ec2mock.WithRevokeSecurityGroupEgressOutput(&ec2.RevokeSecurityGroupEgressOutput{}),
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
		vpcAttachedActivation(),
		output,
	)

	s.Require().NoError(err)

	// The caller's group specifically, and before the endpoint group's own rules are
	// touched. Asserting only that some revoke happened would pass on the endpoint
	// group's revoke alone and prove nothing.
	ec2Svc.AssertCalledWith(
		&s.Suite,
		"RevokeSecurityGroupEgress",
		0,
		plugintestutils.Any,
		&ec2.RevokeSecurityGroupEgressInput{
			GroupId:              aws.String("sg-caller"),
			SecurityGroupRuleIds: []string{"sgr-caller-egress"},
		},
	)
	ec2Svc.AssertCalled(&s.Suite, "DeleteSecurityGroup")
}
