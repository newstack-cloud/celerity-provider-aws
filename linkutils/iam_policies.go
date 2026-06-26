package linkutils

import (
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
)

// This file holds the small link-data field conventions shared by links that grant
// IAM permissions. The role-policy read-modify-write itself is handled by the
// permission allocator in role_access_policy.go (see ReconcileRoleAccessPolicy).

const (
	// PermissionFieldName is the name of the field in the link data that
	// contains the permission (statement) object.
	// This should be used across all link implementations that store permissions
	// in the link data.
	PermissionFieldName = "permission"
	// ManagedPolicyArnFieldName is the name of the field in the link data that holds
	// the ARN of the allocator's managed policy when a grant overflows from the
	// role's inline policy into an attached managed policy.
	ManagedPolicyArnFieldName = "managedPolicyArn"
)

// PermissionFieldPath returns the field path for a permission (statement)
// object in the link data, keyed by the execution role's link-data name.
// This should be used across all link implementations that store permissions
// in the link data.
func PermissionFieldPath(executionRoleName string) string {
	return fmt.Sprintf("[%q].%s", executionRoleName, PermissionFieldName)
}

// ManagedPolicyArnFieldPath returns the field path for the managed policy ARN in the
// link data, keyed by the execution role's link-data name.
func ManagedPolicyArnFieldPath(executionRoleName string) string {
	return fmt.Sprintf("[%q].%s", executionRoleName, ManagedPolicyArnFieldName)
}

// InlineAccessStatementPath returns the ResourceDataMappings key that targets this
// link's statement (by Sid) within the role's shared inline allocator policy, so the
// role's drift/deploy attributes the statement to the link instead of stripping it.
func InlineAccessStatementPath(roleResourceName, sid string) string {
	return fmt.Sprintf(
		"%s::spec.policies[@.policyName=%q].policyDocument.statement[@.sid=%q]",
		roleResourceName,
		inlineAccessPolicyName,
		sid,
	)
}

// ManagedAccessArnPath returns the ResourceDataMappings key that targets the
// allocator's attached managed policy ARN within the role's managedPolicyArns, so
// the role's drift/deploy attributes the attachment to the link instead of
// detaching it. Used for managed (overflow) placements.
func ManagedAccessArnPath(roleResourceName, arn string) string {
	return fmt.Sprintf("%s::spec.managedPolicyArns[@=%q]", roleResourceName, arn)
}

// AppendRoleAccessMapping records, into mappings and the role's link-data node, the
// attribution for an allocator placement so the role does not treat the grant as
// drift. Inline placements map the statement (by Sid); managed (overflow)
// placements map the attached managed policy ARN (and add it to the link data).
// roleLinkData is the per-role link-data object (the value stored under the role's
// link-data key) and may be mutated to carry the managed policy ARN.
func AppendRoleAccessMapping(
	mappings map[string]string,
	roleLinkData *core.MappingNode,
	roleResourceName, linkDataKey, sid string,
	result RoleAccessResult,
) {
	if result.PlacedSlotInline {
		mappings[InlineAccessStatementPath(roleResourceName, sid)] = PermissionFieldPath(linkDataKey)
		return
	}
	if result.PlacedSlotARN != "" {
		roleLinkData.Fields[ManagedPolicyArnFieldName] = core.MappingNodeFromString(result.PlacedSlotARN)
		mappings[ManagedAccessArnPath(roleResourceName, result.PlacedSlotARN)] = ManagedPolicyArnFieldPath(linkDataKey)
	}
}
