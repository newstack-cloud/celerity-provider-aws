package lambdasqs

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *functionQueueLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	annotations := getQueueLinkAnnotations(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)

	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)
	envVarName := queueEnvVarName(
		annotations.envVarName,
		&input.ResourceBChanges.AppliedResourceInfo,
	)
	envVarFieldPath := fmt.Sprintf(
		"%s.environmentVariables[\"%s\"]",
		pluginutils.GetResourceName(&input.ResourceAChanges.AppliedResourceInfo),
		envVarName,
	)
	_, linkDataHasEnvVar := pluginutils.GetValueByPath(
		fmt.Sprintf("$.%s", envVarFieldPath),
		currentLinkData,
	)

	if !annotations.populateEnvVars && linkDataHasEnvVar {
		changes.RemovedFields = append(changes.RemovedFields, envVarFieldPath)
	} else if annotations.populateEnvVars {
		err := linkhelpers.CollectChanges(
			"$.spec.queueUrl",
			fmt.Sprintf("$.%s", envVarFieldPath),
			currentLinkData,
			input.ResourceBChanges,
			changes,
		)
		if err != nil {
			return nil, err
		}
	}

	if pluginutils.IsResourceNew(input.ResourceAChanges) ||
		pluginutils.IsResourceNew(input.ResourceBChanges) {
		// The execution-role permission data for this link references the queue ARN and is only
		// known at deploy time.
		changes.FieldChangesKnownOnDeploy = append(
			changes.FieldChangesKnownOnDeploy,
			createLinkDataExecutionRoleName(&input.ResourceAChanges.AppliedResourceInfo),
		)
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
