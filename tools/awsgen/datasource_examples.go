package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Writes the example markdown for a data source to
// examples/datasources/<stem>.md. A curated override (Layer 2) is used as-is when
// present; otherwise, a representative example is generated from the data source's
// filter and export fields.
func generateDataSourceExample(outDir, curatedDir string, ds *irDataSource) error {
	dir := filepath.Join(outDir, "examples", "datasources")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	fileName := exampleStem(ds.CFNType) + ".md"

	content, ok := curatedExample(curatedDir, fileName)
	if !ok {
		generated, err := renderDataSourceExample(ds)
		if err != nil {
			return err
		}
		content = generated
	}
	return os.WriteFile(filepath.Join(dir, fileName), []byte(content), 0o644)
}

func curatedExample(curatedDir, fileName string) (string, bool) {
	if curatedDir == "" {
		return "", false
	}
	data, err := os.ReadFile(filepath.Join(curatedDir, fileName))
	if err != nil {
		return "", false
	}
	return string(data), true
}

const dataSourceExampleVersion = "2025-11-02"

// Builds a representative data source example (blueprintlang,
// yaml and jsonc) from the data source's filter and export fields.
func renderDataSourceExample(ds *irDataSource) (string, error) {
	name := exampleDataSourceName(ds.BlueprintType)
	filterField := representativeFilterField(ds.FilterFields)
	filterValue := exampleFilterValue(filterField)
	exports := representativeExports(ds.ExportFields)
	if len(exports) == 0 {
		// Should not happen for the onboarded types (all expose `arn`), but keep the
		// example valid by exporting the filter field.
		exports = []irDataSourceField{{Name: filterField, Type: "string"}}
	}
	primary := exports[0]
	exportRefName := name + upperFirst(strings.ReplaceAll(primary.Name, ".", ""))

	var b strings.Builder
	fmt.Fprintf(&b, "Look up an existing %s by %s and export its %s.\n\n", ds.Label, filterField, primary.Name)

	// blueprintlang
	b.WriteString("```blueprintlang\n")
	fmt.Fprintf(&b, "version %q\n\n", dataSourceExampleVersion)
	fmt.Fprintf(&b, "data %s: %s {\n", name, ds.BlueprintType)
	fmt.Fprintf(&b, "    filter %q == %q\n\n", filterField, filterValue)
	for _, export := range exports {
		fmt.Fprintf(&b, "    export %s: %s\n", export.Name, export.Type)
	}
	b.WriteString("}\n\n")
	fmt.Fprintf(&b, "export %s: %s {\n", exportRefName, primary.Type)
	fmt.Fprintf(&b, "    field = datasources.%s.%s\n", name, primary.Name)
	b.WriteString("}\n```\n\n")

	// yaml
	b.WriteString("```yaml\n")
	fmt.Fprintf(&b, "version: %s\n\n", dataSourceExampleVersion)
	b.WriteString("datasources:\n")
	fmt.Fprintf(&b, "  %s:\n", name)
	fmt.Fprintf(&b, "    type: %s\n", ds.BlueprintType)
	b.WriteString("    filter:\n")
	fmt.Fprintf(&b, "      field: %s\n", filterField)
	b.WriteString("      operator: \"=\"\n")
	fmt.Fprintf(&b, "      search: %s\n", filterValue)
	b.WriteString("    exports:\n")
	for _, export := range exports {
		fmt.Fprintf(&b, "      %s:\n        type: %s\n", export.Name, export.Type)
	}
	b.WriteString("\nexports:\n")
	fmt.Fprintf(&b, "  %s:\n    type: %s\n    field: datasources.%s.%s\n", exportRefName, primary.Type, name, primary.Name)
	b.WriteString("```\n\n")

	// jsonc
	b.WriteString("```javascript\n{\n")
	fmt.Fprintf(&b, "  \"version\": %q,\n", dataSourceExampleVersion)
	b.WriteString("  \"datasources\": {\n")
	fmt.Fprintf(&b, "    %q: {\n", name)
	fmt.Fprintf(&b, "      \"type\": %q,\n", ds.BlueprintType)
	fmt.Fprintf(&b, "      \"filter\": { \"field\": %q, \"operator\": \"=\", \"search\": %q },\n", filterField, filterValue)
	b.WriteString("      \"exports\": {\n")
	for i, export := range exports {
		comma := ","
		if i == len(exports)-1 {
			comma = ""
		}
		fmt.Fprintf(&b, "        %q: { \"type\": %q }%s\n", export.Name, export.Type, comma)
	}
	b.WriteString("      }\n    }\n  }\n}\n```\n")

	return b.String(), nil
}

func exampleDataSourceName(blueprintType string) string {
	parts := strings.Split(blueprintType, "/")
	return "example" + upperFirst(parts[len(parts)-1])
}

// Prefer a name-style filter over arn/region for the example.
func representativeFilterField(filterFields []string) string {
	for _, field := range filterFields {
		if field != "arn" && field != "region" {
			return field
		}
	}
	if len(filterFields) > 0 {
		return filterFields[0]
	}
	return "name"
}

// Picks up to three top-level scalar export fields for the
// example, preferring `arn` first (dotted/array fields are skipped to keep the example
// simple and valid).
func representativeExports(fields []irDataSourceField) []irDataSourceField {
	var arn *irDataSourceField
	var scalars []irDataSourceField
	for i := range fields {
		field := fields[i]
		if strings.Contains(field.Name, ".") || field.Type == "array" {
			continue
		}
		if field.Name == "arn" {
			arn = &fields[i]
			continue
		}
		scalars = append(scalars, field)
	}

	var chosen []irDataSourceField
	if arn != nil {
		chosen = append(chosen, *arn)
	}

	for _, field := range scalars {
		if len(chosen) >= 3 {
			break
		}
		chosen = append(chosen, field)
	}
	return chosen
}

func exampleFilterValue(field string) string {
	switch {
	case field == "arn":
		return "arn:aws:service:us-west-2:123456789012:resource/example"
	case strings.Contains(strings.ToLower(field), "url"):
		return "https://sqs.us-west-2.amazonaws.com/123456789012/example"
	default:
		return "example-" + strings.ToLower(field)
	}
}
