package sqslinks

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *sqsQueueQueueLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	sourceQueueResourceName := linkhelpers.GetResourceNameFromChanges(input.ResourceAChanges)

	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	// Collect deadLetterTargetArn changes from the dead-letter queue's ARN
	deadLetterTargetArnWriteToPath := fmt.Sprintf(
		"$[%q].redrivePolicy.deadLetterTargetArn",
		sourceQueueResourceName,
	)
	err := linkhelpers.CollectChanges(
		"$.spec.arn",
		deadLetterTargetArnWriteToPath,
		currentLinkData,
		input.ResourceBChanges,
		changes,
	)
	if err != nil {
		return nil, err
	}

	// Collect maxReceiveCount changes from annotations
	maxReceiveCountWriteToPath := fmt.Sprintf(
		"$[%q].redrivePolicy.maxReceiveCount",
		sourceQueueResourceName,
	)

	// Get the annotation value (default to 10 if not present)
	maxReceiveCountAnnotation := linkhelpers.GetAnnotation(
		input.ResourceAChanges,
		"aws.sqs.redrive.maxReceiveCount",
		core.MappingNodeFromInt(10),
	)

	err = linkhelpers.CollectLinkDataChanges(
		maxReceiveCountWriteToPath,
		currentLinkData,
		changes,
		maxReceiveCountAnnotation,
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
