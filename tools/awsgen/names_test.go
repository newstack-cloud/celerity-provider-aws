package main

import "testing"

func TestCamelFromCFN(t *testing.T) {
	cases := map[string]string{
		"QueueName":           "queueName",
		"DeadLetterTargetArn": "deadLetterTargetArn",
		"KMSMasterKeyId":      "kmsMasterKeyId",
		"SSESpecification":    "sseSpecification",
		"SSEEnabled":          "sseEnabled",
		"SSEType":             "sseType",
		"ARN":                 "arn",
		"A":                   "a",
		"":                    "",
	}
	for in, want := range cases {
		if got := camelFromCFN(in); got != want {
			t.Errorf("camelFromCFN(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestCollectNameOverrides verifies overrides are recorded only for fields whose
// CFN name a first-character flip cannot recover, keyed by camelCase path with "*"
// for array/map descent.
func TestCollectNameOverrides(t *testing.T) {
	root := &irSchema{
		Type: "object",
		Attributes: []irAttribute{
			{Name: "queueName", CFNName: "QueueName", Schema: &irSchema{Type: "string"}},
			{Name: "sseSpecification", CFNName: "SSESpecification", Schema: &irSchema{
				Type: "object",
				Attributes: []irAttribute{
					{Name: "sseEnabled", CFNName: "SSEEnabled", Schema: &irSchema{Type: "boolean"}},
					{Name: "kmsMasterKeyId", CFNName: "KMSMasterKeyId", Schema: &irSchema{Type: "string"}},
				},
			}},
			{Name: "replicas", CFNName: "Replicas", Schema: &irSchema{
				Type: "array",
				Items: &irSchema{
					Type: "object",
					Attributes: []irAttribute{
						{Name: "region", CFNName: "Region", Schema: &irSchema{Type: "string"}},
						{Name: "sseSpecification", CFNName: "SSESpecification", Schema: &irSchema{Type: "object"}},
					},
				},
			}},
		},
	}

	got := map[string]string{}
	collectNameOverrides(root, "", got)

	want := map[string]string{
		"sseSpecification":                "SSESpecification",
		"sseSpecification.sseEnabled":     "SSEEnabled",
		"sseSpecification.kmsMasterKeyId": "KMSMasterKeyId",
		"replicas.*.sseSpecification":     "SSESpecification",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d overrides, want %d: %v", len(got), len(want), got)
	}
	for path, cfn := range want {
		if got[path] != cfn {
			t.Errorf("override[%q] = %q, want %q", path, got[path], cfn)
		}
	}
}
