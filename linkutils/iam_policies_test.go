//go:build unit

package linkutils

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/stretchr/testify/suite"
)

type IAMPoliciesSuite struct {
	suite.Suite
}

func (s *IAMPoliciesSuite) Test_inline_placement_maps_statement_by_sid() {
	roleLinkData := core.MappingNodeFields(PermissionFieldName, core.MappingNodeFromString("stmt"))
	mappings := map[string]string{}

	AppendRoleAccessMapping(
		mappings, roleLinkData, "ordersRole", "ordersFunctionExecutionRole", "DynamoDBAccessordersTable",
		RoleAccessResult{PlacedSlot: inlineAccessPolicyName, PlacedSlotInline: true},
	)

	s.Equal(
		PermissionFieldPath("ordersFunctionExecutionRole"),
		mappings[InlineAccessStatementPath("ordersRole", "DynamoDBAccessordersTable")],
	)
	// Inline placements do not record a managed policy ARN in the link data.
	s.NotContains(roleLinkData.Fields, ManagedPolicyArnFieldName)
}

func (s *IAMPoliciesSuite) Test_managed_overflow_maps_arn_and_records_link_data() {
	const arn = "arn:aws:iam::123456789012:policy/bluelink-link-access-1"
	roleLinkData := core.MappingNodeFields(PermissionFieldName, core.MappingNodeFromString("stmt"))
	mappings := map[string]string{}

	AppendRoleAccessMapping(
		mappings, roleLinkData, "ordersRole", "ordersFunctionExecutionRole", "DynamoDBAccessordersTable",
		RoleAccessResult{PlacedSlot: "bluelink-link-access-1", PlacedSlotInline: false, PlacedSlotARN: arn},
	)

	// The attached managed policy ARN is mapped onto the role's managedPolicyArns so
	// the role does not detach it as drift.
	s.Equal(
		ManagedPolicyArnFieldPath("ordersFunctionExecutionRole"),
		mappings[ManagedAccessArnPath("ordersRole", arn)],
	)
	s.Equal(arn, core.StringValue(roleLinkData.Fields[ManagedPolicyArnFieldName]))
	// No inline statement mapping for an overflow placement.
	s.NotContains(mappings, InlineAccessStatementPath("ordersRole", "DynamoDBAccessordersTable"))
}

func (s *IAMPoliciesSuite) Test_removed_grant_records_no_mapping() {
	roleLinkData := core.MappingNodeFields(PermissionFieldName, core.MappingNodeFromString("stmt"))
	mappings := map[string]string{}

	AppendRoleAccessMapping(
		mappings, roleLinkData, "ordersRole", "ordersFunctionExecutionRole", "DynamoDBAccessordersTable",
		RoleAccessResult{},
	)

	s.Empty(mappings)
	s.NotContains(roleLinkData.Fields, ManagedPolicyArnFieldName)
}

func TestIAMPoliciesSuite(t *testing.T) {
	suite.Run(t, new(IAMPoliciesSuite))
}
