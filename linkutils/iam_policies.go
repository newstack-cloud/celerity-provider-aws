package linkutils

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// CreateNewRolePolicy creates a new role policy with the given statements.
// This is useful for links that require permissions to be added to a specific role
// to "activate" the link.
func CreateNewRolePolicy(
	ctx context.Context,
	iamService iamservice.Service,
	instanceName string,
	roleName string,
	statements []map[string]any,
) (string, error) {
	policyDoc := map[string]any{
		"Version":   "2012-10-17",
		"Statement": statements,
	}

	policyName, err := utils.IAMPolicyNameGenerator(&provider.ResourceDeployInput{
		InstanceName: instanceName,
		Changes: &provider.Changes{
			AppliedResourceInfo: provider.ResourceInfo{
				ResourceName: roleName,
			},
		},
	})
	if err != nil {
		return "", err
	}

	policyDocJSON, err := json.Marshal(policyDoc)
	if err != nil {
		return "", err
	}

	_, err = iamService.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(string(policyDocJSON)),
	})
	return policyName, err
}

// UpdateExistingRolePolicy updates an existing role policy with the given statements.
// This is useful for links that require permissions to be added to a specific role
// to "activate" the link.
func UpdateExistingRolePolicy(
	ctx context.Context,
	iamService iamservice.Service,
	roleName string,
	policyName string,
	statements []map[string]any,
	currentPermSIDs []string,
) error {
	policyDocument, err := iamService.GetRolePolicy(ctx, &iam.GetRolePolicyInput{
		RoleName:   aws.String(roleName),
		PolicyName: aws.String(policyName),
	})
	if err != nil {
		return err
	}

	policyDoc := map[string]any{}
	err = json.Unmarshal([]byte(aws.ToString(policyDocument.PolicyDocument)), &policyDoc)
	if err != nil {
		return err
	}

	// Ensure that the current permission SID for this link is removed
	// to avoid duplicate permissions and ensuring old versions of the permission
	// are cleaned up.
	statementList, isList := policyDoc["Statement"].([]any)
	if !isList {
		return fmt.Errorf("policy document \"Statement\" property is not a list")
	}

	filteredStatements := withoutStatementIDs(
		statementList,
		currentPermSIDs,
	)
	policyDoc["Statement"] = append(
		filteredStatements,
		statements,
	)

	policyDocJSON, err := json.Marshal(policyDoc)
	if err != nil {
		return err
	}

	_, err = iamService.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(string(policyDocJSON)),
	})
	return err
}

// RemoveExistingRolePolicyPermissions removes the specified permissions from an existing role policy.
// If the specified permissions are the only ones in the policy, the role policy will be removed.
func RemoveExistingRolePolicyPermissions(
	ctx context.Context,
	iamService iamservice.Service,
	roleName string,
	policyName string,
	permissionsToRemove []string,
) error {
	policyDocument, err := iamService.GetRolePolicy(ctx, &iam.GetRolePolicyInput{
		RoleName:   aws.String(roleName),
		PolicyName: aws.String(policyName),
	})
	if err != nil {
		return err
	}

	policyDoc := map[string]any{}
	err = json.Unmarshal([]byte(aws.ToString(policyDocument.PolicyDocument)), &policyDoc)
	if err != nil {
		return err
	}

	statementList, isList := policyDoc["Statement"].([]any)
	if !isList {
		return fmt.Errorf("policy document \"Statement\" property is not a list")
	}

	statements := withoutStatementIDs(
		statementList,
		permissionsToRemove,
	)

	if len(statements) == 0 {
		_, err = iamService.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
			RoleName:   aws.String(roleName),
			PolicyName: aws.String(policyName),
		})
		return err
	}

	policyDocJSON, err := json.Marshal(policyDoc)
	if err != nil {
		return err
	}

	_, err = iamService.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyName:     aws.String(policyName),
		PolicyDocument: aws.String(string(policyDocJSON)),
	})
	return err
}

func withoutStatementIDs(
	statements []any,
	excludeStatementIDs []string,
) []any {
	filteredStatements := []any{}
	for _, statement := range statements {
		statementMap, isMap := statement.(map[string]any)
		sid, _ := statementMap["Sid"].(string)
		if isMap && !slices.Contains(excludeStatementIDs, sid) {
			filteredStatements = append(filteredStatements, statement)
		}
	}
	return filteredStatements
}

// PermissionFieldPath returns the field path for a permission (statement)
// object in the link data.
// This should be used across all link implementations that store permissions
// in the link data.
func PermissionFieldPath(executionRoleName string) string {
	return fmt.Sprintf("[%q].%s", executionRoleName, PermissionFieldName)
}

const (
	// PermissionFieldName is the name of the field in the link data that
	// contains the permission (statement) object.
	// This should be used across all link implementations that store permissions
	// in the link data.
	PermissionFieldName = "permission"
	// PolicyNameFieldName is the name of the field in the link data that
	// contains the policy name.
	// This should be used across all link implementations that store policy names
	// in the link data.
	PolicyNameFieldName = "policyName"
)

// ExtractPolicyNameAndCurrentPermissionSID extracts the policy name and current
// permission SID from the link data.
func ExtractPolicyNameAndCurrentPermissionSID(
	currentLinkState *state.LinkState,
	executionRoleName string,
	rolePolicies *iam.ListRolePoliciesOutput,
) (string, string) {
	if currentLinkState == nil {
		return rolePolicies.PolicyNames[0], ""
	}

	linkData := linkhelpers.GetLinkDataFromState(currentLinkState)
	policyName, _ := pluginutils.GetValueByPath(
		fmt.Sprintf("$[%q].%s", executionRoleName, PolicyNameFieldName),
		linkData,
	)
	if policyName == nil {
		return rolePolicies.PolicyNames[0], ""
	}

	permission, _ := pluginutils.GetValueByPath(
		fmt.Sprintf("$%s", PermissionFieldPath(executionRoleName)),
		linkData,
	)
	if permission == nil {
		return core.StringValue(policyName), ""
	}

	currentPermSID, _ := pluginutils.GetValueByPath(
		fmt.Sprintf("$%s.Sid", PermissionFieldPath(executionRoleName)),
		linkData,
	)
	if currentPermSID == nil {
		return core.StringValue(policyName), ""
	}

	return core.StringValue(policyName), core.StringValue(currentPermSID)
}
