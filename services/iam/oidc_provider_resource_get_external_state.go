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

func (i *iamOIDCProviderResourceActions) GetExternalState(
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
		fallbackArn, err := i.lookupOIDCProviderARNByTags(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup OIDC provider by tags: %w", err)
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

	// Get the OIDC provider details
	result, err := iamService.GetOpenIDConnectProvider(ctx, &iam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(arnStr),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC provider: %w", err)
	}

	// Extract the URL from the ARN
	url, err := extractUrlFromArn(arnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract URL from ARN: %w", err)
	}

	// Build the external state
	externalState := map[string]*core.MappingNode{
		"arn": core.MappingNodeFromString(arnStr),
		"url": core.MappingNodeFromString(url),
	}

	// Add client IDs if present
	if len(result.ClientIDList) > 0 {
		clientIdItems := make([]*core.MappingNode, len(result.ClientIDList))
		for i, clientId := range result.ClientIDList {
			clientIdItems[i] = core.MappingNodeFromString(clientId)
		}
		externalState["clientIdList"] = &core.MappingNode{
			Items: clientIdItems,
		}
	}

	// Add thumbprints if present
	if len(result.ThumbprintList) > 0 {
		thumbprintItems := make([]*core.MappingNode, len(result.ThumbprintList))
		for i, thumbprint := range result.ThumbprintList {
			thumbprintItems[i] = core.MappingNodeFromString(thumbprint)
		}
		externalState["thumbprintList"] = &core.MappingNode{
			Items: thumbprintItems,
		}
	}

	// Get tags
	tagsResult, err := iamService.ListOpenIDConnectProviderTags(ctx, &iam.ListOpenIDConnectProviderTagsInput{
		OpenIDConnectProviderArn: aws.String(arnStr),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	if len(tagsResult.Tags) > 0 {
		externalState["tags"] = extractUserIAMTags(tagsResult.Tags, input.ProviderContext)
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: externalState,
		},
	}, nil
}

// lookupOIDCProviderARNByTags attempts to find an IAM OIDC provider by its Bluelink provenance tags.
// This is used as a fallback when the ARN is not available (e.g., interrupted resource creation).
// Returns empty string if no matching resource is found.
func (i *iamOIDCProviderResourceActions) lookupOIDCProviderARNByTags(
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
		ResourceTypeFilters: []string{"iam:oidc-provider"},
	})
	if err != nil {
		return "", err
	}

	if len(result.ResourceTagMappingList) == 0 {
		return "", nil
	}

	return aws.ToString(result.ResourceTagMappingList[0].ResourceARN), nil
}
