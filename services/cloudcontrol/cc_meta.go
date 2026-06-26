package cloudcontrol

// Internal engine-owned spec fields used to thread Cloud Control state between
// the deploy and stabilisation phases. They are declared on every generated
// schema as Computed + IgnoreDrift so they persist in state (the only way to
// carry them from Deploy to HasStabilised) without ever producing drift noise.
const (
	// Carries the Cloud Control operation RequestToken from
	// Create/Update so HasStabilised can poll GetResourceRequestStatus.
	fieldRequestToken = "__ccRequestToken"
	// Carries Cloud Control's opaque primary identifier
	// (used for Get/Update/Delete), captured at deploy time.
	fieldPrimaryIdentifier = "__ccPrimaryIdentifier"
)

// TagShape describes how a Cloud Control resource type models its tag property,
// since CloudFormation tag shapes are not uniform across types.
type TagShape int

const (
	// TagShapeNone indicates the resource type is not taggable.
	TagShapeNone TagShape = iota
	// TagShapeKeyValueList models tags as a list of {"Key": k, "Value": v} objects.
	TagShapeKeyValueList
	// TagShapeStringMap models tags as a {k: v} string map.
	TagShapeStringMap
)

// CCResourceMeta carries the per-type information the generic engine cannot infer
// from a resource spec at runtime. It is supplied by codegen from the
// CloudFormation resource type schema.
type CCResourceMeta struct {
	// PrimaryIdentifierField is the user-facing spec field that mirrors the CFN
	// primaryIdentifier which is also used as the resource definition's IDField.
	// For composite identifiers it holds the first component.
	PrimaryIdentifierField string
	// PrimaryIdentifierFields holds all components of a composite CFN
	// primaryIdentifier, in order. When set (length > 1), the engine joins the
	// component spec values with "|" to form Cloud Control's composite identifier
	// if no stored identifier is available. Empty for single-identifier types.
	PrimaryIdentifierFields []string
	// ComputedFields lists spec field paths (without the "spec." prefix) that map
	// to CFN readOnlyProperties. They are stripped from desired state and update patches.
	ComputedFields []string
	// CreateOnlyFields lists spec field paths mapping to CFN createOnlyProperties.
	// Changes to these route through MustRecreate, never an update patch.
	CreateOnlyFields []string
	// WriteOnlyFields lists spec field paths mapping to CFN writeOnlyProperties.
	// Cloud Control never returns these, so they are excluded from update diffs
	// and external state reconciliation.
	WriteOnlyFields []string
	// TagPropertyName is the spec/CFN property holding tags, or "" when untaggable.
	TagPropertyName string
	// TagShape describes how TagPropertyName encodes tags.
	TagShape TagShape
	// FieldNameOverrides maps a camelCase field path (with "*" for array/map
	// descent) to its CloudFormation property name. This is for fields where a simple
	// first-character flip does not round-trip, such as
	// "sseSpecification" <-> "SSESpecification".
	// Absent paths fall back to flipping the first character of the last path segment.
	FieldNameOverrides map[string]string
	// JSONStringFields lists camelCase field paths (with "*" for array/map descent)
	// that hold a CloudFormation free-form object encoded as a JSON string. The
	// engine parses the string into a JSON object when building Cloud Control
	// desired state and serialises the object back to a string when reading
	// external state.
	JSONStringFields []string
}

// IsTaggable reports whether the resource type exposes a tag property.
func (m CCResourceMeta) IsTaggable() bool {
	return m.TagPropertyName != "" && m.TagShape != TagShapeNone
}
