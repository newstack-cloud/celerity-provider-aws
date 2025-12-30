package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaCodeSigningConfigResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda service: %w", err)
	}

	// Try to get ARN from current spec first
	codeSigningConfigArn := ""
	arnValue, hasArn := pluginutils.GetValueByPath("$.codeSigningConfigArn", input.CurrentResourceSpec)
	if hasArn {
		codeSigningConfigArn = core.StringValue(arnValue)
	}

	// If no ARN, attempt fallback lookup by Bluelink tags
	if codeSigningConfigArn == "" {
		fallbackArn, err := l.lookupCodeSigningConfigARNByTags(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup code signing config by tags: %w", err)
		}
		if fallbackArn == "" {
			// Resource doesn't exist yet
			return &provider.ResourceGetExternalStateOutput{
				ResourceSpecState: &core.MappingNode{
					Fields: map[string]*core.MappingNode{},
				},
			}, nil
		}
		codeSigningConfigArn = fallbackArn
	}

	getCodeSigningConfigInput := &lambda.GetCodeSigningConfigInput{
		CodeSigningConfigArn: aws.String(codeSigningConfigArn),
	}

	result, err := lambdaService.GetCodeSigningConfig(ctx, getCodeSigningConfigInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda code signing config: %w", err)
	}

	// Build resource spec state from AWS response
	resourceSpecState := l.buildBaseResourceSpecState(result)

	// Add optional fields if they exist
	err = l.addOptionalConfigurationsToSpec(result, resourceSpecState.Fields)
	if err != nil {
		return nil, err
	}

	// Get tags
	listTagsInput := &lambda.ListTagsInput{
		Resource: aws.String(codeSigningConfigArn),
	}
	tagsOutput, err := lambdaService.ListTags(ctx, listTagsInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda code signing config tags: %w", err)
	}

	// Add tags if present (filtering out Bluelink provenance tags)
	if tagsOutput != nil && len(tagsOutput.Tags) > 0 {
		prefix := utils.GetBluelinkTagPrefix(input.ProviderContext.TaggingConfig())
		filteredTags := utils.FilterTagsMap(tagsOutput.Tags, prefix)

		if len(filteredTags) > 0 {
			tagNodes := make([]*core.MappingNode, 0, len(filteredTags))
			for key, value := range filteredTags {
				tagNodes = append(tagNodes, &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"key":   core.MappingNodeFromString(key),
						"value": core.MappingNodeFromString(value),
					},
				})
			}
			resourceSpecState.Fields["tags"] = &core.MappingNode{
				Items: tagNodes,
			}
		}
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}

func (l *lambdaCodeSigningConfigResourceActions) buildBaseResourceSpecState(
	output *lambda.GetCodeSigningConfigOutput,
) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"codeSigningConfigArn": core.MappingNodeFromString(
				aws.ToString(output.CodeSigningConfig.CodeSigningConfigArn),
			),
			"codeSigningConfigId": core.MappingNodeFromString(
				aws.ToString(output.CodeSigningConfig.CodeSigningConfigId),
			),
		},
	}
}

func (l *lambdaCodeSigningConfigResourceActions) addOptionalConfigurationsToSpec(
	output *lambda.GetCodeSigningConfigOutput,
	specFields map[string]*core.MappingNode,
) error {
	extractors := []pluginutils.OptionalValueExtractor[*lambda.GetCodeSigningConfigOutput]{
		{
			Name: "description",
			Condition: func(output *lambda.GetCodeSigningConfigOutput) bool {
				return output.CodeSigningConfig.Description != nil
			},
			Fields: []string{"description"},
			Values: func(output *lambda.GetCodeSigningConfigOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(aws.ToString(output.CodeSigningConfig.Description)),
				}, nil
			},
		},
		{
			Name: "allowedPublishers",
			Condition: func(output *lambda.GetCodeSigningConfigOutput) bool {
				return output.CodeSigningConfig.AllowedPublishers != nil
			},
			Fields: []string{"allowedPublishers"},
			Values: func(output *lambda.GetCodeSigningConfigOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"signingProfileVersionArns": buildSigningProfileVersionArnsNode(
								output.CodeSigningConfig.AllowedPublishers.SigningProfileVersionArns,
							),
						},
					},
				}, nil
			},
		},
		{
			Name: "codeSigningPolicies",
			Condition: func(output *lambda.GetCodeSigningConfigOutput) bool {
				return output.CodeSigningConfig.CodeSigningPolicies != nil
			},
			Fields: []string{"codeSigningPolicies"},
			Values: func(output *lambda.GetCodeSigningConfigOutput) ([]*core.MappingNode, error) {
				codeSigningPolicies := &core.MappingNode{
					Fields: map[string]*core.MappingNode{},
				}
				if output.CodeSigningConfig.CodeSigningPolicies.UntrustedArtifactOnDeployment != "" {
					codeSigningPolicies.Fields["untrustedArtifactOnDeployment"] = core.MappingNodeFromString(
						string(output.CodeSigningConfig.CodeSigningPolicies.UntrustedArtifactOnDeployment),
					)
				}
				return []*core.MappingNode{codeSigningPolicies}, nil
			},
		},
	}

	return pluginutils.RunOptionalValueExtractors(output, specFields, extractors)
}

func buildSigningProfileVersionArnsNode(arns []string) *core.MappingNode {
	arnNodes := make([]*core.MappingNode, len(arns))
	for i, arn := range arns {
		arnNodes[i] = core.MappingNodeFromString(arn)
	}
	return &core.MappingNode{
		Items: arnNodes,
	}
}

// lookupCodeSigningConfigARNByTags attempts to find a Lambda code signing config by its Bluelink provenance tags.
// This is used as a fallback when the ARN is not available (e.g., interrupted resource creation).
// Returns empty string if no matching resource is found.
func (l *lambdaCodeSigningConfigResourceActions) lookupCodeSigningConfigARNByTags(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (string, error) {
	tagFilters := utils.BuildBluelinkTagFiltersForLookup(input)
	if tagFilters == nil {
		// Tagging is not enabled, cannot perform fallback lookup
		return "", nil
	}

	taggingService, err := l.getResourceGroupTaggingService(ctx, input.ProviderContext)
	if err != nil {
		return "", err
	}

	result, err := taggingService.GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters:          tagFilters,
		ResourceTypeFilters: []string{"lambda:code-signing-config"},
	})
	if err != nil {
		return "", err
	}

	if len(result.ResourceTagMappingList) == 0 {
		return "", nil
	}

	return aws.ToString(result.ResourceTagMappingList[0].ResourceARN), nil
}
