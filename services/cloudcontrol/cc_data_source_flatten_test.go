//go:build unit

package cloudcontrol

import (
	"encoding/json"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type CCDataSourceFlattenSuite struct {
	suite.Suite
}

func (s *CCDataSourceFlattenSuite) Test_schema_aware_flatten() {
	str := func() *provider.ResourceDefinitionsSchema {
		return &provider.ResourceDefinitionsSchema{Type: provider.ResourceDefinitionsSchemaTypeString}
	}
	schema := &provider.ResourceDefinitionsSchema{
		Type: provider.ResourceDefinitionsSchemaTypeObject,
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"name": str(),
			"nested": {
				Type: provider.ResourceDefinitionsSchemaTypeObject,
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"count": {Type: provider.ResourceDefinitionsSchemaTypeInteger},
				},
			},
			"items": {Type: provider.ResourceDefinitionsSchemaTypeArray, Items: str()},
			"labels": {
				Type:      provider.ResourceDefinitionsSchemaTypeMap,
				MapValues: str(),
			},
		},
	}
	node := core.MappingNodeFields(
		"name", core.MappingNodeFromString("x"),
		"nested", core.MappingNodeFields("count", intNode(3)),
		"items", &core.MappingNode{Items: []*core.MappingNode{
			core.MappingNodeFromString("a"), core.MappingNodeFromString("b"),
		}},
		"labels", core.MappingNodeFields(
			"env", core.MappingNodeFromString("prod"),
			"tier", core.MappingNodeFromString("1"),
		),
	)

	out := FlattenForDataSource(node, schema)

	// Top-level scalar.
	s.Equal("x", core.StringValue(out["name"]))
	// Nested object scalar -> dot-notation key.
	s.Equal(3, core.IntValue(out["nested.count"]))
	// Array kept whole under one key (not recursed, not the map case).
	s.Require().NotNil(out["items"])
	s.Len(out["items"].Items, 2)
	// Map serialised to a JSON string (not dot-flattened by its dynamic keys).
	s.NotContains(out, "labels.env")
	labels := out["labels"]
	s.Require().NotNil(labels)
	s.Require().NotNil(labels.Scalar)
	s.Require().NotNil(labels.Scalar.StringValue)
	var decoded map[string]string
	s.Require().NoError(json.Unmarshal([]byte(*labels.Scalar.StringValue), &decoded))
	s.Equal(map[string]string{"env": "prod", "tier": "1"}, decoded)
}

func TestCCDataSourceFlattenSuite(t *testing.T) {
	suite.Run(t, new(CCDataSourceFlattenSuite))
}
