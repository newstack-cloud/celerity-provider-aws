package cloudcontrol

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Returns a copy of a spec mapping node with every node whose
// schema is marked Computed removed. This excludes Cloud Control readOnly properties
// (and the internal __cc* fields) from desired state and update patches,
// which Cloud Control rejects.
func stripComputedFields(
	node *core.MappingNode,
	schema *provider.ResourceDefinitionsSchema,
) *core.MappingNode {
	if node == nil {
		return nil
	}

	schema = effectiveSchema(schema, node)

	if node.Fields != nil && schema != nil &&
		schema.Type == provider.ResourceDefinitionsSchemaTypeObject {
		out := &core.MappingNode{Fields: make(map[string]*core.MappingNode, len(node.Fields))}
		for key, value := range node.Fields {
			child := schema.Attributes[key]
			if child != nil && child.Computed {
				continue
			}
			out.Fields[key] = stripComputedFields(value, child)
		}
		return out
	}

	if node.Items != nil {
		var itemSchema *provider.ResourceDefinitionsSchema
		if schema != nil {
			itemSchema = schema.Items
		}
		items := make([]*core.MappingNode, len(node.Items))
		for i, item := range node.Items {
			items[i] = stripComputedFields(item, itemSchema)
		}
		return &core.MappingNode{Items: items}
	}

	if node.Fields != nil {
		// Map or object with no schema.
		var valueSchema *provider.ResourceDefinitionsSchema
		if schema != nil {
			valueSchema = schema.MapValues
		}
		out := &core.MappingNode{Fields: make(map[string]*core.MappingNode, len(node.Fields))}
		for key, value := range node.Fields {
			out.Fields[key] = stripComputedFields(value, valueSchema)
		}
		return out
	}

	return node
}
