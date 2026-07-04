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

// Asserts that a resource whose complete example would be
// identical to its basic one (no optional settable fields beyond those required) emits
// only the basic variant, while a resource with optional fields emits both.
func TestExampleVariantCollapse(t *testing.T) {
	// BucketPolicy has only required fields (bucket, policyDocument), so basic == complete.
	policy := loadResource(t, "aws-s3-bucketpolicy.json", "")
	if !exampleSpecsIdentical(policy) {
		t.Error("expected bucketpolicy basic and complete specs to be identical")
	}
	if got := exampleVariantsFor("", policy); len(got) != 1 || got[0] != variantBasic {
		t.Errorf("expected only the basic variant for bucketpolicy, got %v", variantNames(got))
	}

	// Bucket has many optional fields, so the two variants differ and both are emitted.
	bucket := loadResource(t, "aws-s3-bucket.json", "")
	if exampleSpecsIdentical(bucket) {
		t.Error("expected bucket basic and complete specs to differ")
	}
	if got := exampleVariantsFor("", bucket); len(got) != 2 {
		t.Errorf("expected both variants for bucket, got %v", variantNames(got))
	}
}

func variantNames(variants []exampleVariant) []string {
	names := make([]string, len(variants))
	for i, v := range variants {
		names[i] = v.name()
	}
	return names
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
