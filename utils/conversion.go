package utils

import (
	"sort"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// TagsToMappingNode converts a map of tags from an AWS service response
// to a MappingNode suitable for use in a resource spec.
// Tags are sorted by key for consistent ordering in comparisons.
func TagsToMappingNode(tags map[string]string) *core.MappingNode {
	return tagsMapToMappingNode(tags)
}

// UserTagsToMappingNode converts a map of tags from an AWS service response
// to a MappingNode, filtering out Bluelink provenance tags.
// Use this in GetExternalState to ensure only user-defined tags are returned.
// Tags are sorted by key for consistent ordering in comparisons.
func UserTagsToMappingNode(tags map[string]string, providerContext provider.Context) *core.MappingNode {
	prefix := GetBluelinkTagPrefix(providerContext.TaggingConfig())
	filteredTags := FilterTagsMap(tags, prefix)
	return tagsMapToMappingNode(filteredTags)
}

func tagsMapToMappingNode(tags map[string]string) *core.MappingNode {
	// Sort keys for deterministic ordering in comparisons
	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	tagSlice := make([]*core.MappingNode, 0, len(tags))
	for _, key := range keys {
		tagSlice = append(tagSlice, &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString(key),
				"value": core.MappingNodeFromString(tags[key]),
			},
		})
	}

	return &core.MappingNode{
		Items: tagSlice,
	}
}
