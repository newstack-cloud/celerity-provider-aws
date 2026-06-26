package cloudcontrol

import (
	"unicode"
	"unicode/utf8"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Bluelink generated schemas use camelCase field names while Cloud Control expects
// CloudFormation PascalCase property names. For most fields the two differ only in
// the case of the first character, so a first-character case flip round-trips
// exactly. Fields whose CFN name cannot be recovered by a flip
// (for example, "sseSpecification" <-> "SSESpecification") carry an explicit
// override keyed by camelCase path. The translation is applied only to object
// attribute keys (property names); map keys and array items hold user data and are
// left untouched, which is why translation is schema-aware.

// Carries the per-type field-name overrides and free-form JSON field set through a
// translation walk, keyed by camelCase field path (with "*" for array/map descent).
type nameTranslator struct {
	overrides   map[string]string
	jsonStrings map[string]bool
}

func fieldSet(paths []string) map[string]bool {
	if len(paths) == 0 {
		return nil
	}
	set := make(map[string]bool, len(paths))
	for _, path := range paths {
		set[path] = true
	}
	return set
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[size:]
}

// SpecToCFN translates a typed spec node into the CloudFormation property shape,
// translating the resource field names to the CloudFormation property names.
func SpecToCFN(
	node *core.MappingNode,
	schema *provider.ResourceDefinitionsSchema,
	meta CCResourceMeta,
) *core.MappingNode {
	t := &nameTranslator{
		overrides:   meta.FieldNameOverrides,
		jsonStrings: fieldSet(meta.JSONStringFields),
	}
	return t.translateNode(node, schema, "", true)
}

// CFNToSpec translates a CloudFormation property node into the typed spec shape,
// translating the CloudFormation property names to the provider's resource field names.
func CFNToSpec(
	node *core.MappingNode,
	schema *provider.ResourceDefinitionsSchema,
	meta CCResourceMeta,
) *core.MappingNode {
	t := &nameTranslator{
		overrides:   meta.FieldNameOverrides,
		jsonStrings: fieldSet(meta.JSONStringFields),
	}
	return t.translateNode(node, schema, "", false)
}

func (t *nameTranslator) translateNode(
	node *core.MappingNode,
	schema *provider.ResourceDefinitionsSchema,
	path string,
	toCFN bool,
) *core.MappingNode {
	if node == nil {
		return nil
	}

	schema = effectiveSchema(schema, node)

	if node.Fields != nil {
		return t.translateFields(node, schema, path, toCFN)
	}

	if node.Items != nil {
		var itemSchema *provider.ResourceDefinitionsSchema
		if schema != nil {
			itemSchema = schema.Items
		}
		itemPath := joinNamePath(path, "*")
		items := make([]*core.MappingNode, len(node.Items))
		for i, item := range node.Items {
			items[i] = t.translateNode(item, itemSchema, itemPath, toCFN)
		}
		return &core.MappingNode{Items: items}
	}

	if !toCFN {
		// Cloud Control JSON decodes every number as a float, so coerce scalars
		// back to the schema's declared numeric type to avoid unexpected drift.
		return coerceScalar(node, schema)
	}
	return node
}

func coerceScalar(
	node *core.MappingNode,
	schema *provider.ResourceDefinitionsSchema,
) *core.MappingNode {
	if node == nil || node.Scalar == nil || schema == nil {
		return node
	}

	switch schema.Type {
	case provider.ResourceDefinitionsSchemaTypeInteger:
		if node.Scalar.FloatValue != nil {
			return core.MappingNodeFromInt(int(*node.Scalar.FloatValue))
		}
	case provider.ResourceDefinitionsSchemaTypeFloat:
		if node.Scalar.IntValue != nil {
			return core.MappingNodeFromFloat(float64(*node.Scalar.IntValue))
		}
	}

	return node
}

func (t *nameTranslator) translateFields(
	node *core.MappingNode,
	schema *provider.ResourceDefinitionsSchema,
	path string,
	toCFN bool,
) *core.MappingNode {
	isObject := schema != nil && schema.Type == provider.ResourceDefinitionsSchemaTypeObject

	out := &core.MappingNode{Fields: make(map[string]*core.MappingNode, len(node.Fields))}
	for key, value := range node.Fields {
		newKey := key
		childPath := path
		var childSchema *provider.ResourceDefinitionsSchema

		switch {
		case isObject && toCFN:
			childSchema = schema.Attributes[key]
			childPath = joinNamePath(path, key)
			newKey = t.cfnName(childPath, key)
		case isObject && !toCFN:
			camel := t.camelName(schema, path, key)
			newKey = camel
			childSchema = schema.Attributes[camel]
			childPath = joinNamePath(path, camel)
		case schema != nil && schema.Type == provider.ResourceDefinitionsSchemaTypeMap:
			// Map keys are user data, preserve the original key.
			childSchema = schema.MapValues
			childPath = joinNamePath(path, "*")
		}

		if t.jsonStrings[childPath] {
			out.Fields[newKey] = translateJSONStringField(value, toCFN)
			continue
		}
		out.Fields[newKey] = t.translateNode(value, childSchema, childPath, toCFN)
	}
	return out
}

// Bridges a free-form object encoded as a JSON string in the spec and the JSON object
// Cloud Control expects in desired state. Going to CFN it parses the string into an object;
// coming back it serialises the object to a string. The value is passed through unchanged
// if it is not in the expected shape, so a missing or already-correct value never breaks
// the walk.
func translateJSONStringField(value *core.MappingNode, toCFN bool) *core.MappingNode {
	if value == nil {
		return nil
	}

	if toCFN {
		if value.Scalar == nil || value.Scalar.StringValue == nil {
			return value
		}
		parsed, err := cfnJSONToMappingNode(*value.Scalar.StringValue)
		if err != nil {
			return value
		}
		return parsed
	}

	if value.Fields == nil && value.Items == nil {
		return value
	}

	serialised, err := mappingNodeToCFNJSON(value)
	if err != nil {
		return value
	}

	return core.MappingNodeFromString(serialised)
}

func (t *nameTranslator) cfnName(childPath, camelKey string) string {
	if cfn, ok := t.overrides[childPath]; ok {
		return cfn
	}
	return upperFirst(camelKey)
}

// Recovers the camelCase field name for a CloudFormation property key at the given object
// path, inverting any override and otherwise flipping the first character.
func (t *nameTranslator) camelName(
	schema *provider.ResourceDefinitionsSchema,
	path, cfnKey string,
) string {
	if schema != nil {
		for camel := range schema.Attributes {
			if t.cfnName(joinNamePath(path, camel), camel) == cfnKey {
				return camel
			}
		}
	}
	return lowerFirst(cfnKey)
}

func joinNamePath(path, segment string) string {
	if path == "" {
		return segment
	}
	return path + "." + segment
}

// Resolves a union schema down to the branch matching the node's
// shape so the rest of the walk can treat it as a concrete object/array/scalar.
func effectiveSchema(
	schema *provider.ResourceDefinitionsSchema,
	node *core.MappingNode,
) *provider.ResourceDefinitionsSchema {
	if schema == nil || schema.Type != provider.ResourceDefinitionsSchemaTypeUnion {
		return schema
	}

	for _, branch := range schema.OneOf {
		if branch == nil {
			continue
		}
		switch {
		case node.Fields != nil &&
			(branch.Type == provider.ResourceDefinitionsSchemaTypeObject ||
				branch.Type == provider.ResourceDefinitionsSchemaTypeMap):
			return branch
		case node.Items != nil && branch.Type == provider.ResourceDefinitionsSchemaTypeArray:
			return branch
		}
	}

	return schema
}
