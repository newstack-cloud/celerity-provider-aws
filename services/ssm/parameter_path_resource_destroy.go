package ssm

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Destroy is a no-op: there is no AWS object to delete. Parameters beneath the path are
// independent resources with their own lifecycles, and links that granted access to the
// path revoke their grants through their own destroy handling.
func (a *parameterPathResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	return nil
}
