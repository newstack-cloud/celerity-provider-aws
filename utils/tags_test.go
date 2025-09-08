//go:build unit

package utils

import (
	"fmt"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/stretchr/testify/suite"
)

type TagsSuite struct {
	suite.Suite
}

func (s *TagsSuite) Test_diff_tags_for_list_tags() {
	changes := &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"tags": {
							Items: []*core.MappingNode{
								tagMappingNode("key1", "value1"),
								tagMappingNode("key2", "value2-updated"),
								tagMappingNode("key3", "value3-updated"),
							},
						},
					},
				},
			},
			CurrentResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"tags": {
							Items: []*core.MappingNode{
								tagMappingNode("key1", "value1"),
								tagMappingNode("key2", "value2"),
								tagMappingNode("key3", "value3"),
								tagMappingNode("key4", "value4"),
								tagMappingNode("key5", "value5"),
							},
						},
					},
				},
			},
		},
	}
	diffResult := DiffTags(changes, "$.tags", func(tag *Tag) string {
		return fmt.Sprintf("%s:%s", tag.Key, tag.Value)
	})
	s.Equal(&TagsDiffResult[string]{
		ToSet: []string{
			"key1:value1",
			"key2:value2-updated",
			"key3:value3-updated",
		},
		ToRemove: []string{
			"key4",
			"key5",
		},
	}, diffResult)
}

func (s *TagsSuite) Test_diff_tags_for_map_tags() {
	changes := &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"tags": {
							Fields: map[string]*core.MappingNode{
								"key1": core.MappingNodeFromString("value1"),
								"key2": core.MappingNodeFromString("value2-updated"),
								"key3": core.MappingNodeFromString("value3-updated"),
							},
						},
					},
				},
			},
			CurrentResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"tags": {
							Fields: map[string]*core.MappingNode{
								"key1": core.MappingNodeFromString("value1"),
								"key2": core.MappingNodeFromString("value2"),
								"key3": core.MappingNodeFromString("value3"),
								"key4": core.MappingNodeFromString("value4"),
								"key5": core.MappingNodeFromString("value5"),
							},
						},
					},
				},
			},
		},
	}

	diffResult := DiffTags(changes, "$.tags", func(tag *Tag) string {
		return fmt.Sprintf("%s:%s", tag.Key, tag.Value)
	})
	s.Equal(&TagsDiffResult[string]{
		ToSet: []string{
			"key1:value1",
			"key2:value2-updated",
			"key3:value3-updated",
		},
		ToRemove: []string{
			"key4",
			"key5",
		},
	}, diffResult)
}

func tagMappingNode(key, value string) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"key":   core.MappingNodeFromString(key),
			"value": core.MappingNodeFromString(value),
		},
	}
}

func TestTagsSuite(t *testing.T) {
	suite.Run(t, new(TagsSuite))
}
