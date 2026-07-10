package ssm

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Create is a no-op against AWS: a parameter hierarchy has no API object of its own and
// exists implicitly through the parameters beneath it. The deploy engine persists the
// resolved spec (the path) into state, which is all links need.
func (a *parameterPathResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{},
	}, nil
}
