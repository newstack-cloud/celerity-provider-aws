//go:build unit

package linkutils

import (
	"context"
	"errors"
	"testing"

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
	"github.com/stretchr/testify/suite"
)

type ActivateLinkNetworkingSuite struct {
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
		),
	}
}

func networkingInput(
	updateType provider.LinkUpdateType,
	resourceService provider.ResourceService,
) *provider.LinkUpdateIntermediaryResourcesInput {
	return &provider.LinkUpdateIntermediaryResourcesInput{
		LinkID:         "link-1",
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
func (s *ActivateLinkNetworkingSuite) Test_creates_interface_endpoint_when_absent() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
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
	result, err := ActivateLinkNetworking(
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
func (s *ActivateLinkNetworkingSuite) Test_removes_endpoint_on_destroy_when_last_referencer() {
	linkTag := utils.CreateTagBlueprintLinkID("link-1")
	ec2Svc := ec2mock.CreateEc2ServiceMock(
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
	_, err := ActivateLinkNetworking(
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
func (s *ActivateLinkNetworkingSuite) Test_propagates_flex_vpc_lookup_error() {
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupError(errors.New("vpc not found")),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ActivateLinkNetworking(
		context.Background(),
		nil,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		vpcAttachedActivation(),
		output,
	)

	s.Require().Error(err)
	s.Contains(err.Error(), "vpc not found")
}

func TestActivateLinkNetworkingSuite(t *testing.T) {
	suite.Run(t, new(ActivateLinkNetworkingSuite))
}
