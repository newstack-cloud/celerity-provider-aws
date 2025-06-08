package utils

import "github.com/two-hundred/celerity/libs/blueprint/core"

// TagsToMappingNode converts a map of tags from an AWS service response
// to a MappingNode suitable for use in a resource spec.
func TagsToMappingNode(tags map[string]string) *core.MappingNode {
	tagSlice := make([]*core.MappingNode, 0, len(tags))

	for key, value := range tags {
		tagSlice = append(tagSlice, &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString(key),
				"value": core.MappingNodeFromString(value),
			},
		})
	}

	return &core.MappingNode{
		Items: tagSlice,
	}
}
