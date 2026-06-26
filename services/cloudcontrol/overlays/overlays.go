// Package overlays holds hand-written, per-type adjustments applied to generated
// Cloud Control resource schemas. Some CloudFormation constructs (free-form
// objects, gnarly $ref graphs, constraints expressed only in prose) do not
// translate cleanly, so an overlay refines the generated base for those types.
package overlays

import "github.com/newstack-cloud/bluelink/libs/blueprint/provider"

// SchemaOverlay refines a generated schema in place and returns it.
type SchemaOverlay func(*provider.ResourceDefinitionsSchema) *provider.ResourceDefinitionsSchema

var registry = map[string]SchemaOverlay{}

// Register associates an overlay with a Bluelink resource type. It is intended to
// be called from per-type init functions and panics on a duplicate registration.
func Register(blueprintType string, overlay SchemaOverlay) {
	if _, exists := registry[blueprintType]; exists {
		panic("duplicate schema overlay registered for " + blueprintType)
	}
	registry[blueprintType] = overlay
}

// Apply runs the registered overlay for a type, if any, returning the (possibly
// refined) schema. Generated code calls this around its base schema builder.
func Apply(
	blueprintType string,
	schema *provider.ResourceDefinitionsSchema,
) *provider.ResourceDefinitionsSchema {
	if overlay, ok := registry[blueprintType]; ok {
		return overlay(schema)
	}
	return schema
}

// MetaAdjustment describes changes an overlay needs applied to a type's
// CCResourceMeta that a schema overlay alone cannot express, since the Meta lives
// in the cloudcontrol package and overlays must not import it. It is expressed in
// primitive terms to avoid an import cycle.
type MetaAdjustment struct {
	// RemoveJSONStringFields lists field paths the overlay has re-typed as
	// structured objects, so the engine must no longer treat them as JSON strings
	// (parse/serialise) at the Cloud Control boundary.
	RemoveJSONStringFields []string
	// AddFieldNameOverrides maps camelCase field paths the overlay introduced to
	// their CloudFormation property names, for keys a first-character flip cannot
	// recover (e.g. "...principal.aws" -> "AWS").
	AddFieldNameOverrides map[string]string
}

var metaAdjustments = map[string]*MetaAdjustment{}

// RegisterMetaAdjustment associates a meta adjustment with a Bluelink resource
// type. It is intended to be called from per-type init functions and panics on a
// duplicate registration.
func RegisterMetaAdjustment(blueprintType string, adjustment *MetaAdjustment) {
	if _, exists := metaAdjustments[blueprintType]; exists {
		panic("duplicate meta adjustment registered for " + blueprintType)
	}
	metaAdjustments[blueprintType] = adjustment
}

// MetaAdjustmentFor returns the registered meta adjustment for a type, or nil.
func MetaAdjustmentFor(blueprintType string) *MetaAdjustment {
	return metaAdjustments[blueprintType]
}
