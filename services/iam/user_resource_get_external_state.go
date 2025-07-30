package iam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/smithy-go"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamUserResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Get the user ARN from the resource spec
	arn := core.StringValue(input.CurrentResourceSpec.Fields["arn"])
	if arn == "" {
		return nil, fmt.Errorf("ARN is required for get external state operation")
	}

	// Extract user name from ARN
	userName, err := extractUserNameFromARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user name from ARN: %w", err)
	}

	// Get the user details
	getUserOutput, err := iamService.GetUser(ctx, &iam.GetUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			// If the user doesn't exist, return empty state
			if apiError.ErrorCode() == "NoSuchEntity" {
				return &provider.ResourceGetExternalStateOutput{
					ResourceSpecState: &core.MappingNode{
						Fields: map[string]*core.MappingNode{},
					},
				}, nil
			}
		}
		return nil, fmt.Errorf("failed to get user %s: %w", userName, err)
	}

	user := getUserOutput.User

	// Build the external state
	externalState := map[string]*core.MappingNode{
		"arn":      core.MappingNodeFromString(aws.ToString(user.Arn)),
		"userId":   core.MappingNodeFromString(aws.ToString(user.UserId)),
		"userName": core.MappingNodeFromString(aws.ToString(user.UserName)),
		"path":     core.MappingNodeFromString(aws.ToString(user.Path)),
	}

	// Get managed policies
	managedPolicies, err := i.getManagedPolicies(ctx, iamService, userName)
	if err != nil {
		return nil, fmt.Errorf("failed to get managed policies: %w", err)
	}
	if len(managedPolicies) > 0 {
		externalState["managedPolicyArns"] = &core.MappingNode{
			Items: managedPolicies,
		}
	}

	// Get inline policies
	inlinePolicies, err := i.getInlinePolicies(ctx, iamService, userName)
	if err != nil {
		return nil, fmt.Errorf("failed to get inline policies: %w", err)
	}
	if len(inlinePolicies) > 0 {
		externalState["policies"] = &core.MappingNode{
			Items: inlinePolicies,
		}
	}

	// Get permissions boundary
	if user.PermissionsBoundary != nil && user.PermissionsBoundary.PermissionsBoundaryArn != nil {
		externalState["permissionsBoundary"] = core.MappingNodeFromString(aws.ToString(user.PermissionsBoundary.PermissionsBoundaryArn))
	}

	// Get tags
	userTags, err := i.getUserTags(ctx, iamService, userName)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tags: %w", err)
	}
	if len(userTags) > 0 {
		externalState["tags"] = &core.MappingNode{
			Items: userTags,
		}
	}

	// Get groups
	groups, err := i.getUserGroups(ctx, iamService, userName)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}
	if len(groups) > 0 {
		externalState["groups"] = &core.MappingNode{
			Items: groups,
		}
	}

	// Check for login profile (we don't retrieve password for security)
	hasLoginProfile, err := i.hasLoginProfile(ctx, iamService, userName)
	if err != nil {
		return nil, fmt.Errorf("failed to check login profile: %w", err)
	}
	if hasLoginProfile {
		// We can't retrieve the actual password, so we just indicate that a login profile exists
		externalState["loginProfile"] = &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"password":              core.MappingNodeFromString("<hidden>"), // Password is hidden for security
				"passwordResetRequired": core.MappingNodeFromBool(false),        // Default value
			},
		}
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: externalState,
		},
	}, nil
}

func (i *iamUserResourceActions) getManagedPolicies(
	ctx context.Context,
	iamService iamservice.Service,
	userName string,
) ([]*core.MappingNode, error) {
	result, err := iamService.ListAttachedUserPolicies(ctx, &iam.ListAttachedUserPoliciesInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return nil, err
	}

	var policies []*core.MappingNode
	for _, policy := range result.AttachedPolicies {
		policies = append(policies, core.MappingNodeFromString(aws.ToString(policy.PolicyArn)))
	}

	return policies, nil
}

func (i *iamUserResourceActions) getInlinePolicies(
	ctx context.Context,
	iamService iamservice.Service,
	userName string,
) ([]*core.MappingNode, error) {
	// First, list all inline policy names
	listResult, err := iamService.ListUserPolicies(ctx, &iam.ListUserPoliciesInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return nil, err
	}

	var policies []*core.MappingNode
	for _, policyName := range listResult.PolicyNames {
		// Get the policy document for each policy
		policyResult, err := iamService.GetUserPolicy(ctx, &iam.GetUserPolicyInput{
			UserName:   aws.String(userName),
			PolicyName: aws.String(policyName),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get policy %s: %w", policyName, err)
		}

		// Parse the policy document JSON
		var policyDoc map[string]any
		if err := json.Unmarshal([]byte(aws.ToString(policyResult.PolicyDocument)), &policyDoc); err != nil {
			return nil, fmt.Errorf("failed to parse policy document for %s: %w", policyName, err)
		}

		// Convert to MappingNode
		policyDocNode, err := pluginutils.AnyToMappingNode(policyDoc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert policy document to mapping node: %w", err)
		}

		policyNode := &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"policyName":     core.MappingNodeFromString(policyName),
				"policyDocument": policyDocNode,
			},
		}

		policies = append(policies, policyNode)
	}

	return policies, nil
}

func (i *iamUserResourceActions) getUserTags(
	ctx context.Context,
	iamService iamservice.Service,
	userName string,
) ([]*core.MappingNode, error) {
	result, err := iamService.ListUserTags(ctx, &iam.ListUserTagsInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return nil, err
	}

	var tags []*core.MappingNode
	for _, tag := range result.Tags {
		tagNode := &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString(aws.ToString(tag.Key)),
				"value": core.MappingNodeFromString(aws.ToString(tag.Value)),
			},
		}
		tags = append(tags, tagNode)
	}

	return tags, nil
}

func (i *iamUserResourceActions) getUserGroups(
	ctx context.Context,
	iamService iamservice.Service,
	userName string,
) ([]*core.MappingNode, error) {
	result, err := iamService.ListGroupsForUser(ctx, &iam.ListGroupsForUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return nil, err
	}

	var groups []*core.MappingNode
	for _, group := range result.Groups {
		groups = append(groups, core.MappingNodeFromString(aws.ToString(group.GroupName)))
	}

	return groups, nil
}

func (i *iamUserResourceActions) hasLoginProfile(
	ctx context.Context,
	iamService iamservice.Service,
	userName string,
) (bool, error) {
	_, err := iamService.GetLoginProfile(ctx, &iam.GetLoginProfileInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		// Check if the error is because the login profile doesn't exist
		var apiError smithy.APIError
		if errors.As(err, &apiError) && apiError.ErrorCode() == "NoSuchEntity" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
