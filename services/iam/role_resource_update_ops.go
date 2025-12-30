package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type roleUpdate struct {
	updateRoleInput             *iam.UpdateRoleInput
	updateAssumeRolePolicyInput *iam.UpdateAssumeRolePolicyInput
}

func (u *roleUpdate) Name() string {
	return "update IAM role"
}

func (u *roleUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	roleName := saveOpCtx.ProviderUpstreamID

	hasChanges := false

	// Check for basic role updates (description, maxSessionDuration, path)
	updateRoleInput, hasRoleUpdate, err := changesToUpdateRoleInput(roleName, specData, changes)
	if err != nil {
		return false, saveOpCtx, err
	}
	if hasRoleUpdate {
		u.updateRoleInput = updateRoleInput
		hasChanges = true
	}

	// Check for assume role policy updates
	updateAssumeRolePolicyInput, hasAssumeRolePolicyUpdate, err := changesToUpdateAssumeRolePolicyInput(roleName, specData, changes)
	if err != nil {
		return false, saveOpCtx, err
	}
	if hasAssumeRolePolicyUpdate {
		u.updateAssumeRolePolicyInput = updateAssumeRolePolicyInput
		hasChanges = true
	}

	return hasChanges, saveOpCtx, nil
}

func (u *roleUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data:               saveOpCtx.Data,
		ProviderUpstreamID: saveOpCtx.ProviderUpstreamID,
	}

	// Execute basic role updates
	if u.updateRoleInput != nil {
		updateRoleOutput, err := iamService.UpdateRole(ctx, u.updateRoleInput)
		if err != nil {
			return saveOpCtx, err
		}
		newSaveOpCtx.Data["updateRoleOutput"] = updateRoleOutput
	}

	// Execute assume role policy updates
	if u.updateAssumeRolePolicyInput != nil {
		updateAssumeRolePolicyOutput, err := iamService.UpdateAssumeRolePolicy(ctx, u.updateAssumeRolePolicyInput)
		if err != nil {
			return saveOpCtx, err
		}
		newSaveOpCtx.Data["updateAssumeRolePolicyOutput"] = updateAssumeRolePolicyOutput
	}

	return newSaveOpCtx, nil
}

func changesToUpdateRoleInput(
	roleName string,
	specData *core.MappingNode,
	changes *provider.Changes,
) (*iam.UpdateRoleInput, bool, error) {
	input := &iam.UpdateRoleInput{
		RoleName: aws.String(roleName),
	}

	hasValuesToSave := false

	// Check if description was modified
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.description" {
			if description, exists := pluginutils.GetValueByPath("$.description", specData); exists {
				input.Description = aws.String(core.StringValue(description))
				hasValuesToSave = true
			}
		}
		if fieldChange.FieldPath == "spec.maxSessionDuration" {
			if maxSessionDuration, exists := pluginutils.GetValueByPath("$.maxSessionDuration", specData); exists {
				input.MaxSessionDuration = aws.Int32(int32(core.IntValue(maxSessionDuration)))
				hasValuesToSave = true
			}
		}
	}

	return input, hasValuesToSave, nil
}

func changesToUpdateAssumeRolePolicyInput(
	roleName string,
	specData *core.MappingNode,
	changes *provider.Changes,
) (*iam.UpdateAssumeRolePolicyInput, bool, error) {
	// Check if assumeRolePolicyDocument was modified
	assumeRolePolicyModified := false
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.assumeRolePolicyDocument" {
			assumeRolePolicyModified = true
			break
		}
	}

	if !assumeRolePolicyModified {
		return nil, false, nil
	}

	assumeRolePolicyDocument, exists := pluginutils.GetValueByPath("$.assumeRolePolicyDocument", specData)
	if !exists {
		return nil, false, nil
	}

	// Convert the structured policy document to JSON string
	policyJSON, err := json.Marshal(assumeRolePolicyDocument)
	if err != nil {
		return nil, false, err
	}

	// URL encode the policy document
	policyDocument := string(policyJSON)
	encodedPolicyDocument := url.QueryEscape(policyDocument)

	input := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyDocument: aws.String(encodedPolicyDocument),
	}

	return input, true, nil
}

type roleInlinePoliciesUpdate struct {
	toAdd    []*core.MappingNode
	toUpdate []*core.MappingNode
	toRemove []string
}

func (r *roleInlinePoliciesUpdate) Name() string {
	return "update inline policies"
}

func (r *roleInlinePoliciesUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if policies were modified
	policiesModified := false
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.policies" {
			policiesModified = true
			break
		}
	}

	if !policiesModified {
		return false, saveOpCtx, nil
	}

	// Compare current and desired inline policies
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentPolicies, _ := pluginutils.GetValueByPath("$.policies", currentStateSpecData)
	newPolicies := specData.Fields["policies"]

	result := diffIAMPolicies(currentPolicies, newPolicies)
	r.toAdd = result.toAdd
	r.toUpdate = result.toUpdate
	r.toRemove = result.toRemove

	return len(r.toAdd) > 0 || len(r.toUpdate) > 0 || len(r.toRemove) > 0, saveOpCtx, nil
}

func (r *roleInlinePoliciesUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	roleName := saveOpCtx.ProviderUpstreamID

	// Remove policies
	for _, policyName := range r.toRemove {
		_, err := iamService.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
			RoleName:   aws.String(roleName),
			PolicyName: aws.String(policyName),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to delete inline policy %s: %w", policyName, err)
		}
	}

	// Add and update policies (both use PutRolePolicy)
	allPolicies := append(r.toAdd, r.toUpdate...)
	for _, policyNode := range allPolicies {
		policyName := core.StringValue(policyNode.Fields["policyName"])
		policyDocNode := policyNode.Fields["policyDocument"]

		policyJSON, err := json.Marshal(policyDocNode)
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to marshal inline policy document: %w", err)
		}

		_, err = iamService.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
			RoleName:       aws.String(roleName),
			PolicyName:     aws.String(policyName),
			PolicyDocument: aws.String(string(policyJSON)),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to put inline policy %s: %w", policyName, err)
		}
	}

	return saveOpCtx, nil
}

type roleManagedPoliciesUpdate struct {
	toAttach []string
	toDetach []string
}

func (r *roleManagedPoliciesUpdate) Name() string {
	return "update managed policies"
}

func (r *roleManagedPoliciesUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if managedPolicyArns was modified
	managedPolicyArnsModified := false
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.managedPolicyArns" {
			managedPolicyArnsModified = true
			break
		}
	}

	if !managedPolicyArnsModified {
		return false, saveOpCtx, nil
	}

	// Compare current and desired managed policies
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentPolicies, _ := pluginutils.GetValueByPath("$.managedPolicyArns", currentStateSpecData)
	newPolicies := specData.Fields["managedPolicyArns"]

	currentSet := make(map[string]bool)
	if currentPolicies != nil {
		for _, policyArn := range currentPolicies.Items {
			currentSet[core.StringValue(policyArn)] = true
		}
	}

	newSet := make(map[string]bool)
	if newPolicies != nil {
		for _, policyArn := range newPolicies.Items {
			newSet[core.StringValue(policyArn)] = true
		}
	}

	// Determine policies to attach and detach
	for arn := range newSet {
		if !currentSet[arn] {
			r.toAttach = append(r.toAttach, arn)
		}
	}

	for arn := range currentSet {
		if !newSet[arn] {
			r.toDetach = append(r.toDetach, arn)
		}
	}

	return len(r.toAttach) > 0 || len(r.toDetach) > 0, saveOpCtx, nil
}

func (r *roleManagedPoliciesUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	roleName := saveOpCtx.ProviderUpstreamID

	// Detach policies
	for _, policyArn := range r.toDetach {
		_, err := iamService.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: aws.String(policyArn),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to detach managed policy %s: %w", policyArn, err)
		}
	}

	// Attach policies
	for _, policyArn := range r.toAttach {
		_, err := iamService.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: aws.String(policyArn),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to attach managed policy %s: %w", policyArn, err)
		}
	}

	return saveOpCtx, nil
}

type roleTagsUpdate struct {
	toAdd    []types.Tag
	toRemove []string
}

func (r *roleTagsUpdate) Name() string {
	return "update tags"
}

func (r *roleTagsUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	deployInput, _ := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
	diffResult := utils.DiffTagsWithBluelink(
		changes,
		deployInput,
		"$.tags",
		toIAMTag,
	)
	r.toAdd = diffResult.ToSet
	r.toRemove = diffResult.ToRemove

	return len(r.toAdd) > 0 || len(r.toRemove) > 0, saveOpCtx, nil
}

func (r *roleTagsUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	roleName := saveOpCtx.ProviderUpstreamID

	// Remove tags
	if len(r.toRemove) > 0 {
		_, err := iamService.UntagRole(ctx, &iam.UntagRoleInput{
			RoleName: aws.String(roleName),
			TagKeys:  r.toRemove,
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to remove tags: %w", err)
		}
	}

	// Add/update tags
	if len(r.toAdd) > 0 {
		_, err := iamService.TagRole(ctx, &iam.TagRoleInput{
			RoleName: aws.String(roleName),
			Tags:     r.toAdd,
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to add tags: %w", err)
		}
	}

	return saveOpCtx, nil
}

type rolePermissionsBoundaryUpdate struct {
	toSet    *string
	toRemove bool
}

func (r *rolePermissionsBoundaryUpdate) Name() string {
	return "update permissions boundary"
}

func (r *rolePermissionsBoundaryUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if permissions boundary was modified
	permissionsBoundaryModified := false
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.permissionsBoundary" {
			permissionsBoundaryModified = true
			break
		}
	}

	if !permissionsBoundaryModified {
		return false, saveOpCtx, nil
	}

	// Get new permissions boundary value
	newPermissionsBoundary, exists := pluginutils.GetValueByPath("$.permissionsBoundary", specData)

	// Determine if we need to set or remove the permissions boundary
	if !exists || newPermissionsBoundary == nil {
		// New value is nil or doesn't exist, remove the permissions boundary
		r.toRemove = true
		r.toSet = nil
	} else {
		// New value exists, set the permissions boundary
		newValue := core.StringValue(newPermissionsBoundary)
		r.toSet = &newValue
		r.toRemove = false
	}

	return true, saveOpCtx, nil
}

func (r *rolePermissionsBoundaryUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	roleName := saveOpCtx.ProviderUpstreamID

	if r.toRemove {
		// Remove permissions boundary
		_, err := iamService.DeleteRolePermissionsBoundary(ctx, &iam.DeleteRolePermissionsBoundaryInput{
			RoleName: aws.String(roleName),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to remove permissions boundary for role %s: %w", roleName, err)
		}
	} else if r.toSet != nil {
		// Set permissions boundary
		_, err := iamService.PutRolePermissionsBoundary(ctx, &iam.PutRolePermissionsBoundaryInput{
			RoleName:            aws.String(roleName),
			PermissionsBoundary: aws.String(*r.toSet),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to set permissions boundary for role %s: %w", roleName, err)
		}
	}

	return saveOpCtx, nil
}
