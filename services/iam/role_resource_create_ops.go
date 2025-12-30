package iam

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type roleCreate struct {
	input                   *iam.CreateRoleInput
	uniqueRoleNameGenerator utils.UniqueNameGenerator
}

func (r *roleCreate) Name() string {
	return "create role"
}

func (r *roleCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	createInput, hasValues, err := changesToCreateRoleInput(specData)
	if err != nil {
		return false, saveOpCtx, err
	}

	deployInput, ok := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
	if !ok || deployInput == nil {
		return false, saveOpCtx, fmt.Errorf("ResourceDeployInput not found in SaveOperationContext.Data")
	}

	// Generate unique role name if not provided
	if createInput.RoleName == nil || aws.ToString(createInput.RoleName) == "" {
		// Use the injected generator or default to the IAM role generator
		generator := r.uniqueRoleNameGenerator
		if generator == nil {
			generator = utils.IAMRoleNameGenerator
		}

		uniqueRoleName, err := generator(deployInput)
		if err != nil {
			return false, saveOpCtx, err
		}

		createInput.RoleName = aws.String(uniqueRoleName)
		hasValues = true
	}

	// Merge Bluelink system tags with user-defined tags
	createInput.Tags = mergeBluelinkTagsWithIAMTags(deployInput, createInput.Tags)

	r.input = createInput
	return hasValues, saveOpCtx, nil
}

func (r *roleCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	createRoleOutput, err := iamService.CreateRole(ctx, r.input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create IAM role: %w", err)
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(createRoleOutput.Role.Arn)
	newSaveOpCtx.Data["createRoleOutput"] = createRoleOutput
	newSaveOpCtx.Data["roleArn"] = aws.ToString(createRoleOutput.Role.Arn)

	return newSaveOpCtx, nil
}

func newRoleCreate(generator utils.UniqueNameGenerator) *roleCreate {
	return &roleCreate{
		uniqueRoleNameGenerator: generator,
	}
}

type roleInlinePoliciesCreate struct {
	policies []*core.MappingNode
}

func (r *roleInlinePoliciesCreate) Name() string {
	return "create inline policies"
}

func (r *roleInlinePoliciesCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if there are inline policies to create
	if policiesNode, ok := specData.Fields["policies"]; ok && policiesNode != nil && len(policiesNode.Items) > 0 {
		r.policies = policiesNode.Items
		return true, saveOpCtx, nil
	}
	return false, saveOpCtx, nil
}

func (r *roleInlinePoliciesCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Get the role name from the created role
	createRoleOutput, ok := saveOpCtx.Data["createRoleOutput"].(*iam.CreateRoleOutput)
	if !ok {
		return saveOpCtx, fmt.Errorf("createRoleOutput not found in save operation context")
	}

	roleName := aws.ToString(createRoleOutput.Role.RoleName)

	// Create each inline policy
	for _, policyNode := range r.policies {
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

type roleManagedPoliciesCreate struct {
	managedPolicyArns []*core.MappingNode
}

func (r *roleManagedPoliciesCreate) Name() string {
	return "create managed policies"
}

func (r *roleManagedPoliciesCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if there are managed policies to attach
	if managedPolicyArnsNode, ok := specData.Fields["managedPolicyArns"]; ok && managedPolicyArnsNode != nil && len(managedPolicyArnsNode.Items) > 0 {
		r.managedPolicyArns = managedPolicyArnsNode.Items
		return true, saveOpCtx, nil
	}
	return false, saveOpCtx, nil
}

func (r *roleManagedPoliciesCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Get the role name from the created role
	createRoleOutput, ok := saveOpCtx.Data["createRoleOutput"].(*iam.CreateRoleOutput)
	if !ok {
		return saveOpCtx, fmt.Errorf("createRoleOutput not found in save operation context")
	}

	roleName := aws.ToString(createRoleOutput.Role.RoleName)

	// Attach each managed policy
	for _, policyArnNode := range r.managedPolicyArns {
		policyArn := core.StringValue(policyArnNode)

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

type rolePermissionsBoundaryCreate struct {
	permissionsBoundary string
}

func (r *rolePermissionsBoundaryCreate) Name() string {
	return "set permissions boundary"
}

func (r *rolePermissionsBoundaryCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if there is a permissions boundary to set
	if permsBoundaryNode, ok := specData.Fields["permissionsBoundary"]; ok && permsBoundaryNode != nil {
		r.permissionsBoundary = core.StringValue(permsBoundaryNode)
		return true, saveOpCtx, nil
	}
	return false, saveOpCtx, nil
}

func (r *rolePermissionsBoundaryCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Get the role name from the created role
	createRoleOutput, ok := saveOpCtx.Data["createRoleOutput"].(*iam.CreateRoleOutput)
	if !ok {
		return saveOpCtx, fmt.Errorf("createRoleOutput not found in save operation context")
	}

	roleName := aws.ToString(createRoleOutput.Role.RoleName)

	_, err := iamService.PutRolePermissionsBoundary(ctx, &iam.PutRolePermissionsBoundaryInput{
		RoleName:            aws.String(roleName),
		PermissionsBoundary: aws.String(r.permissionsBoundary),
	})
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to set permissions boundary for role %s: %w", roleName, err)
	}

	return saveOpCtx, nil
}
