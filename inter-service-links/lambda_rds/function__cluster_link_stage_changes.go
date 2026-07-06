package lambdards

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *functionClusterLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	annotations := getClusterLinkAnnotations(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)
	functionName := pluginutils.GetResourceName(&input.ResourceAChanges.AppliedResourceInfo)
	prefix := clusterEnvVarPrefix(annotations.envVarPrefix, &input.ResourceBChanges.AppliedResourceInfo)
	names := clusterEnvVarNames(prefix)
	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	hostFieldPath := fmt.Sprintf("%s.environmentVariables[%q]", functionName, names.host)
	readerHostFieldPath := fmt.Sprintf("%s.environmentVariables[%q]", functionName, names.readerHost)

	if !annotations.populateEnvVars {
		// Remove any env vars this link previously set.
		for _, name := range names.all() {
			path := fmt.Sprintf("%s.environmentVariables[%q]", functionName, name)
			if _, has := pluginutils.GetValueByPath(fmt.Sprintf("$.%s", path), currentLinkData); has {
				changes.RemovedFields = append(changes.RemovedFields, path)
			}
		}
	} else {
		// The host env var derives from the cluster writer endpoint (known on deploy when the
		// cluster is new).
		if err := linkhelpers.CollectChanges(
			"$.spec.endpoint.address",
			fmt.Sprintf("$.%s", hostFieldPath),
			currentLinkData,
			input.ResourceBChanges,
			changes,
		); err != nil {
			return nil, err
		}

		if err := collectReaderEndpointChanges(
			annotations,
			readerHostFieldPath,
			currentLinkData,
			input,
			changes,
		); err != nil {
			return nil, err
		}

	}

	if annotations.authMode == "iam" &&
		(pluginutils.IsResourceNew(input.ResourceAChanges) || pluginutils.IsResourceNew(input.ResourceBChanges)) {
		// The rds-db:connect grant's resource ARN derives from the cluster, known on deploy.
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

func collectReaderEndpointChanges(
	annotations *clusterLinkAnnotations,
	readerHostFieldPath string,
	currentLinkData *core.MappingNode,
	input *provider.LinkStageChangesInput,
	changes *provider.LinkChanges,
) error {
	if annotations.readerEndpoint {
		// The reader host env var derives from the cluster reader endpoint.
		if err := linkhelpers.CollectChanges(
			"$.spec.readEndpoint.address",
			fmt.Sprintf("$.%s", readerHostFieldPath),
			currentLinkData,
			input.ResourceBChanges,
			changes,
		); err != nil {
			return err
		}
	}

	return nil
}
