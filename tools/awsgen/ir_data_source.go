package main

import (
	"sort"
	"strings"
)

// The generator's intermediate representation of a Cloud
// Control–backed data source, derived from a resource's IR plus its per-type data
// source config.
type irDataSource struct {
	BlueprintType           string
	CFNType                 string
	Label                   string
	Description             string
	FilterFields            []string
	DeriveIdentifierFromARN bool
	ExportFields            []irDataSourceField
	Warnings                []string
}

type irDataSourceField struct {
	Name string
	// Type is one of "string", "integer", "float", "boolean", "array".
	Type string
}

// Builds a data source IR from a resource IR. Export fields are the
// resource's readable fields flattened (objects -> dot-notation, arrays kept whole),
// excluding top-level write-only fields. This matches the engine's runtime flatten.
func deriveDataSource(res *irResource, cfg dataSourceConfig) *irDataSource {
	ds := &irDataSource{
		BlueprintType:           res.BlueprintType,
		CFNType:                 res.CFNType,
		Label:                   res.Label,
		Description:             res.Description,
		FilterFields:            cfg.FilterFields,
		DeriveIdentifierFromARN: cfg.DeriveIdentifierFromARN,
	}

	writeOnly := map[string]bool{}
	for _, field := range res.WriteOnlyFields {
		// The engine only strips top-level write-only fields.
		if !strings.Contains(field, ".") {
			writeOnly[field] = true
		}
	}

	fields := map[string]string{}
	if res.Schema != nil {
		for _, attr := range res.Schema.Attributes {
			if writeOnly[attr.Name] {
				continue
			}
			flattenExportField(attr.Name, attr.Schema, fields)
		}
	}

	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		ds.ExportFields = append(ds.ExportFields, irDataSourceField{Name: name, Type: fields[name]})
	}
	return ds
}

// Flattens one schema field into export entries: objects recurse
// into dot-notation, arrays are kept whole, and scalars/unions map to a leaf type.
func flattenExportField(prefix string, schema *irSchema, out map[string]string) {
	if schema == nil {
		return
	}
	switch schema.Type {
	case "object":
		for _, attr := range schema.Attributes {
			flattenExportField(prefix+"."+attr.Name, attr.Schema, out)
		}
	case "array":
		out[prefix] = "array"
	case "map":
		// The export contract has no map type, so maps are exported as a JSON string
		// (the engine serialises the map value to JSON at fetch time).
		out[prefix] = "string"
	case "union":
		out[prefix] = unionExportType(schema)
	case "string", "integer", "float", "boolean":
		out[prefix] = schema.Type
	default:
		out[prefix] = "string"
	}
}

// Picks an export type for a union: "array" when any branch is an
// array (e.g. DynamoDB keySchema is array|object), otherwise "string".
func unionExportType(schema *irSchema) string {
	for _, branch := range schema.OneOf {
		if branch != nil && branch.Type == "array" {
			return "array"
		}
	}
	return "string"
}
