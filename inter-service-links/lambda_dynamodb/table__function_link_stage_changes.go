package lambdadynamodb

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *dynamoDBTableLambdaFunctionLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	// Check if either resource is being (re)created
	if pluginutils.IsResourceNew(input.ResourceAChanges) ||
		pluginutils.IsResourceNew(input.ResourceBChanges) {
		// When either resource is (re)created, the event source mapping
		// will need to be recreated with new ARNs.
		changes.FieldChangesKnownOnDeploy = append(
			changes.FieldChangesKnownOnDeploy,
			"eventSourceMapping",
		)
	}

	// Track changes to the DynamoDB table's stream ARN
	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)
	err := linkhelpers.CollectChanges(
		"$.spec.latestStreamArn",
		"$.eventSourceMapping.eventSourceArn",
		currentLinkData,
		input.ResourceAChanges,
		changes,
	)
	if err != nil {
		return nil, err
	}

	// Track changes to the Lambda function's ARN
	err = linkhelpers.CollectChanges(
		"$.spec.arn",
		"$.eventSourceMapping.functionArn",
		currentLinkData,
		input.ResourceBChanges,
		changes,
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
