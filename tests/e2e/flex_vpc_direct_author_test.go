//go:build integration

package e2e

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/require"
)

// slowE2EEnvVar gates the fixtures whose cost is dominated by AWS rather than by
// the provider: standing up an Aurora cluster takes several minutes and bills for
// the privilege, so it runs on request rather than on every integration run.
//
//	E2E_SLOW=1 bash scripts/run-tests.sh --integration
const slowE2EEnvVar = "E2E_SLOW"

func requireSlowE2E(t *testing.T) {
	t.Helper()

	if os.Getenv(slowE2EEnvVar) == "" {
		t.Skipf(
			"%s not set; skipping the slow direct-author fixture (set %s=1 to run it)",
			slowE2EEnvVar, slowE2EEnvVar,
		)
	}
}

// TestFlexVPCDirectAuthorParity deploys the whole direct-author path: a flex VPC,
// a Lambda placed in it, an Aurora cluster referencing a VPC-minted group, and
// the link between them. No policies, no rules and no IAM statements are declared
// anywhere in the fixture.
//
// This is a parity assertion rather than a behaviour one. Every mechanism it
// touches has unit coverage; what only a real account can settle is whether an
// author who writes nothing but placement and links ends up with a working,
// least-privilege deployment. The four defects the one-off RDS probe surfaced
// were all of that kind: each was invisible to mocks and fatal in practice.
//
// Gated behind E2E_SLOW. The cheap flex_vpc_networking fixture covers the same
// links against SQS and DynamoDB and runs every time; this one exists to be run
// before a release, or after a change to group selection or the RDS links.
func TestFlexVPCDirectAuthorParity(t *testing.T) {
	requireSlowE2E(t)
	t.Parallel()

	h := Setup(t)
	vpcName := h.Name("vpc")
	deployed := h.Deploy(t, "flex_vpc_direct_author.blueprint", map[string]*core.ScalarValue{
		"region": core.ScalarFromString(h.Region),
	})

	vpcSpec := deployed.ResourceSpec(t, "appVPC")

	// The group the VPC minted, which is the identity the whole target side hangs
	// off. Read from the VPC's own state rather than from AWS, because a resource
	// references it by what state says it is.
	byName, hasByName := pluginutils.GetValueByPath("$.securityGroupIdsByName", vpcSpec)
	require.True(t, hasByName, "the VPC should expose securityGroupIdsByName")
	mintedGroupID := core.StringValue(byName.Fields["db"])
	require.NotEmpty(t, mintedGroupID, "the VPC should have minted a group named db")

	// What the probe established and this fixture now guards: the cluster reports
	// back exactly the group that was submitted, and exactly one. AWS substitutes
	// nothing and adds nothing of its own, so the link's several-matches branch
	// guards against author-managed groups rather than against AWS behaviour.
	clusterSpec := deployed.ResourceSpec(t, "ordersDb")
	clusterGroups, hasClusterGroups := pluginutils.GetValueByPath(
		"$.vpcSecurityGroupIds",
		clusterSpec,
	)
	require.True(t, hasClusterGroups, "the cluster should record its security groups")
	require.Len(
		t,
		clusterGroups.Items, 1,
		"the cluster should carry exactly the one group the blueprint gave it",
	)
	require.Equal(
		t,
		mintedGroupID,
		core.StringValue(clusterGroups.Items[0]),
		"the cluster should carry the VPC-minted group, not a substitute",
	)

	// Placement: the function has a group of its own rather than sharing the VPC's.
	workloadGroups := h.WorkloadSecurityGroups(t, vpcName)
	require.Len(t, workloadGroups, 1, "expected one per-workload security group")
	callerGroup, hasCallerGroup := workloadGroups["appFunction"]
	require.True(t, hasCallerGroup, "expected a security group for the placed function")
	callerGroupID := aws.ToString(callerGroup.GroupId)

	// The pairing, which is the thing an author never wrote and the link had to
	// derive. Both directions, on the database's port, between exactly these two
	// groups.
	callerRules := h.SecurityGroupRules(t, callerGroupID)
	require.True(
		t,
		hasGroupEgress(callerRules, mintedGroupID, 5432),
		"the caller's group should have egress to the database's group on 5432",
	)

	dbRules := h.SecurityGroupRules(t, mintedGroupID)
	require.True(
		t,
		hasGroupIngress(dbRules, callerGroupID, 5432),
		"the database's group should admit the caller's group on 5432",
	)

	// Least privilege is the claim, so the negative half matters as much: neither
	// group may keep the allow-all egress AWS attaches to a new group.
	require.False(
		t,
		hasAllTrafficEgress(callerRules),
		"the workload group should not keep the default allow-all egress",
	)
	require.False(
		t,
		hasAllTrafficEgress(dbRules),
		"the minted group should not keep the default allow-all egress",
	)

	// The pairing is directional. A database is only ever a target, so it must not
	// have acquired egress of its own from being linked to.
	require.Empty(
		t,
		egressRules(dbRules),
		"a target group should have no egress rules at all",
	)

	// Every rule carries the identity of the link that made it, which is what lets
	// one link's rules be revoked without disturbing another's on a shared group.
	for _, rule := range append(egressRules(callerRules), ingressRules(dbRules)...) {
		require.Truef(
			t,
			hasLinkIDTag(rule.Tags),
			"rule %s carries no link identity tag",
			aws.ToString(rule.SecurityGroupRuleId),
		)
	}

	// The credential never appears in the blueprint: Secrets Manager holds it and
	// the cluster reports the secret back through a computed leaf.
	secretARN, hasSecret := pluginutils.GetValueByPath(
		"$.masterUserSecret.secretArn",
		clusterSpec,
	)
	require.True(t, hasSecret, "the cluster should report its managed master user secret")
	require.NotEmpty(
		t,
		core.StringValue(secretARN),
		"the managed secret should have an ARN once the cluster is available",
	)
}
