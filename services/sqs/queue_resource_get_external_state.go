package sqs

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (s *sqsQueueResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	// TODO: Implement SQS queue external state retrieval logic
	return &provider.ResourceGetExternalStateOutput{}, nil
}
