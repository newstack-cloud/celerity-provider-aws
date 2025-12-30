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

func (i *iamSAMLProviderResourceActions) GetExternalState(
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
		fallbackArn, err := i.lookupSAMLProviderARNByTags(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup SAML provider by tags: %w", err)
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

	// Get the SAML provider details
	result, err := iamService.GetSAMLProvider(ctx, &iam.GetSAMLProviderInput{
		SAMLProviderArn: aws.String(arnStr),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get SAML provider: %w", err)
	}

	// Extract the name from the ARN
	name, err := extractNameFromArn(arnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract name from ARN: %w", err)
	}

	// Build the external state
	externalState := map[string]*core.MappingNode{
		"arn":  core.MappingNodeFromString(arnStr),
		"name": core.MappingNodeFromString(name),
	}

	// Add SAML metadata document if present
	if result.SAMLMetadataDocument != nil {
		externalState["samlMetadataDocument"] = core.MappingNodeFromString(aws.ToString(result.SAMLMetadataDocument))
	}

	// Get tags
	tagsResult, err := iamService.ListSAMLProviderTags(ctx, &iam.ListSAMLProviderTagsInput{
		SAMLProviderArn: aws.String(arnStr),
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

// lookupSAMLProviderARNByTags attempts to find an IAM SAML provider by its Bluelink provenance tags.
// This is used as a fallback when the ARN is not available (e.g., interrupted resource creation).
// Returns empty string if no matching resource is found.
func (i *iamSAMLProviderResourceActions) lookupSAMLProviderARNByTags(
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
		ResourceTypeFilters: []string{"iam:saml-provider"},
	})
	if err != nil {
		return "", err
	}

	if len(result.ResourceTagMappingList) == 0 {
		return "", nil
	}

	return aws.ToString(result.ResourceTagMappingList[0].ResourceARN), nil
}
