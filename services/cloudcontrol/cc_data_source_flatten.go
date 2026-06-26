package cloudcontrol

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// FlattenForDataSource flattens a typed spec node into the data source export shape,
// guided by the schema (the export contract is flat, scalars + arrays-of-primitives):
//   - nested objects -> dot-notation scalar keys (e.g. "provisionedThroughput.readCapacityUnits")
//   - arrays (of scalars or objects) -> kept whole under a single key
//   - maps -> serialised to a JSON string (the export contract has no map/object type)
//   - scalars -> mapped directly
//
// It is schema-aware so a map (dynamic keys) is not mistaken for an object and
// dot-flattened by its data; the codegen derives export fields by the same rules.
func FlattenForDataSource(
	node *core.MappingNode,
	schema *provider.ResourceDefinitionsSchema,
) map[string]*core.MappingNode {
	out := map[string]*core.MappingNode{}
	flattenSchemaAware(node, schema, "", out)
	return out
}

func flattenSchemaAware(
	node *core.MappingNode,
	schema *provider.ResourceDefinitionsSchema,
	prefix string,
	out map[string]*core.MappingNode,
) {
	if node == nil {
		return
	}

	// Descend into objects only (guided by the schema, not the value shape, so maps
	// are not treated as objects).
	if schema != nil && schema.Type == provider.ResourceDefinitionsSchemaTypeObject && node.Fields != nil {
		for name, child := range node.Fields {
			key := name
			if prefix != "" {
				key = prefix + "." + name
			}
			var childSchema *provider.ResourceDefinitionsSchema
			if schema.Attributes != nil {
				childSchema = schema.Attributes[name]
			}
			flattenSchemaAware(child, childSchema, key, out)
		}
		return
	}

	if prefix == "" {
		// Top-level value that is not an object: nothing to flatten.
		return
	}

	// Maps have no flat representation, so they are exported as a JSON string.
	if schema != nil && schema.Type == provider.ResourceDefinitionsSchemaTypeMap {
		if jsonStr, err := mappingNodeToCFNJSON(node); err == nil {
			out[prefix] = core.MappingNodeFromString(jsonStr)
		}
		return
	}

	// Arrays, scalars, unions and unknown-schema nodes are kept whole as a leaf.
	out[prefix] = node
}

// Restricts a flattened field map to the declared export fields, so the returned data
// matches the data source schema exactly.
func selectExports(
	flattened map[string]*core.MappingNode,
	exportFields map[string]*provider.DataSourceSpecSchema,
) map[string]*core.MappingNode {
	data := make(map[string]*core.MappingNode, len(exportFields))
	for name := range exportFields {
		if value, ok := flattened[name]; ok && value != nil {
			data[name] = value
		}
	}
	return data
}
