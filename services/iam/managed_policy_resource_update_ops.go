package iam

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type managedPolicyVersionUpdate struct {
	policyDocument *core.MappingNode
}

func (m *managedPolicyVersionUpdate) Name() string {
	return "update policy version"
}

func (m *managedPolicyVersionUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if the policy document has changed
	if policyDocNode, ok := specData.Fields["policyDocument"]; ok && policyDocNode != nil {
		m.policyDocument = policyDocNode
		return true, saveOpCtx, nil
	}
	return false, saveOpCtx, nil
}

func (m *managedPolicyVersionUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	policyArn := saveOpCtx.ProviderUpstreamID

	// Convert the structured policy document to JSON string
	policyJSON, err := json.Marshal(m.policyDocument)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to marshal policy document: %w", err)
	}

	// Create a new policy version
	_, err = iamService.CreatePolicyVersion(ctx, &iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(policyArn),
		PolicyDocument: aws.String(string(policyJSON)),
		SetAsDefault:   true,
	})
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create new policy version for %s: %w", policyArn, err)
	}

	return saveOpCtx, nil
}

type managedPolicyTagsUpdate struct {
	tagsToAdd    []types.Tag
	tagsToRemove []string
}

func (m *managedPolicyTagsUpdate) Name() string {
	return "update tags"
}

func (m *managedPolicyTagsUpdate) Prepare(
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

	m.tagsToAdd = diffResult.ToSet
	m.tagsToRemove = diffResult.ToRemove

	return len(m.tagsToAdd) > 0 || len(m.tagsToRemove) > 0, saveOpCtx, nil
}

func (m *managedPolicyTagsUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	policyArn := saveOpCtx.ProviderUpstreamID

	// Remove tags that are no longer present
	if len(m.tagsToRemove) > 0 {
		_, err := iamService.UntagPolicy(ctx, &iam.UntagPolicyInput{
			PolicyArn: aws.String(policyArn),
			TagKeys:   m.tagsToRemove,
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to remove tags from policy %s: %w", policyArn, err)
		}
	}

	// Add new tags
	if len(m.tagsToAdd) > 0 {
		_, err := iamService.TagPolicy(ctx, &iam.TagPolicyInput{
			PolicyArn: aws.String(policyArn),
			Tags:      m.tagsToAdd,
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to add tags to policy %s: %w", policyArn, err)
		}
	}

	return saveOpCtx, nil
}
