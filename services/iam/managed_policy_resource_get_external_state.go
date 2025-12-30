package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamManagedPolicyResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Try to get ARN from current spec first
	arnStr := ""
	arn, hasArn := pluginutils.GetValueByPath("$.arn", input.CurrentResourceSpec)
	if hasArn {
		arnStr = core.StringValue(arn)
	}

	// If no ARN, attempt fallback lookup by Bluelink tags
	if arnStr == "" {
		fallbackArn, err := i.lookupManagedPolicyARNByTags(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup managed policy by tags: %w", err)
		}
		if fallbackArn == "" {
			// Resource doesn't exist yet
			return &provider.ResourceGetExternalStateOutput{
				ResourceSpecState: &core.MappingNode{
					Fields: map[string]*core.MappingNode{},
				},
			}, nil
		}
		arnStr = fallbackArn
	}

	// Get the managed policy
	getPolicyOutput, err := iamService.GetPolicy(ctx, &iam.GetPolicyInput{
		PolicyArn: aws.String(arnStr),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get IAM managed policy %s: %w", arnStr, err)
	}

	// Get the policy tags
	listPolicyTagsOutput, err := iamService.ListPolicyTags(ctx, &iam.ListPolicyTagsInput{
		PolicyArn: aws.String(arnStr),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tags for IAM managed policy %s: %w", arnStr, err)
	}

	// Build the external state
	externalState := map[string]*core.MappingNode{
		"policyName": core.MappingNodeFromString(aws.ToString(getPolicyOutput.Policy.PolicyName)),
		"path":       core.MappingNodeFromString(aws.ToString(getPolicyOutput.Policy.Path)),
		"arn":        core.MappingNodeFromString(aws.ToString(getPolicyOutput.Policy.Arn)),
		"id":         core.MappingNodeFromString(aws.ToString(getPolicyOutput.Policy.PolicyId)),
	}

	// Add optional fields if they exist
	if getPolicyOutput.Policy.Description != nil {
		externalState["description"] = core.MappingNodeFromString(aws.ToString(getPolicyOutput.Policy.Description))
	}

	// Add computed fields
	if getPolicyOutput.Policy.AttachmentCount != nil {
		externalState["attachmentCount"] = core.MappingNodeFromInt(int(*getPolicyOutput.Policy.AttachmentCount))
	}
	if getPolicyOutput.Policy.CreateDate != nil {
		externalState["createDate"] = core.MappingNodeFromString(getPolicyOutput.Policy.CreateDate.Format("2006-01-02T15:04:05Z"))
	}
	if getPolicyOutput.Policy.DefaultVersionId != nil {
		externalState["defaultVersionId"] = core.MappingNodeFromString(aws.ToString(getPolicyOutput.Policy.DefaultVersionId))
	}
	if getPolicyOutput.Policy.IsAttachable {
		externalState["isAttachable"] = core.MappingNodeFromBool(getPolicyOutput.Policy.IsAttachable)
	}
	if getPolicyOutput.Policy.PermissionsBoundaryUsageCount != nil {
		externalState["permissionsBoundaryUsageCount"] = core.MappingNodeFromInt(int(*getPolicyOutput.Policy.PermissionsBoundaryUsageCount))
	}
	if getPolicyOutput.Policy.UpdateDate != nil {
		externalState["updateDate"] = core.MappingNodeFromString(getPolicyOutput.Policy.UpdateDate.Format("2006-01-02T15:04:05Z"))
	}

	// Add tags if they exist (filtering out Bluelink provenance tags)
	if len(listPolicyTagsOutput.Tags) > 0 {
		externalState["tags"] = extractUserIAMTags(listPolicyTagsOutput.Tags, input.ProviderContext)
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: externalState,
		},
	}, nil
}

// lookupManagedPolicyARNByTags attempts to find an IAM managed policy by its Bluelink provenance tags.
// This is used as a fallback when the ARN is not available (e.g., interrupted resource creation).
// Returns empty string if no matching resource is found.
func (i *iamManagedPolicyResourceActions) lookupManagedPolicyARNByTags(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (string, error) {
	tagFilters := utils.BuildBluelinkTagFiltersForLookup(input)
	if tagFilters == nil {
		// Tagging is not enabled, cannot perform fallback lookup
		return "", nil
	}

	taggingService, err := i.getResourceGroupTaggingService(ctx, input.ProviderContext)
	if err != nil {
		return "", err
	}

	result, err := taggingService.GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters:          tagFilters,
		ResourceTypeFilters: []string{"iam:policy"},
	})
	if err != nil {
		return "", err
	}

	if len(result.ResourceTagMappingList) == 0 {
		return "", nil
	}

	return aws.ToString(result.ResourceTagMappingList[0].ResourceARN), nil
}
