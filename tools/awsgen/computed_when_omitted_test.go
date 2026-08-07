package main

import (
	"strings"
	"testing"
)

const autoNamedSchema = `{
  "typeName": "AWS::Example::AutoNamed",
  "description": "A synthetic type exercising auto-named identifier mapping.",
  "properties": {
    "WidgetName": {"type": "string"},
    "Region": {"type": "string"},
    "Arn": {"type": "string"}
  },
  "required": ["Region"],
  "primaryIdentifier": ["/properties/WidgetName", "/properties/Region"],
  "readOnlyProperties": ["/properties/Arn"],
  "createOnlyProperties": ["/properties/WidgetName"]
}`

func TestConvertMarksOptionalPrimaryIdentifierAsComputedWhenOmitted(t *testing.T) {
	schema, err := loadCFNSchema([]byte(autoNamedSchema))
	if err != nil {
		t.Fatal(err)
	}
	resource, err := convert(schema, "aws/example/autoNamed")
	if err != nil {
		t.Fatal(err)
	}

	widgetName := mustFindAttribute(t, resource.Schema, "widgetName")
	if !widgetName.ComputedWhenOmitted {
		t.Error("optional primary identifier component should be computed-when-omitted")
	}

	region := mustFindAttribute(t, resource.Schema, "region")
	if region.ComputedWhenOmitted {
		t.Error("required primary identifier component should not be computed-when-omitted")
	}

	arn := mustFindAttribute(t, resource.Schema, "arn")
	if arn.ComputedWhenOmitted || !arn.Computed {
		t.Error("read-only fields should stay plain computed")
	}

	source, err := emitResourceFile(resource)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "ComputedWhenOmitted: true,") {
		t.Error("emitted source should carry the ComputedWhenOmitted flag")
	}
}
