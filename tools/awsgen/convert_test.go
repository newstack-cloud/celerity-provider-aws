package main

import (
	"slices"
	"testing"
)

const syntheticSchema = `{
  "typeName": "AWS::Example::Widget",
  "description": "A synthetic type covering the mapping table.",
  "definitions": {
    "Endpoint": {
      "type": "object",
      "properties": {
        "Host": {"type": "string"},
        "Port": {"type": "integer", "minimum": 1, "maximum": 65535}
      },
      "required": ["Host"]
    },
    "Tag": {
      "type": "object",
      "properties": {
        "Key": {"type": "string"},
        "Value": {"type": "string"}
      }
    }
  },
  "properties": {
    "Id": {"type": "string"},
    "Name": {"type": "string"},
    "Region": {"type": "string"},
    "Secret": {"type": "string"},
    "Mode": {"type": "string", "enum": ["FAST", "SLOW"]},
    "Endpoint": {"$ref": "#/definitions/Endpoint"},
    "FreeForm": {"type": "object"},
    "Labels": {
      "type": "object",
      "additionalProperties": {"type": "string"}
    },
    "Selector": {
      "oneOf": [
        {"type": "string"},
        {"type": "array", "items": {"type": "string"}}
      ]
    },
    "StructuredUnion": {
      "oneOf": [
        {"type": "array", "items": {"type": "string"}},
        {"type": "object"}
      ]
    },
    "Tags": {
      "type": "array",
      "items": {"$ref": "#/definitions/Tag"}
    }
  },
  "required": ["Name"],
  "primaryIdentifier": ["/properties/Name", "/properties/Region"],
  "readOnlyProperties": ["/properties/Id"],
  "createOnlyProperties": ["/properties/Name"],
  "writeOnlyProperties": ["/properties/Secret"],
  "tagging": {"taggable": true, "tagProperty": "/properties/Tags"}
}`

func convertSynthetic(t *testing.T) *irResource {
	t.Helper()
	schema, err := loadCFNSchema([]byte(syntheticSchema))
	if err != nil {
		t.Fatal(err)
	}
	resource, err := convert(schema, "aws/example/widget")
	if err != nil {
		t.Fatal(err)
	}
	return resource
}

func TestConvertStampsPointerLists(t *testing.T) {
	resource := convertSynthetic(t)
	root := resource.Schema

	if id := findAttribute(root, "id"); id == nil || !id.Computed {
		t.Error("readOnly /properties/Id should map to Computed")
	}
	if name := findAttribute(root, "name"); name == nil || !name.MustRecreate {
		t.Error("createOnly /properties/Name should map to MustRecreate")
	}
	secret := findAttribute(root, "secret")
	if secret == nil || !secret.Sensitive || !secret.IgnoreDrift {
		t.Error("writeOnly /properties/Secret should map to Sensitive + IgnoreDrift")
	}

	assertContains(t, resource.ComputedFields, "id")
	assertContains(t, resource.CreateOnlyFields, "name")
	assertContains(t, resource.WriteOnlyFields, "secret")
}

func TestConvertCompositePrimaryIdentifier(t *testing.T) {
	resource := convertSynthetic(t)
	if resource.PrimaryIdentifierField != "name" {
		t.Errorf("first identifier component should be the IDField, got %q", resource.PrimaryIdentifierField)
	}
	want := []string{"name", "region"}
	if len(resource.PrimaryIdentifierFields) != len(want) {
		t.Fatalf("composite primaryIdentifier should retain all components, got %v", resource.PrimaryIdentifierFields)
	}
	for i, field := range want {
		if resource.PrimaryIdentifierFields[i] != field {
			t.Errorf("PrimaryIdentifierFields[%d] = %q, want %q", i, resource.PrimaryIdentifierFields[i], field)
		}
	}
}

func TestConvertTypeMappings(t *testing.T) {
	root := convertSynthetic(t).Schema

	if mode := findAttribute(root, "mode"); mode == nil || len(mode.Enum) != 2 {
		t.Error("enum should map to AllowedValues")
	}
	endpoint := findAttribute(root, "endpoint")
	if endpoint == nil || endpoint.Type != "object" {
		t.Fatal("$ref Endpoint should resolve to an object")
	}
	if port := findAttribute(endpoint, "port"); port == nil || port.Type != "integer" || port.Minimum == nil {
		t.Error("nested $ref properties and numeric constraints should be preserved")
	}
	if free := findAttribute(root, "freeForm"); free == nil || free.Type != "map" || free.MapValues != nil {
		t.Error("free-form object should be modelled as an open map (nil MapValues)")
	}
	if labels := findAttribute(root, "labels"); labels == nil || labels.Type != "map" || labels.MapValues == nil {
		t.Error("additionalProperties object should map to a typed map")
	}
	if selector := findAttribute(root, "selector"); selector == nil || selector.Type != "union" {
		t.Error("oneOf should map to a union")
	}
	// A oneOf of a typed branch and a redundant free-form `{"type":"object"}`
	// branch collapses to the typed branch (the free-form branch would otherwise
	// make the whole field a JSON string and lose name translation).
	if su := findAttribute(root, "structuredUnion"); su == nil || su.Type != "array" {
		t.Error("oneOf with a typed branch and a free-form object branch should collapse to the typed branch")
	}
}

func TestConvertTagging(t *testing.T) {
	resource := convertSynthetic(t)
	if resource.TagPropertyName != "tags" {
		t.Errorf("tag property should be 'tags', got %q", resource.TagPropertyName)
	}
	if resource.TagShape != "cloudcontrol.TagShapeKeyValueList" {
		t.Errorf("array-of-objects tags should be a key/value list, got %q", resource.TagShape)
	}
}

func assertContains(t *testing.T, values []string, want string) {
	t.Helper()
	if slices.Contains(values, want) {
		return
	}
	t.Errorf("expected %v to contain %q", values, want)
}
