//go:build unit

package linkutils

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/stretchr/testify/require"
)

func vpcStateWithMintedGroups(groupsByName map[string]string) *state.ResourceState {
	fields := map[string]*core.MappingNode{}
	for name, groupID := range groupsByName {
		fields[name] = core.MappingNodeFromString(groupID)
	}

	return &state.ResourceState{
		Name: "appVpc",
		SpecData: core.MappingNodeFields(
			"name", core.MappingNodeFromString("orders-vpc"),
			"securityGroupIdsByName", &core.MappingNode{Fields: fields},
		),
	}
}

func TestTargetGroupMintedByVPCSelectsTheMintedGroup(t *testing.T) {
	groupID, err := targetGroupMintedByVPC(
		[]string{"sg-db"},
		vpcStateWithMintedGroups(map[string]string{"db": "sg-db"}),
		linkInputForResources("getOrderFunction", "ordersDb"),
	)

	require.NoError(t, err)
	require.Equal(t, "sg-db", groupID)
}

// A target commonly carries a group the author manages alongside the one the VPC minted.
// Pairing against the author's group would open a path to every other resource sharing
// it, which is the leak a group per target exists to close.
func TestTargetGroupMintedByVPCIgnoresGroupsTheAuthorManages(t *testing.T) {
	groupID, err := targetGroupMintedByVPC(
		// The author's group is listed first, which is exactly what taking Items[0]
		// used to pick.
		[]string{"sg-authorAdmin", "sg-db"},
		vpcStateWithMintedGroups(map[string]string{"db": "sg-db"}),
		linkInputForResources("getOrderFunction", "ordersDb"),
	)

	require.NoError(t, err)
	require.Equal(t, "sg-db", groupID)
}

func TestTargetGroupMintedByVPCFailsWhenTheTargetReferencesNone(t *testing.T) {
	_, err := targetGroupMintedByVPC(
		[]string{"sg-authorAdmin"},
		vpcStateWithMintedGroups(map[string]string{"db": "sg-db"}),
		linkInputForResources("getOrderFunction", "ordersDb"),
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "ordersDb")
	require.Contains(t, err.Error(), "securityGroups")
}

// Two minted groups on one target is ambiguous, and guessing would be the same defect in
// a new form. The message names both so the author can drop one.
func TestTargetGroupMintedByVPCFailsWhenTheTargetReferencesSeveral(t *testing.T) {
	_, err := targetGroupMintedByVPC(
		[]string{"sg-db", "sg-cache"},
		vpcStateWithMintedGroups(map[string]string{"db": "sg-db", "cache": "sg-cache"}),
		linkInputForResources("getOrderFunction", "ordersDb"),
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "ordersDb")
	require.Contains(t, err.Error(), "sg-cache")
	require.Contains(t, err.Error(), "sg-db")
}

// A VPC that has never declared a name mints nothing, so every target in it fails the
// same way rather than falling back to pairing against something arbitrary.
func TestTargetGroupMintedByVPCFailsWhenTheVPCMintedNothing(t *testing.T) {
	_, err := targetGroupMintedByVPC(
		[]string{"sg-db"},
		&state.ResourceState{
			Name:     "appVpc",
			SpecData: core.MappingNodeFields("name", core.MappingNodeFromString("orders-vpc")),
		},
		linkInputForResources("getOrderFunction", "ordersDb"),
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "ordersDb")
}
