package ssm

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Stabilised always reports true: PutParameter provisions synchronously and returns no
// eventually-consistent computed fields (the version is known immediately), so there is no
// stabilisation phase to wait on.
func (a *parameterResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
