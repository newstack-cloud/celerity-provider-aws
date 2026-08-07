//go:build integration

package e2e

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink-provider-aws/flex"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/require"
)

// TestFlexVPCNetworkingActivation deploys a VPC-attached function linked to SQS and
// DynamoDB, and asserts the security group rules and endpoints that result.
//
// Everything here is an assertion about what EC2 accepts and returns, which is why it
// cannot be a unit test. The mocks accepted the previous malformed group reference for
// as long as it existed, so only a real account can settle whether the corrected shape
// is the one EC2 takes.
func TestFlexVPCNetworkingActivation(t *testing.T) {
	t.Parallel()

	h := Setup(t)
	vpcName := h.Name("vpc")
	deployed := h.Deploy(t, "flex_vpc_networking.blueprint", map[string]*core.ScalarValue{
		"region": core.ScalarFromString(h.Region),
	})

	vpcSpec := deployed.ResourceSpec(t, "netVPC")
	vpcIDNode, _ := pluginutils.GetValueByPath("$.vpcId", vpcSpec)
	vpcID := core.StringValue(vpcIDNode)
	require.NotEmpty(t, vpcID, "the flex VPC should expose its VPC ID")

	// The placed function gets its own group rather than the VPC's shared one, so a
	// grant to one workload is not a grant to every workload in the VPC.
	workloadGroups := h.WorkloadSecurityGroups(t, vpcName)
	require.Len(t, workloadGroups, 1, "expected exactly one per-workload security group")

	callerGroup, hasCallerGroup := workloadGroups["netFunction"]
	require.True(t, hasCallerGroup, "expected a security group for the placed function")
	callerGroupID := aws.ToString(callerGroup.GroupId)

	_, taggedAsBaseGroup := TagValue(callerGroup.Tags, flex.TagFlexVPCSecurityGroup)
	require.False(
		t,
		taggedAsBaseGroup,
		"a workload group carrying the base-group tag would be adopted into the VPC's own state",
	)

	callerRules := h.SecurityGroupRules(t, callerGroupID)

	// The default allow-all egress must be gone, or the per-workload group grants
	// everything and the whole model collapses.
	require.False(
		t,
		hasAllTrafficEgress(callerRules),
		"the default allow-all egress rule should have been revoked on the workload group",
	)

	endpoints := h.VPCEndpointsForVPC(t, vpcID)
	interfaceEndpoint, hasInterface := endpointOfType(endpoints, ec2types.VpcEndpointTypeInterface)
	require.True(t, hasInterface, "expected an interface endpoint for SQS")
	gatewayEndpoint, hasGateway := endpointOfType(endpoints, ec2types.VpcEndpointTypeGateway)
	require.True(t, hasGateway, "expected a gateway endpoint for DynamoDB")

	// The endpoint's group must admit the caller by group ID. The previous call
	// passed an ID into SourceSecurityGroupName, an EC2-Classic field, so the rule it
	// produced could never match a VPC group reference.
	require.NotEmpty(
		t,
		interfaceEndpoint.Groups,
		"the interface endpoint should be attached to a security group",
	)
	endpointGroupID := aws.ToString(interfaceEndpoint.Groups[0].GroupId)
	endpointRules := h.SecurityGroupRules(t, endpointGroupID)
	require.True(
		t,
		hasGroupIngress(endpointRules, callerGroupID, 443),
		"the interface endpoint's group should admit the caller's group on 443",
	)

	// This is the half that makes the endpoint reachable at all. Without egress on the
	// caller's own group the endpoint exists, admits the caller, and the caller still
	// cannot send anything to it.
	require.True(
		t,
		hasGroupEgress(callerRules, endpointGroupID, 443),
		"the caller's group should have egress to the interface endpoint's group",
	)

	// The gateway half: a gateway endpoint has no security group, so the caller's
	// egress goes to the service's managed prefix list.
	require.True(
		t,
		hasPrefixListEgress(callerRules),
		"the caller's group should have egress to the gateway service's prefix list",
	)
	require.Equal(
		t,
		ec2types.VpcEndpointTypeGateway,
		gatewayEndpoint.VpcEndpointType,
		"DynamoDB should be reached through a gateway endpoint",
	)

	// The VPC creates a group for each name it declares, and creates it empty.
	// The group is an identity for a resource to reference, not a grant, so a resource
	// carrying it should reach nothing until a link opens a path.
	namedGroups := h.NamedSecurityGroups(t, vpcName)
	require.Contains(t, namedGroups, "db", "the VPC should have minted a group named db")

	dbGroupID := aws.ToString(namedGroups["db"].GroupId)
	dbGroupRules := h.SecurityGroupRules(t, dbGroupID)
	require.Empty(
		t,
		egressRules(dbGroupRules),
		"a minted group must not keep the allow-all egress AWS attaches to a new group",
	)
	require.Empty(
		t,
		ingressRules(dbGroupRules),
		"a minted group starts with no ingress; only a link may open one",
	)

	// The name is what a resource references the group by, so it has to survive into
	// the VPC's own state rather than only existing as a tag in AWS.
	byName, hasByName := pluginutils.GetValueByPath("$.securityGroupIdsByName", vpcSpec)
	require.True(t, hasByName, "the VPC should expose securityGroupIdsByName")
	require.Equal(
		t,
		dbGroupID,
		core.StringValue(byName.Fields["db"]),
		"the exposed ID should be the group that was actually minted",
	)

	// Every rule a link created carries that link's identity, which is what lets a
	// teardown revoke exactly its own rules and leave a shared group's others intact.
	for _, rule := range append(egressRules(callerRules), ingressRules(endpointRules)...) {
		require.Truef(
			t,
			hasLinkIDTag(rule.Tags),
			"rule %s carries no link identity tag, so nothing can revoke just this link's rules",
			aws.ToString(rule.SecurityGroupRuleId),
		)
	}
}

// TestFlexVPCNetworkingTeardown probes a previous destroy ordering failure.
//
// On destroy an access link reads the caller's attachment from the function's state,
// and nothing orders it against the placement link's own destroy, which clears
// vpcConfig and deletes the workload group. If placement runs first the access link
// sees an unattached caller and never revokes its rules, and a group that is still
// referenced by another group's rules may refuse to delete.
//
// The assertion is deliberately about the end state rather than the ordering: whatever
// order the framework picks, teardown has to leave nothing behind.
func TestFlexVPCNetworkingTeardown(t *testing.T) {
	t.Parallel()

	h := Setup(t)
	vpcName := h.Name("vpc")
	deployed := h.Deploy(t, "flex_vpc_networking.blueprint", map[string]*core.ScalarValue{
		"region": core.ScalarFromString(h.Region),
	})

	vpcSpec := deployed.ResourceSpec(t, "netVPC")
	vpcIDNode, _ := pluginutils.GetValueByPath("$.vpcId", vpcSpec)
	vpcID := core.StringValue(vpcIDNode)
	require.NotEmpty(t, vpcID)

	require.Len(t, h.WorkloadSecurityGroups(t, vpcName), 1, "expected the workload group to exist before teardown")

	// Destroy inside the test body so the account can be inspected afterwards; the
	// registered cleanup skips a second destroy.
	//
	// DestroyNow re-runs until the teardown converges. A destroy of a VPC-attached
	// workload waits on AWS releasing the network interfaces that hold its security
	// groups, which was measured between 18 and 977 seconds for the same operation, so
	// the provider yields rather than blocking and the work finishes across attempts.
	deployed.DestroyNow()

	// Convergence is what teardown promises, so it is what is asserted: everything the
	// VPC owns is gone, and the VPC with it.
	require.Empty(
		t,
		h.WorkloadSecurityGroups(t, vpcName),
		"the placement link's workload security group should have been deleted",
	)

	require.False(
		t,
		h.VPCExists(t, vpcID),
		"the VPC should have been deleted once teardown converged",
	)

	// Nothing the links created may outlive them either: a surviving endpoint security
	// group holds an ingress rule referencing the workload group, which is what blocked
	// the VPC from being deletable at all before this was fixed.
	require.Empty(
		t,
		h.VPCEndpointsForVPC(t, vpcID),
		"the links' VPC endpoints should have been removed",
	)
}

func endpointOfType(
	endpoints []ec2types.VpcEndpoint,
	endpointType ec2types.VpcEndpointType,
) (ec2types.VpcEndpoint, bool) {
	for _, endpoint := range endpoints {
		if endpoint.VpcEndpointType == endpointType {
			return endpoint, true
		}
	}

	return ec2types.VpcEndpoint{}, false
}

func hasPrefixListEgress(rules []ec2types.SecurityGroupRule) bool {
	for _, rule := range egressRules(rules) {
		if rule.PrefixListId != nil && aws.ToString(rule.PrefixListId) != "" {
			return true
		}
	}

	return false
}

func hasLinkIDTag(tags []ec2types.Tag) bool {
	for _, tag := range tags {
		if len(aws.ToString(tag.Key)) > len(utils.TagBlueprintLinkIDPrefix) &&
			aws.ToString(tag.Key)[:len(utils.TagBlueprintLinkIDPrefix)] == utils.TagBlueprintLinkIDPrefix {
			return true
		}
	}

	return false
}
