package lambdalinks

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionFunctionLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	annotations := getLambdaFunctionLinkAnnotations(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)

	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)
	envVarFieldPath := fmt.Sprintf(
		"%s.environmentVariables[\"%s\"]",
		pluginutils.GetResourceName(&input.ResourceAChanges.AppliedResourceInfo),
		annotations.envVarName,
	)
	_, linkDataHasEnvVar := pluginutils.GetValueByPath(
		fmt.Sprintf("$.%s", envVarFieldPath),
		currentLinkData,
	)

	if !annotations.populateEnvVars && linkDataHasEnvVar {
		changes.RemovedFields = append(changes.RemovedFields, envVarFieldPath)
	} else {
		targetFunctionFieldPath := fmt.Sprintf(
			"$.%s",
			envVarFieldPath,
		)
		err := linkhelpers.CollectChanges(
			"$.spec.arn",
			targetFunctionFieldPath,
			linkhelpers.GetLinkDataFromState(input.CurrentLinkState),
			input.ResourceAChanges,
			changes,
		)
		if err != nil {
			return nil, err
		}
	}

	if pluginutils.IsResourceNew(input.ResourceAChanges) ||
		pluginutils.IsResourceNew(input.ResourceBChanges) {
		// Indicate that when either of the resources will be (re)created,
		// the execution role permission data specific to this link will change
		// but the value will not be known until deployment.
		// As the permissions are to enable one lambda to invoke another,
		// the ARN has to change to trigger a change to the permissions.
		changes.FieldChangesKnownOnDeploy = append(
			changes.FieldChangesKnownOnDeploy,
			createLinkDataExecutionRoleName(&input.ResourceAChanges.AppliedResourceInfo),
		)
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
