//go:build unit

package linkutils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/smithy-go"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

func sgPairActivation() NetworkingActivation {
	return NetworkingActivation{
		Caller: CallerNetworking{
			VPCID:            "vpc-1",
			SubnetIDs:        []string{"subnet-1"},
			SecurityGroupIDs: []string{"sg-caller"},
		},
		Region:                "us-west-2",
		TargetSecurityGroupID: "sg-db",
		TargetPort:            5432,
	}
}

// A security-group-pair activation authorises ingress on the target's group from the caller's
// group, and egress on the caller's group to the target's group, both on the target port.
func (s *ActivateLinkNetworkingSuite) Test_security_group_pair_opens_ingress_and_egress() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithAuthorizeSecurityGroupIngressOutput(&ec2.AuthorizeSecurityGroupIngressOutput{}),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	result, err := ActivateLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		sgPairActivation(),
		output,
	)
	s.Require().NoError(err)
	s.Equal(output, result)

	ec2Svc.AssertCalledWith(&s.Suite, "AuthorizeSecurityGroupIngress", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.AuthorizeSecurityGroupIngressInput)
		if !ok || len(in.IpPermissions) != 1 || len(in.IpPermissions[0].UserIdGroupPairs) != 1 {
			return false
		}
		perm := in.IpPermissions[0]
		return aws.ToString(in.GroupId) == "sg-db" &&
			aws.ToInt32(perm.FromPort) == 5432 &&
			aws.ToInt32(perm.ToPort) == 5432 &&
			aws.ToString(perm.IpProtocol) == "tcp" &&
			aws.ToString(perm.UserIdGroupPairs[0].GroupId) == "sg-caller"
	})
	ec2Svc.AssertCalledWith(&s.Suite, "AuthorizeSecurityGroupEgress", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.AuthorizeSecurityGroupEgressInput)
		if !ok || len(in.IpPermissions) != 1 || len(in.IpPermissions[0].UserIdGroupPairs) != 1 {
			return false
		}
		perm := in.IpPermissions[0]
		return aws.ToString(in.GroupId) == "sg-caller" &&
			aws.ToInt32(perm.FromPort) == 5432 &&
			aws.ToString(perm.UserIdGroupPairs[0].GroupId) == "sg-db"
	})
}

// On destroy the pairing is left in place (shared across links), so no rules are revoked.
func (s *ActivateLinkNetworkingSuite) Test_security_group_pair_noop_on_destroy() {
	ec2Svc := ec2mock.CreateEc2ServiceMock()
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	result, err := ActivateLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeDestroy, rs),
		sgPairActivation(),
		output,
	)
	s.Require().NoError(err)
	s.Equal(output, result)
	ec2Svc.AssertNotCalled(&s.Suite, "AuthorizeSecurityGroupIngress")
	ec2Svc.AssertNotCalled(&s.Suite, "AuthorizeSecurityGroupEgress")
}

// An already-present rule (InvalidPermission.Duplicate) is treated as success, so the pairing
// is idempotent across links and re-runs.
func (s *ActivateLinkNetworkingSuite) Test_security_group_pair_ignores_duplicate_rule() {
	duplicate := &smithy.GenericAPIError{Code: "InvalidPermission.Duplicate", Message: "the rule already exists"}
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithAuthorizeSecurityGroupIngressError(duplicate),
		ec2mock.WithAuthorizeSecurityGroupEgressError(duplicate),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	result, err := ActivateLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		sgPairActivation(),
		output,
	)
	s.Require().NoError(err)
	s.Equal(output, result)
}
