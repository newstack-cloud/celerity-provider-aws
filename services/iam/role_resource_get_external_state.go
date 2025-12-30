package iam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamRoleResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Try to get ARN from current spec.
	// NOTE: Tag-based fallback lookup is not possible for IAM roles because the
	// AWS Resource Groups Tagging API does not support GetResources for iam:role.
	// See: https://docs.aws.amazon.com/resourcegroupstagging/latest/APIReference/supported-services.html
	arnValue, hasArn := pluginutils.GetValueByPath("$.arn", input.CurrentResourceSpec)
	if !hasArn {
		// Resource doesn't exist yet or ARN not available
		return &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{},
			},
		}, nil
	}
	arn := core.StringValue(arnValue)

	roleName, err := extractRoleNameFromARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to extract role name from ARN %s: %w", arn, err)
	}

	roleOutput, err := iamService.GetRole(
		ctx,
		&iam.GetRoleInput{
			RoleName: &roleName,
		},
	)
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			// If the role doesn't exist, return empty state
			if apiError.ErrorCode() == "NoSuchEntity" {
				return &provider.ResourceGetExternalStateOutput{
					ResourceSpecState: &core.MappingNode{
						Fields: map[string]*core.MappingNode{},
					},
				}, nil
			}
		}
		return nil, err
	}

	role := roleOutput.Role

	// URL decode the policy document
	decodedPolicyDocument, err := url.QueryUnescape(aws.ToString(role.AssumeRolePolicyDocument))
	if err != nil {
		return nil, err
	}

	// Parse the JSON policy document into a structured format
	var policyDocument interface{}
	if err := json.Unmarshal([]byte(decodedPolicyDocument), &policyDocument); err != nil {
		return nil, err
	}

	// Convert back to MappingNode structure
	policyMappingNode, err := pluginutils.AnyToMappingNode(policyDocument)
	if err != nil {
		return nil, err
	}

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString(
				aws.ToString(role.RoleName),
			),
			"assumeRolePolicyDocument": policyMappingNode,
			"arn": core.MappingNodeFromString(
				aws.ToString(role.Arn),
			),
			"roleId": core.MappingNodeFromString(
				aws.ToString(role.RoleId),
			),
		},
	}

	// Add optional fields if they exist
	if role.Description != nil {
		resourceSpecState.Fields["description"] = core.MappingNodeFromString(
			aws.ToString(role.Description),
		)
	}

	if role.MaxSessionDuration != nil {
		resourceSpecState.Fields["maxSessionDuration"] = core.MappingNodeFromInt(
			int(aws.ToInt32(role.MaxSessionDuration)),
		)
	}

	if role.Path != nil {
		resourceSpecState.Fields["path"] = core.MappingNodeFromString(
			aws.ToString(role.Path),
		)
	}

	if role.PermissionsBoundary != nil && role.PermissionsBoundary.PermissionsBoundaryArn != nil {
		resourceSpecState.Fields["permissionsBoundary"] = core.MappingNodeFromString(
			aws.ToString(role.PermissionsBoundary.PermissionsBoundaryArn),
		)
	}

	// Fetch and add tags (filtering out Bluelink provenance tags)
	if len(role.Tags) > 0 {
		userTags := extractUserIAMTags(role.Tags, input.ProviderContext)
		if len(userTags.Items) > 0 {
			resourceSpecState.Fields["tags"] = userTags
		}
	}

	// Fetch and add attached managed policy ARNs
	listAttachedPoliciesOutput, err := iamService.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
		RoleName: role.RoleName,
	})
	if err != nil {
		return nil, err
	}
	if len(listAttachedPoliciesOutput.AttachedPolicies) > 0 {
		managedPolicyArns := make([]*core.MappingNode, 0, len(listAttachedPoliciesOutput.AttachedPolicies))
		for _, policy := range listAttachedPoliciesOutput.AttachedPolicies {
			managedPolicyArns = append(managedPolicyArns, core.MappingNodeFromString(
				aws.ToString(policy.PolicyArn),
			))
		}
		resourceSpecState.Fields["managedPolicyArns"] = &core.MappingNode{Items: managedPolicyArns}
	}

	// Fetch and add inline policies as structured objects
	listPoliciesOutput, err := iamService.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
		RoleName: role.RoleName,
	})
	if err != nil {
		return nil, err
	}
	if len(listPoliciesOutput.PolicyNames) > 0 {
		policies := make([]*core.MappingNode, 0, len(listPoliciesOutput.PolicyNames))
		for _, policyName := range listPoliciesOutput.PolicyNames {
			getPolicyOutput, err := iamService.GetRolePolicy(ctx, &iam.GetRolePolicyInput{
				RoleName:   role.RoleName,
				PolicyName: aws.String(policyName),
			})
			if err != nil {
				return nil, err
			}
			// URL decode the policy document (AWS returns URL-encoded JSON)
			decodedInlinePolicyDocument, err := url.QueryUnescape(aws.ToString(getPolicyOutput.PolicyDocument))
			if err != nil {
				return nil, err
			}
			var policyDoc interface{}
			if err := json.Unmarshal([]byte(decodedInlinePolicyDocument), &policyDoc); err != nil {
				return nil, err
			}
			policyDocNode, err := pluginutils.AnyToMappingNode(policyDoc)
			if err != nil {
				return nil, err
			}
			policies = append(policies, &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"policyName":     core.MappingNodeFromString(policyName),
					"policyDocument": policyDocNode,
				},
			})
		}
		resourceSpecState.Fields["policies"] = &core.MappingNode{Items: policies}
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}
