package main

import (
	"fmt"
	"strings"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/blueprint/substitutions"
)

const blueprintVersion = "2025-11-02"

// Selects how much of a resource's spec an example populates.
type exampleVariant int

const (
	variantBasic exampleVariant = iota
	variantComplete
)

func (v exampleVariant) name() string {
	if v == variantBasic {
		return "basic"
	}
	return "complete"
}

func (v exampleVariant) description(label string) string {
	if v == variantBasic {
		return fmt.Sprintf("A basic %s with the minimum configuration.", label)
	}
	return fmt.Sprintf("A %s configured with the full set of available properties.", label)
}

// Constructs a single-resource blueprint that exercises the
// generated schema for the given variant, choosing representative values.
func buildExampleBlueprint(res *irResource, variant exampleVariant) *schema.Blueprint {
	name := resourceExampleName(res)
	spec := buildExampleObject(res.Schema, variant, res.BlueprintType)

	resources := &schema.ResourceMap{
		Values: map[string]*schema.Resource{
			name: {
				Type: &schema.ResourceTypeWrapper{Value: res.BlueprintType},
				Metadata: &schema.Metadata{
					DisplayName: literalString(fmt.Sprintf("%s %s", res.Label, variant.name())),
				},
				Spec: spec,
			},
		},
	}

	return &schema.Blueprint{
		Version:   core.ScalarFromString(blueprintVersion),
		Resources: resources,
	}
}

func buildExampleObject(
	objectSchema *irSchema,
	variant exampleVariant,
	blueprintType string,
) *core.MappingNode {
	fields := map[string]*core.MappingNode{}
	required := toSet(objectSchema.Required)

	for _, attr := range objectSchema.Attributes {
		if attr.Schema.Computed || strings.HasPrefix(attr.Name, "__") {
			continue
		}
		if variant == variantBasic && !required[attr.Name] {
			continue
		}
		if value := buildExampleValue(attr.Name, attr.Schema, variant, blueprintType); value != nil {
			fields[attr.Name] = value
		}
	}

	// A basic example with no required fields still needs at least one property to
	// be a meaningful blueprint. Fall back to the first settable scalar.
	if variant == variantBasic && len(fields) == 0 {
		if name, value := firstSettableScalar(objectSchema, blueprintType); value != nil {
			fields[name] = value
		}
	}

	return &core.MappingNode{Fields: fields}
}

func buildExampleValue(
	name string,
	fieldSchema *irSchema,
	variant exampleVariant,
	blueprintType string,
) *core.MappingNode {
	if seeded := seededExampleValue(blueprintType, name); seeded != nil {
		return seeded
	}

	switch fieldSchema.Type {
	case "object":
		return buildExampleObject(fieldSchema, variantComplete, blueprintType)
	case "array":
		if fieldSchema.Items == nil {
			return nil
		}
		item := buildExampleValue(singular(name), fieldSchema.Items, variantComplete, blueprintType)
		if item == nil {
			return nil
		}
		return &core.MappingNode{Items: []*core.MappingNode{item}}
	case "map":
		// A map with no declared value schema is a free-form object. Seeded
		// values are handled above. Without a seed example, emit a small illustrative
		// object so the field is demonstrated rather than skipped.
		if fieldSchema.MapValues == nil {
			return &core.MappingNode{Fields: map[string]*core.MappingNode{
				"exampleKey": core.MappingNodeFromString("example-value"),
			}}
		}
		value := buildExampleValue(name, fieldSchema.MapValues, variantComplete, blueprintType)
		if value == nil {
			return nil
		}
		return &core.MappingNode{Fields: map[string]*core.MappingNode{"example": value}}
	case "union":
		for _, branch := range fieldSchema.OneOf {
			if value := buildExampleValue(name, branch, variant, blueprintType); value != nil {
				return value
			}
		}
		return nil
	default:
		return buildScalarValue(name, fieldSchema)
	}
}

func buildScalarValue(name string, fieldSchema *irSchema) *core.MappingNode {
	switch fieldSchema.Type {
	case "integer":
		return core.MappingNodeFromInt(scalarInt(fieldSchema))
	case "float":
		return core.MappingNodeFromFloat(scalarFloat(fieldSchema))
	case "boolean":
		return core.MappingNodeFromBool(false)
	default:
		if len(fieldSchema.Enum) > 0 {
			return core.MappingNodeFromString(fieldSchema.Enum[0])
		}
		return core.MappingNodeFromString(fmt.Sprintf("example-%s", kebab(name)))
	}
}

func scalarInt(fieldSchema *irSchema) int {
	if fieldSchema.Minimum != nil {
		return int(*fieldSchema.Minimum)
	}
	return 1
}

func scalarFloat(fieldSchema *irSchema) float64 {
	if fieldSchema.Minimum != nil {
		return *fieldSchema.Minimum
	}
	return 1
}

func firstSettableScalar(objectSchema *irSchema, blueprintType string) (string, *core.MappingNode) {
	// Prefer the canonical name field, then any curated (seeded) field, then the
	// first settable scalar, so a minimal example shows a meaningful property.
	preferences := []func(attr irAttribute) bool{
		func(attr irAttribute) bool { return strings.HasSuffix(attr.Name, "Name") },
		func(attr irAttribute) bool { return seededExampleValue(blueprintType, attr.Name) != nil },
		func(irAttribute) bool { return true },
	}
	for _, prefer := range preferences {
		for _, attr := range objectSchema.Attributes {
			if attr.Schema.Computed || strings.HasPrefix(attr.Name, "__") || !isScalarType(attr.Schema.Type) {
				continue
			}
			if prefer(attr) {
				return attr.Name, buildExampleValue(attr.Name, attr.Schema, variantComplete, blueprintType)
			}
		}
	}
	return "", nil
}

func resourceExampleName(res *irResource) string {
	parts := strings.Split(res.CFNType, "::")
	if len(parts) < 3 {
		return "example"
	}
	return lowerFirst(parts[2])
}

func literalString(value string) *substitutions.StringOrSubstitutions {
	return &substitutions.StringOrSubstitutions{
		Values: []*substitutions.StringOrSubstitution{
			{StringValue: &value},
		},
	}
}

func singular(name string) string {
	switch {
	case strings.HasSuffix(name, "ies"):
		return name[:len(name)-3] + "y"
	case strings.HasSuffix(name, "s"):
		return name[:len(name)-1]
	default:
		return name
	}
}

func kebab(name string) string {
	var b strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('-')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}
