//go:build integration

package e2e

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/newstack-cloud/bluelink/libs/blueprint/changes"
	"github.com/newstack-cloud/bluelink/libs/blueprint/container"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/includes"
	"github.com/newstack-cloud/bluelink/libs/blueprint/subengine"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// Determines how long we wait for any single message from the engine
// before giving up; generous because real AWS create/destroy/stabilise can be slow.
const channelTimeout = 30 * time.Minute

func newChangeStagingChannels() *container.ChangeStagingChannels {
	return &container.ChangeStagingChannels{
		ResourceChangesChan: make(chan container.ResourceChangesMessage),
		ChildChangesChan:    make(chan container.ChildChangesMessage),
		LinkChangesChan:     make(chan container.LinkChangesMessage),
		CompleteChan:        make(chan changes.BlueprintChanges),
		ErrChan:             make(chan error),
	}
}

func consumeStage(t *testing.T, channels *container.ChangeStagingChannels) *changes.BlueprintChanges {
	t.Helper()
	changeSet, err := consumeStageErr(channels)
	require.NoError(t, err, "stage changes")
	return changeSet
}

func consumeStageErr(channels *container.ChangeStagingChannels) (*changes.BlueprintChanges, error) {
	for {
		select {
		case <-channels.ResourceChangesChan:
		case <-channels.ChildChangesChan:
		case <-channels.LinkChangesChan:
		case changeSet := <-channels.CompleteChan:
			return &changeSet, nil
		case err := <-channels.ErrChan:
			return nil, err
		case <-time.After(channelTimeout):
			return nil, errors.New("timed out waiting for change staging to complete")
		}
	}
}

func consumeDeploy(t *testing.T, channels *container.DeployChannels) container.DeploymentFinishedMessage {
	t.Helper()
	finished, err := consumeDeployErr(channels)
	require.NoError(t, err, "deploy")
	return finished
}

func consumeDeployErr(channels *container.DeployChannels) (container.DeploymentFinishedMessage, error) {
	for {
		select {
		case <-channels.ResourceUpdateChan:
		case <-channels.ChildUpdateChan:
		case <-channels.LinkUpdateChan:
		case <-channels.DeploymentUpdateChan:
		case finished := <-channels.FinishChan:
			return finished, nil
		case err := <-channels.ErrChan:
			return container.DeploymentFinishedMessage{}, err
		case <-time.After(channelTimeout):
			return container.DeploymentFinishedMessage{}, errors.New("timed out waiting for deployment to finish")
		}
	}
}

// Datisfies includes.ChildResolver. The e2e blueprints contain no
// child includes, so Resolve should never be called; it errors if it ever is.
type fsChildResolver struct {
	fs afero.Fs
}

func (r *fsChildResolver) Resolve(
	ctx context.Context,
	includeName string,
	include *subengine.ResolvedInclude,
	params core.BlueprintParams,
) (*includes.ChildBlueprintInfo, error) {
	return nil, fmt.Errorf("e2e blueprints do not support child includes (include %q)", includeName)
}
