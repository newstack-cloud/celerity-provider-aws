package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// The generator's intermediate representation of a resource type,
// decoupled from both the CloudFormation schema and the emitted Go.
type irResource struct {
	BlueprintType           string
	CFNType                 string
	Label                   string
	Description             string
	Schema                  *irSchema
	PrimaryIdentifierField  string
	PrimaryIdentifierFields []string
	ComputedFields          []string
	CreateOnlyFields        []string
	WriteOnlyFields         []string
	TagPropertyName         string
	TagShape                string
	// FieldNameOverrides maps a camelCase field path (with "*" for array/map
	// descent) to its CloudFormation property name, for the fields where a simple
	// first-character flip does not round-trip such as
	// "sseSpecification" <-> "SSESpecification".
	FieldNameOverrides map[string]string
	// JSONStringFields lists camelCase field paths (with "*" for array/map descent)
	// that hold a CloudFormation free-form object encoded as a JSON string.
	JSONStringFields []string
	Warnings         []string
}

type irSchema struct {
	Type         string
	Label        string
	Description  string
	Attributes   []irAttribute
	Items        *irSchema
	MapValues    *irSchema
	OneOf        []*irSchema
	Required     []string
	Nullable     bool
	Computed     bool
	MustRecreate bool
	Sensitive    bool
	IgnoreDrift  bool
	// JSONString marks a string field that encodes a CloudFormation free-form
	// object (which the repo cannot model as an open object). The engine parses it
	// back to a JSON object when sending desired state to Cloud Control and
	// serialises it to a string when reading external state.
	JSONString bool
	Enum       []string
	Pattern    string
	// DroppedPattern records a CloudFormation `pattern` that Go's RE2 regexp
	// cannot compile (CloudFormation patterns are ECMA-flavoured and may use
	// lookarounds). These are omitted from the emitted schema where it would
	// hard-fail spec validation, and are surfaced as comments instead.
	DroppedPattern string
	Minimum        *float64
	Maximum        *float64
	MinLength      *int
	MaxLength      *int
}

type irAttribute struct {
	Name    string
	CFNName string
	Schema  *irSchema
}

// Walks a single CloudFormation schema, resolving $ref against its
// definitions while guarding against reference cycles.
type converter struct {
	schema    *cfnSchema
	resolving map[string]bool
	warnings  []string
}

func convert(schema *cfnSchema, blueprintType string) (*irResource, error) {
	c := &converter{schema: schema, resolving: map[string]bool{}}

	root := c.convertObject(schema.Properties, schema.Required, labelFromTypeName(schema.TypeName))
	root.Description = schema.Description

	sanitiseSchemaDescriptions(root)
	resourceDescription := root.Description
	if resourceDescription == "" {
		resourceDescription = fallbackDescription(schema.TypeName)
	}
	root.Description = resourceDescription

	resource := &irResource{
		BlueprintType: blueprintType,
		CFNType:       schema.TypeName,
		Label:         labelFromTypeName(schema.TypeName),
		Description:   resourceDescription,
		Schema:        root,
	}

	c.stampPointerLists(resource, root)
	c.applyTagging(resource, root)

	overrides := map[string]string{}
	collectNameOverrides(root, "", overrides)
	if len(overrides) > 0 {
		resource.FieldNameOverrides = overrides
	}

	var jsonStringFields []string
	collectJSONStringFields(root, "", &jsonStringFields)
	resource.JSONStringFields = jsonStringFields

	resource.Warnings = c.warnings

	return resource, nil
}

// Collects every object attribute whose CloudFormation
// name is not recoverable by a first-character flip of the camelCase field name,
// the camelCase path -> CFN name mapping the engine needs to translate exactly.
// Union branches share their parent's path because the engine resolves a union to
// a branch without adding a path segment.
func collectNameOverrides(schema *irSchema, path string, out map[string]string) {
	if schema == nil {
		return
	}
	switch schema.Type {
	case "object":
		for _, attr := range schema.Attributes {
			childPath := joinPath(path, attr.Name)
			if attr.CFNName != "" && upperFirst(attr.Name) != attr.CFNName {
				out[childPath] = attr.CFNName
			}
			collectNameOverrides(attr.Schema, childPath, out)
		}
	case "array":
		collectNameOverrides(schema.Items, joinPath(path, "*"), out)
	case "map":
		collectNameOverrides(schema.MapValues, joinPath(path, "*"), out)
	case "union":
		for _, branch := range schema.OneOf {
			collectNameOverrides(branch, path, out)
		}
	}
}

// Records the camelCase paths of every free-form object
// field encoded as a JSON string, so the engine can parse/serialise them at the
// Cloud Control boundary. Union branches share their parent's path.
func collectJSONStringFields(schema *irSchema, path string, out *[]string) {
	if schema == nil {
		return
	}
	if schema.JSONString {
		*out = append(*out, path)
		return
	}
	switch schema.Type {
	case "object":
		for _, attr := range schema.Attributes {
			collectJSONStringFields(attr.Schema, joinPath(path, attr.Name), out)
		}
	case "array":
		collectJSONStringFields(schema.Items, joinPath(path, "*"), out)
	case "map":
		collectJSONStringFields(schema.MapValues, joinPath(path, "*"), out)
	case "union":
		for _, branch := range schema.OneOf {
			collectJSONStringFields(branch, path, out)
		}
	}
}

// Rewrites every description in the schema tree to
// strip CloudFormation identifiers, links and framing carried over from the
// registry schema.
func sanitiseSchemaDescriptions(schema *irSchema) {
	if schema == nil {
		return
	}
	schema.Description = sanitiseDescription(schema.Description)
	for _, attr := range schema.Attributes {
		sanitiseSchemaDescriptions(attr.Schema)
	}
	sanitiseSchemaDescriptions(schema.Items)
	sanitiseSchemaDescriptions(schema.MapValues)
	for _, branch := range schema.OneOf {
		sanitiseSchemaDescriptions(branch)
	}
}

func joinPath(path, segment string) string {
	if path == "" {
		return segment
	}
	return path + "." + segment
}

func (c *converter) stampPointerLists(resource *irResource, root *irSchema) {
	for _, pointer := range c.schema.ReadOnlyProperties {
		stampPointer(root, pointer, func(s *irSchema) { s.Computed = true })
		resource.ComputedFields = appendTopLevelField(resource.ComputedFields, pointer)
	}
	for _, pointer := range c.schema.CreateOnlyProperties {
		stampPointer(root, pointer, func(s *irSchema) { s.MustRecreate = true })
		resource.CreateOnlyFields = appendTopLevelField(resource.CreateOnlyFields, pointer)
	}
	for _, pointer := range c.schema.WriteOnlyProperties {
		stampPointer(root, pointer, func(s *irSchema) {
			s.Sensitive = true
			s.IgnoreDrift = true
		})
		resource.WriteOnlyFields = appendTopLevelField(resource.WriteOnlyFields, pointer)
	}

	for _, pointer := range c.schema.PrimaryIdentifier {
		resource.PrimaryIdentifierFields = append(
			resource.PrimaryIdentifierFields,
			topLevelField(pointer),
		)
	}
	switch len(resource.PrimaryIdentifierFields) {
	case 0:
		c.warnings = append(c.warnings, "no primaryIdentifier defined")
	default:
		// The first component is the resource definition's IDField. The engine
		// joins the full ordered set with "|" to form Cloud Control's composite
		// identifier when no stored identifier is available.
		resource.PrimaryIdentifierField = resource.PrimaryIdentifierFields[0]
	}
}

func (c *converter) applyTagging(resource *irResource, root *irSchema) {
	if c.schema.Tagging == nil || !c.schema.Tagging.Taggable || c.schema.Tagging.TagProperty == "" {
		return
	}
	field := topLevelField(c.schema.Tagging.TagProperty)
	resource.TagPropertyName = field
	resource.TagShape = detectTagShape(findAttribute(root, field))
}

func detectTagShape(tags *irSchema) string {
	if tags == nil {
		return "cloudcontrol.TagShapeKeyValueList"
	}
	switch tags.Type {
	case "map":
		return "cloudcontrol.TagShapeStringMap"
	default:
		return "cloudcontrol.TagShapeKeyValueList"
	}
}

// Builds an object schema from a CloudFormation property map.
func (c *converter) convertObject(
	properties map[string]*cfnProperty,
	required []string,
	label string,
) *irSchema {
	requiredSet := toSet(required)
	schema := &irSchema{Type: "object", Label: label}

	names := sortedKeys(properties)
	for _, name := range names {
		child := c.convertProperty(properties[name], name)
		child.Nullable = !requiredSet[name]
		fieldName := camelFromCFN(name)
		schema.Attributes = append(schema.Attributes, irAttribute{
			Name:    fieldName,
			CFNName: name,
			Schema:  child,
		})
		if requiredSet[name] {
			schema.Required = append(schema.Required, fieldName)
		}
	}
	return schema
}

func (c *converter) convertProperty(prop *cfnProperty, name string) *irSchema {
	if prop == nil {
		return &irSchema{Type: "string"}
	}

	if prop.Ref != "" {
		return c.convertRef(prop.Ref, name)
	}

	if branches := firstNonEmpty(prop.OneOf, prop.AnyOf); branches != nil {
		return c.convertUnion(branches, prop, name)
	}
	if len(prop.AllOf) > 0 {
		return c.convertAllOf(prop.AllOf, name)
	}

	schema := c.convertByType(prop, name)
	c.applyConstraints(schema, prop, name)
	if schema.Description == "" {
		schema.Description = prop.Description
	}
	return schema
}

func (c *converter) convertByType(prop *cfnProperty, name string) *irSchema {
	switch resolveType(prop) {
	case "object":
		return c.convertObjectProperty(prop, name)
	case "array":
		items := c.convertProperty(prop.Items, name)
		return &irSchema{Type: "array", Label: labelFromName(name), Items: items}
	case "integer":
		return &irSchema{Type: "integer"}
	case "number":
		return &irSchema{Type: "float"}
	case "boolean":
		return &irSchema{Type: "boolean"}
	case "multi":
		return c.convertMultiType(prop, name)
	default:
		return &irSchema{Type: "string"}
	}
}

func (c *converter) convertObjectProperty(prop *cfnProperty, name string) *irSchema {
	if len(prop.Properties) > 0 {
		return c.convertObject(prop.Properties, prop.Required, labelFromName(name))
	}
	if values := c.mapValueSchema(prop, name); values != nil {
		return &irSchema{Type: "map", Label: labelFromName(name), MapValues: values}
	}
	// Free-form object: modelled as a free-form map (string keys, arbitrary values) so
	// it can be authored as a native blueprint object. The engine passes the value
	// through as is to Cloud Control (and back) with no JSON-string round-trip.
	c.warnings = append(c.warnings, fmt.Sprintf(
		"property %q is a free-form object; modelled as an open map", name,
	))
	return &irSchema{Type: "map", Label: labelFromName(name)}
}

// The value schema when an object models a typed map via
// patternProperties or a schema-valued additionalProperties.
func (c *converter) mapValueSchema(prop *cfnProperty, name string) *irSchema {
	for _, valueProp := range prop.PatternProperties {
		return c.convertProperty(valueProp, name)
	}
	if len(prop.AdditionalProperties) > 0 && prop.AdditionalProperties[0] == '{' {
		var valueProp cfnProperty
		if err := json.Unmarshal(prop.AdditionalProperties, &valueProp); err == nil {
			return c.convertProperty(&valueProp, name)
		}
	}
	return nil
}

func (c *converter) convertMultiType(prop *cfnProperty, name string) *irSchema {
	var branches []*irSchema
	seen := map[string]bool{}
	for _, t := range prop.Type {
		branch := c.convertByType(&cfnProperty{
			Type:       stringList{t},
			Properties: prop.Properties,
			Items:      prop.Items,
		}, name)
		key := branchKey(branch)
		if seen[key] {
			continue
		}
		seen[key] = true
		branches = append(branches, branch)
	}
	branches = reduceFreeFormUnion(branches)
	if len(branches) == 1 {
		return branches[0]
	}
	return &irSchema{Type: "union", Label: labelFromName(name), OneOf: branches}
}

// Simplifies a union that contains a free-form map branch
// (produced from a bare `{"type": "object"}`). AWS CFN schemas frequently model a
// value as a `oneOf`/multi-type that combines a structured form with a loose object
// or a stringified-JSON form. The structured form is authoritative:
//   - if a more specific typed branch (array, typed object, scalar, typed map) is
//     present, the open-map and any plain-string branches are dropped, e.g.
//     DynamoDB KeySchema `oneOf: [array, {"type":"object"}]` -> a typed array;
//   - otherwise the union collapses to the single open map, dropping the redundant
//     stringified-JSON string branch, e.g. an IAM policy document or EventBridge
//     event pattern `["object", "string"]` -> an open map authored as a native
//     object.
//
// Unions without an open-map branch are returned unchanged.
func reduceFreeFormUnion(branches []*irSchema) []*irSchema {
	var openMaps, specific []*irSchema
	for _, branch := range branches {
		switch {
		case isOpenMap(branch):
			openMaps = append(openMaps, branch)
		case isSpecificTyped(branch):
			specific = append(specific, branch)
		}
	}
	if len(openMaps) == 0 {
		return branches
	}
	if len(specific) > 0 {
		return specific
	}
	return openMaps[:1]
}

func isOpenMap(s *irSchema) bool {
	return s != nil && s.Type == "map" && s.MapValues == nil
}

func isSpecificTyped(s *irSchema) bool {
	if s == nil {
		return false
	}
	switch s.Type {
	case "array", "object", "integer", "float", "boolean", "union":
		return true
	case "map":
		return s.MapValues != nil
	case "string":
		return len(s.Enum) > 0
	}
	return false
}

func (c *converter) convertUnion(branches []*cfnProperty, prop *cfnProperty, name string) *irSchema {
	var converted []*irSchema
	seen := map[string]bool{}
	for _, branch := range branches {
		schema := c.convertProperty(branch, name)
		key := branchKey(schema)
		if seen[key] {
			continue
		}
		seen[key] = true
		converted = append(converted, schema)
	}
	converted = reduceFreeFormUnion(converted)
	if len(converted) == 1 {
		return converted[0]
	}
	result := &irSchema{Type: "union", Label: labelFromName(name), OneOf: converted}
	if result.Description == "" {
		result.Description = prop.Description
	}
	return result
}

// Merges object branches into a single object, non-object branches
// fall back to the first branch.
func (c *converter) convertAllOf(branches []*cfnProperty, name string) *irSchema {
	merged := &irSchema{Type: "object", Label: labelFromName(name)}
	seen := map[string]bool{}
	for _, branch := range branches {
		schema := c.convertProperty(branch, name)
		if schema.Type != "object" {
			return schema
		}
		for _, attr := range schema.Attributes {
			if seen[attr.Name] {
				continue
			}
			seen[attr.Name] = true
			merged.Attributes = append(merged.Attributes, attr)
		}
		merged.Required = append(merged.Required, schema.Required...)
	}

	return merged
}

func (c *converter) convertRef(ref, name string) *irSchema {
	defName := strings.TrimPrefix(ref, "#/definitions/")
	if c.resolving[defName] {
		// Cycle: model the recursive reference as a free-form map (arbitrary nested
		// object) the engine passes through as-is to Cloud Control.
		c.warnings = append(c.warnings, fmt.Sprintf(
			"recursive definition %q modelled as an open map", defName,
		))
		return &irSchema{Type: "map", Label: labelFromName(name)}
	}
	def, ok := c.schema.Definitions[defName]
	if !ok {
		c.warnings = append(c.warnings, fmt.Sprintf("unknown definition %q", defName))
		return &irSchema{Type: "string"}
	}

	c.resolving[defName] = true
	defer delete(c.resolving, defName)

	schema := c.convertProperty(def, name)
	if schema.Type == "object" && schema.Label == "" {
		schema.Label = labelFromName(defName)
	}
	return schema
}

func (c *converter) applyConstraints(schema *irSchema, prop *cfnProperty, name string) {
	if !isScalarType(schema.Type) {
		return
	}
	schema.Pattern = prop.Pattern
	// CloudFormation patterns are ECMA-flavoured; a pattern Go's RE2 regexp
	// cannot compile (e.g. lookarounds) would hard-fail blueprint spec
	// validation, so it is dropped. Length/enum/type constraints still apply
	// and AWS remains the authoritative validator at deploy time.
	if _, err := regexp.Compile(schema.Pattern); err != nil {
		c.warnings = append(c.warnings, fmt.Sprintf(
			"property %q pattern %q is not RE2-compatible; dropped", name, schema.Pattern,
		))
		schema.DroppedPattern = schema.Pattern
		schema.Pattern = ""
	}
	schema.Minimum = prop.Minimum
	schema.Maximum = prop.Maximum
	schema.MinLength = prop.MinLength
	schema.MaxLength = prop.MaxLength
	schema.Enum = stringEnum(prop.Enum)
}

func resolveType(prop *cfnProperty) string {
	switch len(prop.Type) {
	case 0:
		if len(prop.Properties) > 0 {
			return "object"
		}
		return "string"
	case 1:
		return prop.Type[0]
	default:
		return "multi"
	}
}

func stampPointer(root *irSchema, pointer string, apply func(*irSchema)) {
	segments := pointerSegments(pointer)
	stampSegments(root, segments, apply)
}

func stampSegments(schema *irSchema, segments []string, apply func(*irSchema)) {
	if schema == nil {
		return
	}
	if len(segments) == 0 {
		apply(schema)
		return
	}
	segment := segments[0]
	rest := segments[1:]

	if segment == "*" {
		stampSegments(schema.Items, rest, apply)
		stampSegments(schema.MapValues, rest, apply)
		return
	}
	if child := findAttribute(schema, camelFromCFN(segment)); child != nil {
		stampSegments(child, rest, apply)
	}
}

func findAttribute(schema *irSchema, name string) *irSchema {
	if schema == nil {
		return nil
	}
	for _, attr := range schema.Attributes {
		if attr.Name == name {
			return attr.Schema
		}
	}
	return nil
}

func pointerSegments(pointer string) []string {
	trimmed := strings.TrimPrefix(pointer, "/properties/")
	return strings.Split(trimmed, "/")
}

func topLevelField(pointer string) string {
	return camelFromCFN(pointerSegments(pointer)[0])
}

func appendTopLevelField(fields []string, pointer string) []string {
	field := topLevelField(pointer)
	if slices.Contains(fields, field) {
		return fields
	}
	return append(fields, field)
}

func stringEnum(values []json.RawMessage) []string {
	result := make([]string, 0, len(values))
	for _, raw := range values {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil
		}
		result = append(result, s)
	}
	return result
}

func branchKey(schema *irSchema) string {
	return schema.Type
}

func firstNonEmpty(a, b []*cfnProperty) []*cfnProperty {
	if len(a) > 0 {
		return a
	}
	return b
}

func isScalarType(t string) bool {
	switch t {
	case "string", "integer", "float", "boolean":
		return true
	default:
		return false
	}
}

func toSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func sortedKeys(properties map[string]*cfnProperty) []string {
	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[size:]
}

// Converts a CloudFormation PascalCase property name to an idiomatic
// camelCase field name, lowercasing a leading acronym run as a unit:
// "QueueName" -> "queueName", "KMSMasterKeyId" -> "kmsMasterKeyId",
// "SSESpecification" -> "sseSpecification", "ARN" -> "arn".
func camelFromCFN(name string) string {
	runes := []rune(name)
	if len(runes) == 0 {
		return name
	}
	upper := 0
	for upper < len(runes) && unicode.IsUpper(runes[upper]) {
		upper++
	}
	switch {
	case upper <= 1:
		runes[0] = unicode.ToLower(runes[0])
	case upper == len(runes):
		// The whole identifier is an acronym, e.g. "ARN" -> "arn".
		for i := range runes {
			runes[i] = unicode.ToLower(runes[i])
		}
	default:
		// A leading acronym followed by a word: the final uppercase letter starts
		// the next word, so lowercase every letter of the run but that one,
		// e.g. "KMSMasterKeyId" -> "kmsMasterKeyId".
		for i := 0; i < upper-1; i++ {
			runes[i] = unicode.ToLower(runes[i])
		}
	}
	return string(runes)
}

// Turns a PascalCase or camelCase identifier into a spaced label,
// e.g. "QueueName" -> "Queue Name".
func labelFromName(name string) string {
	var b strings.Builder
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) && !unicode.IsUpper(rune(name[i-1])) {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return upperFirst(strings.TrimSpace(b.String()))
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func labelFromTypeName(typeName string) string {
	parts := strings.Split(typeName, "::")
	return "AWS " + strings.Join(parts[1:], " ")
}
