//go:build unit

package linkutils

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/stretchr/testify/assert"
)

func vpcConfigSpec() *core.MappingNode {
	return core.MappingNodeFields(
		"vpcConfig", core.MappingNodeFields(
			"subnetIds", &core.MappingNode{Items: []*core.MappingNode{
				core.MappingNodeFromString("subnet-1"),
			}},
		),
	)
}

func TestStageNetworkAccessKnownOnDeploy(t *testing.T) {
	t.Run("signals when the caller is VPC-attached via current state", func(t *testing.T) {
		changes := &provider.LinkChanges{}
		StageNetworkAccessKnownOnDeploy(&provider.Changes{
			AppliedResourceInfo: provider.ResourceInfo{
				ResourceName:         "apiFunction",
				CurrentResourceState: &state.ResourceState{SpecData: vpcConfigSpec()},
			},
		}, changes)
		assert.Equal(t, []string{"apiFunctionNetworkAccess"}, changes.FieldChangesKnownOnDeploy)
	})

	t.Run("signals when the caller is VPC-attached via resolved spec", func(t *testing.T) {
		changes := &provider.LinkChanges{}
		StageNetworkAccessKnownOnDeploy(&provider.Changes{
			AppliedResourceInfo: provider.ResourceInfo{
				ResourceName: "apiFunction",
				ResourceWithResolvedSubs: &provider.ResolvedResource{
					Spec: vpcConfigSpec(),
				},
			},
		}, changes)
		assert.Equal(t, []string{"apiFunctionNetworkAccess"}, changes.FieldChangesKnownOnDeploy)
	})

	t.Run("stays silent when the caller is not VPC-attached", func(t *testing.T) {
		changes := &provider.LinkChanges{}
		StageNetworkAccessKnownOnDeploy(&provider.Changes{
			AppliedResourceInfo: provider.ResourceInfo{
				ResourceName:         "apiFunction",
				CurrentResourceState: &state.ResourceState{SpecData: core.MappingNodeFields()},
			},
		}, changes)
		assert.Empty(t, changes.FieldChangesKnownOnDeploy)
	})

	t.Run("stays silent for an empty vpcConfig subnet list", func(t *testing.T) {
		changes := &provider.LinkChanges{}
		StageNetworkAccessKnownOnDeploy(&provider.Changes{
			AppliedResourceInfo: provider.ResourceInfo{
				ResourceName: "apiFunction",
				CurrentResourceState: &state.ResourceState{
					SpecData: core.MappingNodeFields(
						"vpcConfig", core.MappingNodeFields("subnetIds", &core.MappingNode{Items: []*core.MappingNode{}}),
					),
				},
			},
		}, changes)
		assert.Empty(t, changes.FieldChangesKnownOnDeploy)
	})

	t.Run("does not panic on nil caller changes", func(t *testing.T) {
		changes := &provider.LinkChanges{}
		StageNetworkAccessKnownOnDeploy(nil, changes)
		assert.Empty(t, changes.FieldChangesKnownOnDeploy)
	})
}
