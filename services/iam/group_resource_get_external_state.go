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

func (i *iamGroupResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Safely get the group ARN from the resource spec
	arn, hasArn := pluginutils.GetValueByPath("$.arn", input.CurrentResourceSpec)
	if !hasArn {
		return nil, fmt.Errorf("ARN is required for get external state operation")
	}

	arnStr := core.StringValue(arn)
	if arnStr == "" {
		return nil, fmt.Errorf("ARN is required for get external state operation")
	}

	// Extract group name from ARN
	groupName, err := extractGroupNameFromARN(arnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract group name from ARN: %w", err)
	}

	// Get group details
	getGroupOutput, err := iamService.GetGroup(ctx, &iam.GetGroupInput{
		GroupName: aws.String(groupName),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			// If the group doesn't exist, return empty state
			if apiError.ErrorCode() == "NoSuchEntity" {
				return &provider.ResourceGetExternalStateOutput{
					ResourceSpecState: &core.MappingNode{
						Fields: map[string]*core.MappingNode{},
					},
				}, nil
			}
		}
		return nil, fmt.Errorf("failed to get group %s: %w", groupName, err)
	}

	group := getGroupOutput.Group

	// Build the external state safely
	externalState := map[string]*core.MappingNode{
		"arn":       core.MappingNodeFromString(aws.ToString(group.Arn)),
		"groupId":   core.MappingNodeFromString(aws.ToString(group.GroupId)),
		"groupName": core.MappingNodeFromString(aws.ToString(group.GroupName)),
		"path":      core.MappingNodeFromString(aws.ToString(group.Path)),
	}

	// Get managed policies
	managedPolicies, err := i.getManagedPolicies(ctx, iamService, groupName)
	if err != nil {
		return nil, fmt.Errorf("failed to get managed policies: %w", err)
	}
	if len(managedPolicies) > 0 {
		externalState["managedPolicyArns"] = &core.MappingNode{
			Items: managedPolicies,
		}
	}

	// Get inline policies
	inlinePolicies, err := i.getInlinePolicies(ctx, iamService, groupName)
	if err != nil {
		return nil, fmt.Errorf("failed to get inline policies: %w", err)
	}
	if len(inlinePolicies) > 0 {
		externalState["policies"] = &core.MappingNode{
			Items: inlinePolicies,
		}
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: externalState,
		},
	}, nil
}

func (i *iamGroupResourceActions) getManagedPolicies(
	ctx context.Context,
	iamService iamservice.Service,
	groupName string,
) ([]*core.MappingNode, error) {
	result, err := iamService.ListAttachedGroupPolicies(ctx, &iam.ListAttachedGroupPoliciesInput{
		GroupName: aws.String(groupName),
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

func (i *iamGroupResourceActions) getInlinePolicies(
	ctx context.Context,
	iamService iamservice.Service,
	groupName string,
) ([]*core.MappingNode, error) {
	// First, list all inline policy names
	listResult, err := iamService.ListGroupPolicies(ctx, &iam.ListGroupPoliciesInput{
		GroupName: aws.String(groupName),
	})
	if err != nil {
		return nil, err
	}

	var policies []*core.MappingNode
	for _, policyName := range listResult.PolicyNames {
		// Get the policy document for each policy
		policyResult, err := iamService.GetGroupPolicy(ctx, &iam.GetGroupPolicyInput{
			GroupName:  aws.String(groupName),
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
