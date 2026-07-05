package lambdakms

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *functionKeyLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	annotations := getKeyLinkAnnotations(
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
	} else if annotations.populateEnvVars {
		targetKeyFieldPath := fmt.Sprintf("$.%s", envVarFieldPath)
		// The environment variable holds the key ARN.
		err := linkhelpers.CollectChanges(
			"$.spec.arn",
			targetKeyFieldPath,
			linkhelpers.GetLinkDataFromState(input.CurrentLinkState),
			input.ResourceBChanges,
			changes,
		)
		if err != nil {
			return nil, err
		}
	}

	if pluginutils.IsResourceNew(input.ResourceAChanges) ||
		pluginutils.IsResourceNew(input.ResourceBChanges) {
		// When either resource will be (re)created, the execution-role permission data
		// specific to this link will change but the value will not be known until
		// deployment, as the granted key ARN derives from the key resource.
		changes.FieldChangesKnownOnDeploy = append(
			changes.FieldChangesKnownOnDeploy,
			createLinkDataExecutionRoleName(&input.ResourceAChanges.AppliedResourceInfo),
		)
	}

	stageKeyGrantChanges(input, annotations, currentLinkData, changes)

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}

// This surfaces the managed KMS grant side effect in staged changes: creating
// or updating the grant when manageKeyGrant is enabled, or revoking it when disabled. The
// grant's name and operations are known at plan time, so the change carries a concrete value.
func stageKeyGrantChanges(
	input *provider.LinkStageChangesInput,
	annotations *keyLinkAnnotations,
	currentLinkData *core.MappingNode,
	changes *provider.LinkChanges,
) {
	currentGrant, hasCurrentGrant := pluginutils.GetValueByPath(
		fmt.Sprintf("$.%s", keyGrantLinkDataField),
		currentLinkData,
	)

	if !annotations.manageKeyGrant {
		if hasCurrentGrant {
			changes.RemovedFields = append(changes.RemovedFields, keyGrantLinkDataField)
		}
		return
	}

	desiredGrant := keyGrantLinkDataNode(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
		annotations.accessLevel,
	)

	switch {
	case !hasCurrentGrant:
		changes.NewFields = append(changes.NewFields, &provider.FieldChange{
			FieldPath: keyGrantLinkDataField,
			NewValue:  desiredGrant,
		})
	case !core.MappingNodeEqual(currentGrant, desiredGrant):
		changes.ModifiedFields = append(changes.ModifiedFields, &provider.FieldChange{
			FieldPath: keyGrantLinkDataField,
			PrevValue: currentGrant,
			NewValue:  desiredGrant,
		})
	default:
		changes.UnchangedFields = append(changes.UnchangedFields, keyGrantLinkDataField)
	}
}
