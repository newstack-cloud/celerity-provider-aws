package ssm

import (
	"context"
	"errors"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// GetExternalState echoes the spec: a parameter hierarchy has no external representation
// to fetch, so the namespace always "exists" and can never drift.
func (a *parameterPathResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	pathValue, hasPath := pluginutils.GetValueByPath("$.path", input.CurrentResourceSpec)
	if !hasPath {
		return nil, errors.New("path is required to get external state for an SSM parameter path")
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"path": core.MappingNodeFromString(core.StringValue(pathValue)),
			},
		},
	}, nil
}
