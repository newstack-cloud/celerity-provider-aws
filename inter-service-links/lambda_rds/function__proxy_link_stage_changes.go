package lambdards

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *functionProxyLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	annotations := getProxyLinkAnnotations(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)
	functionName := pluginutils.GetResourceName(&input.ResourceAChanges.AppliedResourceInfo)
	prefix := proxyEnvVarPrefix(annotations.envVarPrefix, &input.ResourceBChanges.AppliedResourceInfo)
	names := proxyEnvVarNames(prefix)
	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	hostFieldPath := fmt.Sprintf("%s.environmentVariables[%q]", functionName, names.host)

	if !annotations.populateEnvVars {
		// Remove any env vars this link previously set.
		for _, name := range names.all() {
			path := fmt.Sprintf("%s.environmentVariables[%q]", functionName, name)
			if _, has := pluginutils.GetValueByPath(fmt.Sprintf("$.%s", path), currentLinkData); has {
				changes.RemovedFields = append(changes.RemovedFields, path)
			}
		}
	} else {
		// The host env var derives from the proxy endpoint (known on deploy when the proxy is new).
		if err := linkhelpers.CollectChanges(
			"$.spec.endpoint",
			fmt.Sprintf("$.%s", hostFieldPath),
			currentLinkData,
			input.ResourceBChanges,
			changes,
		); err != nil {
			return nil, err
		}
	}

	if annotations.authMode == "iam" &&
		(pluginutils.IsResourceNew(input.ResourceAChanges) || pluginutils.IsResourceNew(input.ResourceBChanges)) {
		// The rds-db:connect grant's resource ARN derives from the proxy, known on deploy.
		changes.FieldChangesKnownOnDeploy = append(
			changes.FieldChangesKnownOnDeploy,
			createLinkDataExecutionRoleName(&input.ResourceAChanges.AppliedResourceInfo),
		)
	}

	// Networking activation for a VPC-attached caller is configured at deploy; surface it as a
	// best-effort known-on-deploy signal.
	linkutils.StageNetworkAccessKnownOnDeploy(input.ResourceAChanges, changes)

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
