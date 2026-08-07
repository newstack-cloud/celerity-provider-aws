//go:build unit

package linkutils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
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
		Region:                 "us-west-2",
		TargetSecurityGroupIDs: []string{"sg-db"},
		TargetPort:             5432,
	}
}

// A security-group-pair activation authorises ingress on the target's group from the caller's
// group, and egress on the caller's group to the target's group, both on the target port.
func (s *ReconcileLinkNetworkingSuite) Test_security_group_pair_opens_ingress_and_egress() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithAuthorizeSecurityGroupIngressOutput(&ec2.AuthorizeSecurityGroupIngressOutput{}),
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

// Every rule the pairing creates is tagged with the link that created it, which is what
// lets destroy revoke exactly its own rules and leave every other writer's alone.
func (s *ReconcileLinkNetworkingSuite) Test_security_group_pair_tags_rules_with_the_link() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithAuthorizeSecurityGroupIngressOutput(&ec2.AuthorizeSecurityGroupIngressOutput{}),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		sgPairActivation(),
		output,
	)
	s.Require().NoError(err)

	expectedTagKey := utils.TagBlueprintLinkIDPrefix + testNetworkingLinkID
	taggedForThisLink := func(specs []ec2types.TagSpecification) bool {
		if len(specs) != 1 || specs[0].ResourceType != ec2types.ResourceTypeSecurityGroupRule {
			return false
		}
		for _, tag := range specs[0].Tags {
			if aws.ToString(tag.Key) == expectedTagKey {
				return true
			}
		}
		return false
	}

	ec2Svc.AssertCalledWith(&s.Suite, "AuthorizeSecurityGroupIngress", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.AuthorizeSecurityGroupIngressInput)
		return ok && taggedForThisLink(in.TagSpecifications)
	})
	ec2Svc.AssertCalledWith(&s.Suite, "AuthorizeSecurityGroupEgress", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.AuthorizeSecurityGroupEgressInput)
		return ok && taggedForThisLink(in.TagSpecifications)
	})
}

// On destroy the link revokes the rules it created on both ends: egress on the caller's
// group and ingress on the target's.
//
// These used to be left in place, which was safe only while every workload shared the
// flex VPC's one security group. A placed workload now has its own group, so a rule
// outliving its link is a grant that workload should no longer hold.
func (s *ReconcileLinkNetworkingSuite) Test_security_group_pair_revokes_its_own_rules_on_destroy() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeSecurityGroupRulesOutput(&ec2.DescribeSecurityGroupRulesOutput{
			SecurityGroupRules: []ec2types.SecurityGroupRule{
				{SecurityGroupRuleId: aws.String("sgr-egress"), IsEgress: aws.Bool(true)},
				{SecurityGroupRuleId: aws.String("sgr-ingress"), IsEgress: aws.Bool(false)},
			},
		}),
		ec2mock.WithRevokeSecurityGroupIngressOutput(&ec2.RevokeSecurityGroupIngressOutput{}),
		ec2mock.WithRevokeSecurityGroupEgressOutput(&ec2.RevokeSecurityGroupEgressOutput{}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	result, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeDestroy, rs),
		sgPairActivation(),
		output,
	)
	s.Require().NoError(err)
	s.Equal(output, result)

	// Scoped to this link: without the tag filter the lookup would return every rule
	// on the group, including ones other links still depend on.
	ec2Svc.AssertCalledWith(&s.Suite, "DescribeSecurityGroupRules", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.DescribeSecurityGroupRulesInput)
		if !ok {
			return false
		}
		byName := map[string][]string{}
		for _, filter := range in.Filters {
			byName[aws.ToString(filter.Name)] = filter.Values
		}
		return len(byName["group-id"]) == 1 &&
			byName["group-id"][0] == "sg-caller" &&
			len(byName["tag-key"]) == 1 &&
			byName["tag-key"][0] == utils.TagBlueprintLinkIDPrefix+testNetworkingLinkID
	})

	// Revoked by rule ID rather than by matching on permissions: two links can produce
	// identical rules, and revoking by shape would remove one another link still needs.
	ec2Svc.AssertCalledWith(&s.Suite, "RevokeSecurityGroupEgress", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.RevokeSecurityGroupEgressInput)
		return ok &&
			aws.ToString(in.GroupId) == "sg-caller" &&
			len(in.SecurityGroupRuleIds) == 1 &&
			in.SecurityGroupRuleIds[0] == "sgr-egress" &&
			in.IpPermissions == nil
	})
	ec2Svc.AssertCalledWith(&s.Suite, "RevokeSecurityGroupIngress", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.RevokeSecurityGroupIngressInput)
		return ok &&
			len(in.SecurityGroupRuleIds) == 1 &&
			in.SecurityGroupRuleIds[0] == "sgr-ingress" &&
			in.IpPermissions == nil
	})
}

// A group holding no rules for this link is left untouched, so a retried teardown does
// not fail and another writer's rules are never revoked.
func (s *ReconcileLinkNetworkingSuite) Test_security_group_pair_destroy_revokes_nothing_when_no_rules_are_its_own() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeSecurityGroupRulesOutput(&ec2.DescribeSecurityGroupRulesOutput{}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeDestroy, rs),
		sgPairActivation(),
		output,
	)
	s.Require().NoError(err)
	ec2Svc.AssertNotCalled(&s.Suite, "RevokeSecurityGroupEgress")
	ec2Svc.AssertNotCalled(&s.Suite, "RevokeSecurityGroupIngress")
}

// An already-present rule (InvalidPermission.Duplicate) is treated as success, so the pairing
// is idempotent across links and re-runs.
func (s *ReconcileLinkNetworkingSuite) Test_security_group_pair_ignores_duplicate_rule() {
	duplicate := &smithy.GenericAPIError{Code: "InvalidPermission.Duplicate", Message: "the rule already exists"}
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithAuthorizeSecurityGroupIngressError(duplicate),
		ec2mock.WithAuthorizeSecurityGroupEgressError(duplicate),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}
	result, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		sgPairActivation(),
		output,
	)
	s.Require().NoError(err)
	s.Equal(output, result)
}
