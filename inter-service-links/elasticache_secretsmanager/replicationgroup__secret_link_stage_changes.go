package elasticachesecretsmanager

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// StageChanges reports the link-data changes for applying the AUTH token. The token itself is
// applied at deploy time via ModifyReplicationGroup, so the only staged change is the
// authToken marker, which is known on deploy (its value derives from the linked secret and the
// deploy-time strategy). No secret value is ever staged.
func (l *replicationGroupSecretLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	// The AUTH token is read from the secret and applied to the replication group at deploy
	// time; the resulting link-data marker cannot be known until then, and it always (re)runs
	// when either resource is (re)created.
	if pluginutils.IsResourceNew(input.ResourceAChanges) ||
		pluginutils.IsResourceNew(input.ResourceBChanges) {
		changes.FieldChangesKnownOnDeploy = append(
			changes.FieldChangesKnownOnDeploy,
			"authToken",
		)
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
