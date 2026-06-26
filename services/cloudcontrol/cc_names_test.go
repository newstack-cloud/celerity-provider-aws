//go:build unit

package cloudcontrol

import (
	"encoding/json"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func namesTestSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:  provider.ResourceDefinitionsSchemaTypeObject,
		Label: "Table",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"tableName": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Table Name"},
			"sseSpecification": {
				Type:  provider.ResourceDefinitionsSchemaTypeObject,
				Label: "SSE Specification",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"sseEnabled":     {Type: provider.ResourceDefinitionsSchemaTypeBoolean, Label: "SSE Enabled"},
					"kmsMasterKeyId": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "KMS Master Key Id"},
				},
			},
			"replicas": {
				Type:  provider.ResourceDefinitionsSchemaTypeArray,
				Label: "Replicas",
				Items: &provider.ResourceDefinitionsSchema{
					Type:  provider.ResourceDefinitionsSchemaTypeObject,
					Label: "Replica",
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"region": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Region"},
						"sseSpecification": {
							Type:  provider.ResourceDefinitionsSchemaTypeObject,
							Label: "SSE Specification",
							Attributes: map[string]*provider.ResourceDefinitionsSchema{
								"sseEnabled": {Type: provider.ResourceDefinitionsSchemaTypeBoolean, Label: "SSE Enabled"},
							},
						},
					},
				},
			},
		},
	}
}

func namesTestOverrides() map[string]string {
	return map[string]string{
		"sseSpecification":                       "SSESpecification",
		"sseSpecification.sseEnabled":            "SSEEnabled",
		"sseSpecification.kmsMasterKeyId":        "KMSMasterKeyId",
		"replicas.*.sseSpecification":            "SSESpecification",
		"replicas.*.sseSpecification.sseEnabled": "SSEEnabled",
	}
}

func namesTestSpec() *core.MappingNode {
	return &core.MappingNode{Fields: map[string]*core.MappingNode{
		"tableName": core.MappingNodeFromString("orders"),
		"sseSpecification": {Fields: map[string]*core.MappingNode{
			"sseEnabled":     core.MappingNodeFromBool(true),
			"kmsMasterKeyId": core.MappingNodeFromString("key-1"),
		}},
		"replicas": {Items: []*core.MappingNode{
			{Fields: map[string]*core.MappingNode{
				"region": core.MappingNodeFromString("eu-west-2"),
				"sseSpecification": {Fields: map[string]*core.MappingNode{
					"sseEnabled": core.MappingNodeFromBool(false),
				}},
			}},
		}},
	}}
}

func Test_SpecToCFN_applies_overrides(t *testing.T) {
	schema := namesTestSchema()
	meta := CCResourceMeta{FieldNameOverrides: namesTestOverrides()}
	cfn := SpecToCFN(namesTestSpec(), schema, meta)

	sse := cfn.Fields["SSESpecification"]
	if sse == nil {
		t.Fatalf("expected SSESpecification key, got fields %v", keys(cfn))
	}
	if _, ok := sse.Fields["KMSMasterKeyId"]; !ok {
		t.Errorf("expected KMSMasterKeyId key, got %v", keys(sse))
	}
	if _, ok := sse.Fields["SSEEnabled"]; !ok {
		t.Errorf("expected SSEEnabled key, got %v", keys(sse))
	}
	replicaSSE := cfn.Fields["Replicas"].Items[0].Fields["SSESpecification"]
	if replicaSSE == nil {
		t.Fatalf("expected nested SSESpecification key")
	}
	if _, ok := replicaSSE.Fields["SSEEnabled"]; !ok {
		t.Errorf("expected nested SSEEnabled key, got %v", keys(replicaSSE))
	}
}

func Test_names_round_trip(t *testing.T) {
	schema := namesTestSchema()
	meta := CCResourceMeta{FieldNameOverrides: namesTestOverrides()}
	spec := namesTestSpec()

	cfn := SpecToCFN(spec, schema, meta)
	back := CFNToSpec(cfn, schema, meta)

	for _, path := range []string{
		"tableName",
		"sseSpecification.sseEnabled",
		"sseSpecification.kmsMasterKeyId",
	} {
		if !hasPath(back, path) {
			t.Errorf("round-trip lost path %q; got %v", path, back.Fields)
		}
	}
	replicaSSE := back.Fields["replicas"].Items[0].Fields["sseSpecification"]
	if replicaSSE == nil || replicaSSE.Fields["sseEnabled"] == nil {
		t.Errorf("round-trip lost nested replicas.*.sseSpecification.sseEnabled")
	}
}

func Test_jsonStringField_to_cfn_parses_object(t *testing.T) {
	schema := &provider.ResourceDefinitionsSchema{
		Type:  provider.ResourceDefinitionsSchemaTypeObject,
		Label: "Role",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"policies": {
				Type:  provider.ResourceDefinitionsSchemaTypeArray,
				Label: "Policies",
				Items: &provider.ResourceDefinitionsSchema{
					Type:  provider.ResourceDefinitionsSchemaTypeObject,
					Label: "Policy",
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"policyName":     {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Policy Name"},
						"policyDocument": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Policy Document"},
					},
				},
			},
		},
	}
	meta := CCResourceMeta{JSONStringFields: []string{"policies.*.policyDocument"}}

	spec := &core.MappingNode{Fields: map[string]*core.MappingNode{
		"policies": {Items: []*core.MappingNode{
			{Fields: map[string]*core.MappingNode{
				"policyName":     core.MappingNodeFromString("access"),
				"policyDocument": core.MappingNodeFromString(`{"Version":"2012-10-17","Statement":[{"Sid":"s1","Effect":"Allow"}]}`),
			}},
		}},
	}}

	cfn := SpecToCFN(spec, schema, meta)
	doc := cfn.Fields["Policies"].Items[0].Fields["PolicyDocument"]
	if doc == nil || doc.Fields == nil {
		t.Fatalf("expected PolicyDocument parsed to an object, got %#v", doc)
	}
	if got := core.StringValue(doc.Fields["Version"]); got != "2012-10-17" {
		t.Errorf("expected parsed Version, got %q", got)
	}

	// Round-trip back: the object must serialise to a JSON string again.
	back := CFNToSpec(cfn, schema, meta)
	backDoc := back.Fields["policies"].Items[0].Fields["policyDocument"]
	if backDoc == nil || backDoc.Scalar == nil || backDoc.Scalar.StringValue == nil {
		t.Fatalf("expected policyDocument serialised back to a string, got %#v", backDoc)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(core.StringValue(backDoc)), &parsed); err != nil {
		t.Errorf("round-tripped policyDocument is not valid JSON: %v", err)
	}
	if parsed["Version"] != "2012-10-17" {
		t.Errorf("round-tripped policyDocument lost Version: %v", parsed)
	}
}

func keys(node *core.MappingNode) []string {
	var out []string
	for k := range node.Fields {
		out = append(out, k)
	}
	return out
}

func hasPath(node *core.MappingNode, dotted string) bool {
	current := node
	for _, segment := range splitDots(dotted) {
		if current == nil || current.Fields == nil {
			return false
		}
		current = current.Fields[segment]
	}
	return current != nil
}

func splitDots(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}
