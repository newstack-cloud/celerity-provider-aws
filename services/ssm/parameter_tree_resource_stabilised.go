package ssm

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Stabilised always reports true: PutParameter provisions synchronously and the tree's
// computed metadata is known as soon as the deploy operations return, so there is no
// stabilisation phase to wait on.
func (a *parameterTreeResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
