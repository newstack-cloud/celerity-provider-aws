package elasticachesecretsmanager

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Validate reports an error, at blueprint validation time, when the linked replication group does
// not enable in-transit encryption. Redis AUTH requires transitEncryptionEnabled to be true;
// applying an AUTH token to a group without it fails at deploy time with a validation error from
// ElastiCache, so this turns that late failure into a plan-time diagnostic.
func (l *replicationGroupSecretLinkActions) Validate(
	ctx context.Context,
	input *provider.LinkValidateInput,
) (*provider.LinkValidateOutput, error) {
	if transitEncryptionEnabled(input.ResourceASpec) {
		return &provider.LinkValidateOutput{}, nil
	}

	return &provider.LinkValidateOutput{
		Diagnostics: []*core.Diagnostic{
			{
				Level: core.DiagnosticLevelError,
				Message: fmt.Sprintf(
					"The ElastiCache replication group %q links to the Secrets Manager secret %q to "+
						"apply a Redis AUTH token, but Redis AUTH requires in-transit encryption. Set "+
						"\"transitEncryptionEnabled\" to true on the replication group, otherwise applying "+
						"the AUTH token will fail at deploy time.",
					input.ResourceAName,
					input.ResourceBName,
				),
			},
		},
	}, nil
}

func transitEncryptionEnabled(replicationGroupSpec *core.MappingNode) bool {
	value, has := pluginutils.GetValueByPath(
		"$.transitEncryptionEnabled",
		replicationGroupSpec,
	)
	return has && core.BoolValue(value)
}
