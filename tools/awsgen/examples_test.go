package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/lang"
)

// TestExampleGeneration builds basic and complete examples for each pilot schema
// and asserts they render all three formats and that the blueprintlang round-trips
// (renderExampleMarkdown returns an error if the emitted blueprintlang fails to
// re-parse).
func TestExampleGeneration(t *testing.T) {
	files, err := vendoredSchemaFiles("schemas")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			resource := loadResource(t, file, "")

			basic := renderVariant(t, resource, variantBasic)
			complete := renderVariant(t, resource, variantComplete)

			for _, language := range []string{"```blueprintlang", "```yaml", "```javascript"} {
				if !strings.Contains(basic, language) {
					t.Errorf("basic example missing %s block", language)
				}
			}
			if basic == complete {
				t.Error("basic and complete examples should differ")
			}
		})
	}
}

func renderVariant(t *testing.T, resource *irResource, variant exampleVariant) string {
	t.Helper()
	blueprint := buildExampleBlueprint(resource, variant)
	content, err := renderExampleMarkdown(blueprint, variant.description(resource.Label))
	if err != nil {
		t.Fatalf("rendering %s example: %v", variant.name(), err)
	}

	// Independently confirm the blueprintlang block parses (the round-trip gate).
	bp, err := lang.Emit(blueprint)
	if err != nil {
		t.Fatalf("emitting blueprintlang: %v", err)
	}
	if _, err := lang.ParseString(bp); err != nil {
		t.Fatalf("emitted blueprintlang does not parse: %v", err)
	}
	return content
}

func loadResource(t *testing.T, schemaFile, blueprintType string) *irResource {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("schemas", schemaFile))
	if err != nil {
		t.Fatal(err)
	}
	schema, err := loadCFNSchema(data)
	if err != nil {
		t.Fatal(err)
	}
	if blueprintType == "" {
		blueprintType = blueprintTypeFor(schema.TypeName)
	}
	resource, err := convert(schema, blueprintType)
	if err != nil {
		t.Fatal(err)
	}
	return resource
}
