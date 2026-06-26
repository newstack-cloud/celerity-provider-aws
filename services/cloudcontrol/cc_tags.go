package cloudcontrol

import (
	"sort"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
)

func extractUserTags(node *core.MappingNode, shape TagShape) map[string]string {
	tags := map[string]string{}
	if node == nil {
		return tags
	}

	switch shape {
	case TagShapeKeyValueList:
		for _, item := range node.Items {
			if item == nil || item.Fields == nil {
				continue
			}
			key := core.StringValue(item.Fields["key"])
			if key == "" {
				continue
			}
			tags[key] = core.StringValue(item.Fields["value"])
		}
	case TagShapeStringMap:
		for key, value := range node.Fields {
			tags[key] = core.StringValue(value)
		}
	}

	return tags
}

// Keys are sorted for deterministic output.
func buildTagNode(tags map[string]string, shape TagShape) *core.MappingNode {
	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	switch shape {
	case TagShapeKeyValueList:
		items := make([]*core.MappingNode, 0, len(keys))
		for _, key := range keys {
			items = append(items, &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"key":   core.MappingNodeFromString(key),
					"value": core.MappingNodeFromString(tags[key]),
				},
			})
		}
		return &core.MappingNode{Items: items}
	case TagShapeStringMap:
		fields := make(map[string]*core.MappingNode, len(keys))
		for _, key := range keys {
			fields[key] = core.MappingNodeFromString(tags[key])
		}
		return &core.MappingNode{Fields: fields}
	}

	return nil
}
