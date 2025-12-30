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

type userCreate struct {
	input                   *iam.CreateUserInput
	uniqueUserNameGenerator utils.UniqueNameGenerator
}

func (u *userCreate) Name() string {
	return "create user"
}

func (u *userCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	createInput, hasValues, err := changesToCreateUserInput(specData)
	if err != nil {
		return false, saveOpCtx, err
	}

	deployInput, ok := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
	if !ok || deployInput == nil {
		return false, saveOpCtx, fmt.Errorf("ResourceDeployInput not found in SaveOperationContext.Data")
	}

	// Generate unique user name if not provided
	if createInput.UserName == nil || aws.ToString(createInput.UserName) == "" {
		// Use the injected generator or default to the IAM user generator
		generator := u.uniqueUserNameGenerator
		if generator == nil {
			generator = utils.IAMUserNameGenerator
		}

		uniqueUserName, err := generator(deployInput)
		if err != nil {
			return false, saveOpCtx, err
		}

		createInput.UserName = aws.String(uniqueUserName)
		hasValues = true
	}

	// Merge Bluelink system tags with user-defined tags
	createInput.Tags = mergeBluelinkTagsWithIAMTags(deployInput, createInput.Tags)

	u.input = createInput
	return hasValues, saveOpCtx, nil
}

func (u *userCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	createUserOutput, err := iamService.CreateUser(ctx, u.input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create IAM user: %w", err)
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(createUserOutput.User.Arn)
	newSaveOpCtx.Data["createUserOutput"] = createUserOutput
	newSaveOpCtx.Data["userArn"] = aws.ToString(createUserOutput.User.Arn)

	return newSaveOpCtx, nil
}

func newUserCreate(generator utils.UniqueNameGenerator) *userCreate {
	return &userCreate{
		uniqueUserNameGenerator: generator,
	}
}

type userLoginProfileCreate struct {
	loginProfile *core.MappingNode
}

func (u *userLoginProfileCreate) Name() string {
	return "create login profile"
}

func (u *userLoginProfileCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if there is a login profile to create
	if loginProfileNode, ok := specData.Fields["loginProfile"]; ok && loginProfileNode != nil {
		u.loginProfile = loginProfileNode
		return true, saveOpCtx, nil
	}
	return false, saveOpCtx, nil
}

func (u *userLoginProfileCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Get the user name from the created user
	createUserOutput, ok := saveOpCtx.Data["createUserOutput"].(*iam.CreateUserOutput)
	if !ok {
		return saveOpCtx, fmt.Errorf("createUserOutput not found in save operation context")
	}

	userName := aws.ToString(createUserOutput.User.UserName)
	password := core.StringValue(u.loginProfile.Fields["password"])

	var passwordResetRequired *bool
	if resetRequiredNode, ok := u.loginProfile.Fields["passwordResetRequired"]; ok && resetRequiredNode != nil {
		resetRequired := core.BoolValue(resetRequiredNode)
		passwordResetRequired = &resetRequired
	}

	createInput := &iam.CreateLoginProfileInput{
		UserName: aws.String(userName),
		Password: aws.String(password),
	}
	if passwordResetRequired != nil {
		createInput.PasswordResetRequired = *passwordResetRequired
	}

	_, err := iamService.CreateLoginProfile(ctx, createInput)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create login profile for user %s: %w", userName, err)
	}

	return saveOpCtx, nil
}

type userInlinePoliciesCreate struct {
	policies []*core.MappingNode
}

func (u *userInlinePoliciesCreate) Name() string {
	return "create inline policies"
}

func (u *userInlinePoliciesCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if there are inline policies to create
	if policiesNode, ok := specData.Fields["policies"]; ok && policiesNode != nil && len(policiesNode.Items) > 0 {
		u.policies = policiesNode.Items
		return true, saveOpCtx, nil
	}
	return false, saveOpCtx, nil
}

func (u *userInlinePoliciesCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Get the user name from the created user
	createUserOutput, ok := saveOpCtx.Data["createUserOutput"].(*iam.CreateUserOutput)
	if !ok {
		return saveOpCtx, fmt.Errorf("createUserOutput not found in save operation context")
	}

	userName := aws.ToString(createUserOutput.User.UserName)

	// Create each inline policy
	for _, policyNode := range u.policies {
		policyName := core.StringValue(policyNode.Fields["policyName"])
		policyDocNode := policyNode.Fields["policyDocument"]

		policyJSON, err := json.Marshal(policyDocNode)
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to marshal inline policy document: %w", err)
		}

		_, err = iamService.PutUserPolicy(ctx, &iam.PutUserPolicyInput{
			UserName:       aws.String(userName),
			PolicyName:     aws.String(policyName),
			PolicyDocument: aws.String(string(policyJSON)),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to put inline policy %s: %w", policyName, err)
		}
	}

	return saveOpCtx, nil
}

type userManagedPoliciesCreate struct {
	managedPolicyArns []*core.MappingNode
}

func (u *userManagedPoliciesCreate) Name() string {
	return "attach managed policies"
}

func (u *userManagedPoliciesCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if there are managed policies to attach
	if managedPolicyArnsNode, ok := specData.Fields["managedPolicyArns"]; ok && managedPolicyArnsNode != nil && len(managedPolicyArnsNode.Items) > 0 {
		u.managedPolicyArns = managedPolicyArnsNode.Items
		return true, saveOpCtx, nil
	}
	return false, saveOpCtx, nil
}

func (u *userManagedPoliciesCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Get the user name from the created user
	createUserOutput, ok := saveOpCtx.Data["createUserOutput"].(*iam.CreateUserOutput)
	if !ok {
		return saveOpCtx, fmt.Errorf("createUserOutput not found in save operation context")
	}

	userName := aws.ToString(createUserOutput.User.UserName)

	// Attach each managed policy
	for _, policyArnNode := range u.managedPolicyArns {
		policyArn := core.StringValue(policyArnNode)

		_, err := iamService.AttachUserPolicy(ctx, &iam.AttachUserPolicyInput{
			UserName:  aws.String(userName),
			PolicyArn: aws.String(policyArn),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to attach managed policy %s: %w", policyArn, err)
		}
	}

	return saveOpCtx, nil
}

type userPermissionsBoundaryCreate struct {
	permissionsBoundary string
}

func (u *userPermissionsBoundaryCreate) Name() string {
	return "set permissions boundary"
}

func (u *userPermissionsBoundaryCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if there is a permissions boundary to set
	if permsBoundaryNode, ok := specData.Fields["permissionsBoundary"]; ok && permsBoundaryNode != nil {
		u.permissionsBoundary = core.StringValue(permsBoundaryNode)
		return true, saveOpCtx, nil
	}
	return false, saveOpCtx, nil
}

func (u *userPermissionsBoundaryCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Get the user name from the created user
	createUserOutput, ok := saveOpCtx.Data["createUserOutput"].(*iam.CreateUserOutput)
	if !ok {
		return saveOpCtx, fmt.Errorf("createUserOutput not found in save operation context")
	}

	userName := aws.ToString(createUserOutput.User.UserName)

	_, err := iamService.PutUserPermissionsBoundary(ctx, &iam.PutUserPermissionsBoundaryInput{
		UserName:            aws.String(userName),
		PermissionsBoundary: aws.String(u.permissionsBoundary),
	})
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to set permissions boundary for user %s: %w", userName, err)
	}

	return saveOpCtx, nil
}

type userGroupMembershipCreate struct {
	groups []*core.MappingNode
}

func (u *userGroupMembershipCreate) Name() string {
	return "add user to groups"
}

func (u *userGroupMembershipCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if there are groups to add the user to
	if groupsNode, ok := specData.Fields["groups"]; ok && groupsNode != nil && len(groupsNode.Items) > 0 {
		u.groups = groupsNode.Items
		return true, saveOpCtx, nil
	}
	return false, saveOpCtx, nil
}

func (u *userGroupMembershipCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Get the user name from the created user
	createUserOutput, ok := saveOpCtx.Data["createUserOutput"].(*iam.CreateUserOutput)
	if !ok {
		return saveOpCtx, fmt.Errorf("createUserOutput not found in save operation context")
	}

	userName := aws.ToString(createUserOutput.User.UserName)

	// Add user to each group
	for _, groupNode := range u.groups {
		groupName := core.StringValue(groupNode)

		_, err := iamService.AddUserToGroup(ctx, &iam.AddUserToGroupInput{
			UserName:  aws.String(userName),
			GroupName: aws.String(groupName),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to add user to group %s: %w", groupName, err)
		}
	}

	return saveOpCtx, nil
}
