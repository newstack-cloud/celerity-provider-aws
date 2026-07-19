package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// Every CloudFormation readOnlyProperty must map into the resource's
// ComputedFields: a read-only property missing from the computed set survives
// update-patch diffing and AWS rejects the patch ("readOnlyProperties ...
// cannot be updated"). This pins the generator contract for every vendored
// schema; read-back properties absent from the vendored schema entirely are
// handled by the engine dropping unmanaged patch operations.
func TestVendoredSchemaReadOnlyPropertiesAreComputed(t *testing.T) {
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

		for _, pointer := range schema.ReadOnlyProperties {
			field := topLevelField(pointer)
			if !slices.Contains(resource.ComputedFields, field) {
				t.Errorf(
					"%s: readOnly property %s (top-level field %q) missing from ComputedFields %v",
					file,
					pointer,
					field,
					resource.ComputedFields,
				)
			}
		}
	}
}
