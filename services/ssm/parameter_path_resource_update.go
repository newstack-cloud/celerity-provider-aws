package ssm

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Update is a no-op: the path is the only field and changing it recreates the resource,
// so there is nothing left to update in place.
func (a *parameterPathResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{},
	}, nil
}
