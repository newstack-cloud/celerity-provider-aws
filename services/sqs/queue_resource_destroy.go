package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (s *sqsQueueResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	sqsService, err := s.getSQSService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	queueURL := core.StringValue(
		input.ResourceState.SpecData.Fields["queueUrl"],
	)
	_, err = sqsService.DeleteQueue(
		ctx,
		&sqs.DeleteQueueInput{
			QueueUrl: &queueURL,
		},
	)

	return err
}
