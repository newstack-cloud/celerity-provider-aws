package eventslinks

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *ruleAPIDestinationLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	if pluginutils.IsResourceNew(input.ResourceAChanges) ||
		pluginutils.IsResourceNew(input.ResourceBChanges) {
		// When either resource is (re)created, the InvokeApiDestination permission
		// attached to the target's role will change, but the value depends on the
		// API destination ARN which is only known at deploy time.
		changes.FieldChangesKnownOnDeploy = append(
			changes.FieldChangesKnownOnDeploy,
			createLinkDataRoleName(&input.ResourceAChanges.AppliedResourceInfo),
		)
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
