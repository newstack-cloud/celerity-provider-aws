package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const patternSchema = `{
  "typeName": "AWS::Example::Patterned",
  "description": "A synthetic type exercising pattern sanitisation.",
  "properties": {
    "Id": {"type": "string"},
    "BucketName": {
      "type": "string",
      "pattern": "^[0-9A-Za-z\\.\\-_]*(?<!\\.)$",
      "minLength": 3,
      "maxLength": 63
    },
    "Name": {"type": "string", "pattern": "^[a-z]+$"}
  },
  "required": ["Name"],
  "primaryIdentifier": ["/properties/Id"],
  "readOnlyProperties": ["/properties/Id"]
}`

func convertPatterned(t *testing.T) *irResource {
	t.Helper()
	schema, err := loadCFNSchema([]byte(patternSchema))
	if err != nil {
		t.Fatal(err)
	}
	resource, err := convert(schema, "aws/example/patterned")
	if err != nil {
		t.Fatal(err)
	}
	return resource
}

func TestConvertDropsRE2IncompatiblePatterns(t *testing.T) {
	resource := convertPatterned(t)

	bucketName := mustFindAttribute(t, resource.Schema, "bucketName")
	if bucketName.Pattern != "" {
		t.Errorf("lookbehind pattern should be dropped, got %q", bucketName.Pattern)
	}
	if bucketName.DroppedPattern != `^[0-9A-Za-z\.\-_]*(?<!\.)$` {
		t.Errorf("dropped pattern should be recorded, got %q", bucketName.DroppedPattern)
	}
	if bucketName.MinLength == nil || *bucketName.MinLength != 3 {
		t.Error("length constraints should survive a dropped pattern")
	}

	name := mustFindAttribute(t, resource.Schema, "name")
	if name.Pattern != "^[a-z]+$" {
		t.Errorf("RE2-compatible pattern should be kept, got %q", name.Pattern)
	}

	warned := false
	for _, warning := range resource.Warnings {
		if strings.Contains(warning, "BucketName") && strings.Contains(warning, "dropped") {
			warned = true
		}
	}
	if !warned {
		t.Errorf("expected a dropped-pattern warning for BucketName, got %v", resource.Warnings)
	}
}

func TestEmitAnnotatesDroppedPatterns(t *testing.T) {
	resource := convertPatterned(t)
	source, err := emitResourceFile(resource)
	if err != nil {
		t.Fatal(err)
	}

	emitted := string(source)
	if !strings.Contains(emitted, "dropped: not compilable by Go's RE2 regexp") {
		t.Error("emitted source should carry a comment for the dropped pattern")
	}
	if strings.Contains(emitted, `Pattern: "^[0-9A-Za-z`) {
		t.Error("dropped pattern should not be emitted as a Pattern field")
	}
	if !strings.Contains(emitted, `Pattern: "^[a-z]+$"`) {
		t.Error("RE2-compatible pattern should be emitted")
	}
}

// Pins the whole class of ECMA-only patterns: every pattern the generator would
// emit from the vendored CloudFormation schemas must compile under Go's RE2.
func TestVendoredSchemaPatternsCompile(t *testing.T) {
	files, err := vendoredSchemaFiles("schemas")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no vendored schemas found")
	}

	for _, file := range files {
		data, err := os.ReadFile(filepath.Join("schemas", file))
		if err != nil {
			t.Fatal(err)
		}
		schema, err := loadCFNSchema(data)
		if err != nil {
			t.Fatalf("loading %s: %v", file, err)
		}
		resource, err := convert(schema, blueprintTypeFor(schema.TypeName))
		if err != nil {
			t.Fatalf("converting %s: %v", file, err)
		}

		walkIRSchemas(resource.Schema, func(node *irSchema) {
			if node.Pattern == "" {
				return
			}
			if _, err := regexp.Compile(node.Pattern); err != nil {
				t.Errorf("%s: pattern %q does not compile: %v", file, node.Pattern, err)
			}
		})
	}
}

func walkIRSchemas(schema *irSchema, visit func(*irSchema)) {
	if schema == nil {
		return
	}
	visit(schema)
	for _, attr := range schema.Attributes {
		walkIRSchemas(attr.Schema, visit)
	}
	walkIRSchemas(schema.Items, visit)
	walkIRSchemas(schema.MapValues, visit)
	for _, branch := range schema.OneOf {
		walkIRSchemas(branch, visit)
	}
}

// Returns the named attribute or fails the test.
//
// Callers previously nil-checked and then dereferenced, which staticcheck reads as a
// possible nil dereference (SA5011) because it cannot see that the check terminates.
// Returning a value the caller can use without checking removes both the repetition and
// the warning.
func mustFindAttribute(t *testing.T, schema *irSchema, name string) *irSchema {
	t.Helper()

	attr := findAttribute(schema, name)
	if attr == nil {
		t.Fatalf("%s attribute not found", name)
	}

	return attr
}
