package utils

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	rgtatypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// MergeBluelinkTagsWithUserTags extracts Bluelink system tags from input,
// converts to AWS format, and merges with user-provided tags.
// User tags take precedence on key conflicts.
// Returns map[string]string suitable for AWS SDK calls.
func MergeBluelinkTagsWithUserTags(
	input *provider.ResourceDeployInput,
	userTags map[string]string,
) map[string]string {
	bluelinkTags := providerv1.GetBluelinkTags(input)
	awsTags := pluginutils.ToAWSTags(bluelinkTags)

	userAWSTags := convertUserTagsToAWSTags(userTags)
	merged := pluginutils.MergeAWSTags(awsTags, userAWSTags)

	return convertAWSTagsToMap(merged)
}

// convertUserTagsToAWSTags converts a map of user tags to []pluginutils.AWSTag.
func convertUserTagsToAWSTags(userTags map[string]string) []pluginutils.AWSTag {
	result := make([]pluginutils.AWSTag, 0, len(userTags))
	for k, v := range userTags {
		key, value := k, v
		result = append(result, pluginutils.AWSTag{
			Key:   &key,
			Value: &value,
		})
	}
	return result
}

// convertAWSTagsToMap converts []pluginutils.AWSTag to map[string]string.
func convertAWSTagsToMap(tags []pluginutils.AWSTag) map[string]string {
	result := make(map[string]string, len(tags))
	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil {
			result[*tag.Key] = *tag.Value
		}
	}
	return result
}

// BuildBluelinkTagFiltersForLookup creates Resource Groups Tagging API tag filters
// for looking up a resource by its Bluelink provenance tags.
// This is used for fallback lookups when the external ID (ARN) is not available.
func BuildBluelinkTagFiltersForLookup(
	input *provider.ResourceGetExternalStateInput,
) []rgtatypes.TagFilter {
	taggingConfig := input.ProviderContext.TaggingConfig()
	if taggingConfig == nil || !taggingConfig.Enabled {
		return nil
	}

	prefix := GetBluelinkTagPrefix(taggingConfig)

	return []rgtatypes.TagFilter{
		{
			Key:    aws.String(fmt.Sprintf("%sinstance-id", prefix)),
			Values: []string{input.InstanceID},
		},
		{
			Key:    aws.String(fmt.Sprintf("%sresource-name", prefix)),
			Values: []string{input.ResourceName},
		},
		{
			Key:    aws.String(fmt.Sprintf("%sprovisioned-by", prefix)),
			Values: []string{"bluelink"},
		},
	}
}

// DefaultBluelinkTagPrefix is the default prefix for Bluelink provenance tags.
const DefaultBluelinkTagPrefix = "bluelink:"

// GetBluelinkTagPrefix returns the tag prefix from tagging config, or the default if not set.
func GetBluelinkTagPrefix(taggingConfig *provider.TaggingConfig) string {
	if taggingConfig == nil || taggingConfig.Prefix == "" {
		return DefaultBluelinkTagPrefix
	}
	return taggingConfig.Prefix
}

// IsBluelinkTag checks if a tag key is a Bluelink provenance tag.
func IsBluelinkTag(key, prefix string) bool {
	return len(key) >= len(prefix) && key[:len(prefix)] == prefix
}

// FilterTags is a generic function that filters out Bluelink provenance tags from a slice.
// It takes a slice of tags, a function to extract the key from each tag, and the prefix to filter.
// Returns a new slice containing only non-Bluelink tags.
func FilterTags[T any](tags []T, getKey func(T) string, prefix string) []T {
	if len(tags) == 0 {
		return tags
	}

	result := make([]T, 0, len(tags))
	for _, tag := range tags {
		key := getKey(tag)
		if !IsBluelinkTag(key, prefix) {
			result = append(result, tag)
		}
	}
	return result
}

// FilterTagsMap filters out Bluelink provenance tags from a map.
// Returns a new map containing only non-Bluelink tags.
func FilterTagsMap(tags map[string]string, prefix string) map[string]string {
	if len(tags) == 0 {
		return tags
	}

	result := make(map[string]string, len(tags))
	for key, value := range tags {
		if !IsBluelinkTag(key, prefix) {
			result[key] = value
		}
	}
	return result
}
