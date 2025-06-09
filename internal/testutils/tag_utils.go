package testutils

import (
	"github.com/stretchr/testify/assert"
	"github.com/two-hundred/celerity/libs/blueprint/core"
)

// CompareTags compares two slices of tag MappingNodes in an order-independent way.
// It extracts key-value pairs using core.StringValue and compares them as maps.
func CompareTags(t assert.TestingT, expectedTags, actualTags []*core.MappingNode) {
	assert.Equal(t, len(expectedTags), len(actualTags), "tag count should match")

	// Helper to extract key-value from MappingNode
	tagKV := func(node *core.MappingNode) (string, string) {
		return core.StringValue(node.Fields["key"]), core.StringValue(node.Fields["value"])
	}

	// Build maps for comparison
	expectedMap := map[string]string{}
	for _, tag := range expectedTags {
		k, v := tagKV(tag)
		expectedMap[k] = v
	}
	actualMap := map[string]string{}
	for _, tag := range actualTags {
		k, v := tagKV(tag)
		actualMap[k] = v
	}
	assert.Equal(t, expectedMap, actualMap, "tags should match regardless of order")
}
