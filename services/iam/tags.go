package iam

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func toIAMTag(tag *utils.Tag) types.Tag {
	return types.Tag{
		Key:   aws.String(tag.Key),
		Value: aws.String(tag.Value),
	}
}

// extractIAMTagsWithPrefix converts IAM tags to a MappingNode, filtering out Bluelink provenance tags.
// If prefix is empty, no filtering is applied.
func extractIAMTagsWithPrefix(tags []types.Tag, prefix string) *core.MappingNode {
	// Filter out Bluelink provenance tags if prefix is provided
	filteredTags := tags
	if prefix != "" {
		filteredTags = utils.FilterTags(tags, func(t types.Tag) string {
			return aws.ToString(t.Key)
		}, prefix)
	}

	tagItems := make([]*core.MappingNode, len(filteredTags))
	for i, tag := range filteredTags {
		tagItems[i] = &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString(aws.ToString(tag.Key)),
				"value": core.MappingNodeFromString(aws.ToString(tag.Value)),
			},
		}
	}
	return &core.MappingNode{
		Items: tagItems,
	}
}

// extractUserIAMTags converts IAM tags to a MappingNode, filtering out Bluelink provenance tags.
// Uses the tagging config from provider context to determine the prefix.
func extractUserIAMTags(tags []types.Tag, providerContext provider.Context) *core.MappingNode {
	prefix := utils.GetBluelinkTagPrefix(providerContext.TaggingConfig())
	return extractIAMTagsWithPrefix(tags, prefix)
}

func iamTagsFromSpecData(specData *core.MappingNode) ([]types.Tag, error) {
	var iamTags []types.Tag
	tags, hasTags := pluginutils.GetValueByPath("$.tags", specData)
	if hasTags {
		parsedTags, err := iamTagsFromValue(tags)
		if err != nil {
			return nil, err
		}
		iamTags = parsedTags
	}
	return iamTags, nil
}

func iamTagsFromValue(value *core.MappingNode) ([]types.Tag, error) {
	iamTags := []types.Tag{}
	if core.IsArrayMappingNode(value) {
		for i, item := range value.Items {
			keyNode, hasKey := pluginutils.GetValueByPath("$.key", item)
			valueNode, hasValue := pluginutils.GetValueByPath("$.value", item)
			if !hasKey || !hasValue {
				return iamTags, fmt.Errorf("invalid tag format at index %d", i)
			}
			iamTags = append(iamTags, types.Tag{
				Key:   aws.String(core.StringValue(keyNode)),
				Value: aws.String(core.StringValue(valueNode)),
			})
		}
	}
	return iamTags, nil
}

func sortTagsByKey(tags []types.Tag) []types.Tag {
	sort.Slice(tags, func(i, j int) bool {
		return aws.ToString(tags[i].Key) < aws.ToString(tags[j].Key)
	})
	return tags
}

// mergeBluelinkTagsWithIAMTags merges Bluelink system tags with user-defined IAM tags.
// User tags take precedence on key conflicts.
// Returns nil if there are no tags to set (preserves original behavior when tagging is not enabled).
func mergeBluelinkTagsWithIAMTags(
	deployInput *provider.ResourceDeployInput,
	userTags []types.Tag,
) []types.Tag {
	// Convert user IAM tags to map for merging
	userTagsMap := make(map[string]string, len(userTags))
	for _, tag := range userTags {
		if tag.Key != nil && tag.Value != nil {
			userTagsMap[*tag.Key] = *tag.Value
		}
	}

	// Merge with Bluelink tags (user tags take precedence)
	mergedMap := utils.MergeBluelinkTagsWithUserTags(deployInput, userTagsMap)

	// Return nil if no tags to set (preserves original behavior when tagging is not enabled)
	if len(mergedMap) == 0 {
		return nil
	}

	// Convert back to IAM types.Tag slice
	result := make([]types.Tag, 0, len(mergedMap))
	for k, v := range mergedMap {
		result = append(result, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}

	return sortTagsByKey(result)
}
