package sqs

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (s *sqsQueueResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	sqsService, err := s.getSQSService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[sqsservice.Service]{
		&queueCreate{},
	}

	saveOpCtx := pluginutils.SaveOperationContext{
		Data: map[string]any{
			"ResourceDeployInput": input,
		},
	}

	time.Sleep(60 * time.Second)

	hasSavedValues, saveOpCtx, err := pluginutils.RunSaveOperations(
		ctx,
		saveOpCtx,
		createOperations,
		input,
		sqsService,
	)
	if err != nil {
		return nil, err
	}
	if !hasSavedValues {
		return nil, fmt.Errorf("no values were saved during queue creation")
	}

	createQueueOutput, ok := saveOpCtx.Data["createQueueOutput"].(*sqs.CreateQueueOutput)
	if !ok {
		return nil, fmt.Errorf("createQueueOutput not found in save operation context")
	}

	computedFields := map[string]*core.MappingNode{
		"spec.arn":      core.MappingNodeFromString(saveOpCtx.ProviderUpstreamID),
		"spec.queueUrl": core.MappingNodeFromString(aws.ToString(createQueueOutput.QueueUrl)),
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}
