package ssm

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Stabilised always reports true: nothing is provisioned, so there is nothing to wait on.
func (a *parameterPathResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
