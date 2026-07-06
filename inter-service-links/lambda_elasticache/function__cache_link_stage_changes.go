package lambdaelasticache

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *functionCacheLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	annotations := getCacheLinkAnnotations(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)
	functionName := pluginutils.GetResourceName(&input.ResourceAChanges.AppliedResourceInfo)
	prefix := cacheEnvVarPrefix(annotations.envVarPrefix, &input.ResourceBChanges.AppliedResourceInfo)
	names := cacheEnvVarNames(prefix)
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
		// The host env var derives from the cache endpoint, the configuration endpoint in cluster
		// mode, otherwise the primary endpoint (known on deploy when the cache is new).
		if err := linkhelpers.CollectChanges(
			cacheEndpointSourcePath(input.ResourceBChanges),
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
		// The elasticache:Connect grant's resource ARNs derive from the cache, known on deploy.
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

// cacheEndpointSourcePath picks the resolved-spec source path for the host env var: the
// configuration endpoint when the cache exposes one (cluster mode), otherwise the primary endpoint.
func cacheEndpointSourcePath(resourceBChanges *provider.Changes) string {
	resolved := pluginutils.GetResolvedResourceSpecData(resourceBChanges)
	if node, has := pluginutils.GetValueByPath("$.configurationEndPoint.address", resolved); has && node != nil {
		return "$.spec.configurationEndPoint.address"
	}
	return "$.spec.primaryEndPoint.address"
}
