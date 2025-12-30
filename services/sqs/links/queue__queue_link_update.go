package sqslinks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *sqsQueueQueueLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	sqsService, err := l.getSQSService(
		ctx,
		provider.NewProviderContextFromLinkContext(
			input.LinkContext,
			"aws",
		),
	)
	if err != nil {
		return nil, err
	}

	sourceQueueSpec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(
		input.ResourceInfo,
	)
	sourceQueueURL, hasSourceQueueURL := pluginutils.GetValueByPath(
		"$.queueUrl",
		sourceQueueSpec,
	)
	if !hasSourceQueueURL {
		return nil, fmt.Errorf(
			"queue URL could not be retrieved from source queue %q",
			pluginutils.GetResourceName(input.ResourceInfo),
		)
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeRedrivePolicyFromQueue(
			ctx,
			input,
			core.StringValue(sourceQueueURL),
			sqsService,
		)
	}

	return l.addRedrivePolicyToQueue(
		ctx,
		input,
		core.StringValue(sourceQueueURL),
		sqsService,
	)
}

func (l *sqsQueueQueueLinkActions) addRedrivePolicyToQueue(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	sourceQueueURL string,
	sqsService sqsservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	deadLetterQueueSpec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(
		input.OtherResourceInfo,
	)
	deadLetterQueueARN, hasDeadLetterQueueARN := pluginutils.GetValueByPath(
		"$.arn",
		deadLetterQueueSpec,
	)
	if !hasDeadLetterQueueARN {
		return nil, fmt.Errorf(
			"queue ARN could not be retrieved from dead-letter queue %q",
			pluginutils.GetResourceName(input.OtherResourceInfo),
		)
	}

	// Get maxReceiveCount from annotations
	maxReceiveCount, _ := pluginutils.GetIntAnnotation(
		input.ResourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.sqs.redrive.maxReceiveCount",
			Default: 10,
		},
	)

	// Construct the redrive policy
	redrivePolicy := map[string]interface{}{
		"deadLetterTargetArn": core.StringValue(deadLetterQueueARN),
		"maxReceiveCount":     maxReceiveCount,
	}

	redrivePolicyJSON, err := json.Marshal(redrivePolicy)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal redrive policy: %w", err)
	}

	// Set the queue attribute with the redrive policy
	_, err = sqsService.SetQueueAttributes(
		ctx,
		&sqs.SetQueueAttributesInput{
			QueueUrl: aws.String(sourceQueueURL),
			Attributes: map[string]string{
				"RedrivePolicy": string(redrivePolicyJSON),
			},
		},
	)
	if err != nil {
		return nil, err
	}

	resourceFieldPath := fmt.Sprintf(
		"%s::spec.redrivePolicy",
		input.ResourceInfo.ResourceName,
	)
	linkFieldPath := fmt.Sprintf(
		"%s.redrivePolicy",
		input.ResourceInfo.ResourceName,
	)

	redrivePolicyNode := core.MappingNodeFields(
		"deadLetterTargetArn",
		core.MappingNodeFromString(core.StringValue(deadLetterQueueARN)),
		"maxReceiveCount",
		core.MappingNodeFromInt(maxReceiveCount),
	)

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(
			input.ResourceInfo.ResourceName,
			core.MappingNodeFields(
				"redrivePolicy",
				redrivePolicyNode,
			),
		),
		ResourceDataMappings: map[string]string{
			resourceFieldPath: linkFieldPath,
		},
	}, nil
}

func (l *sqsQueueQueueLinkActions) removeRedrivePolicyFromQueue(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	sourceQueueURL string,
	sqsService sqsservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	// Remove the redrive policy by setting it to an empty string
	_, err := sqsService.SetQueueAttributes(
		ctx,
		&sqs.SetQueueAttributesInput{
			QueueUrl: aws.String(sourceQueueURL),
			Attributes: map[string]string{
				"RedrivePolicy": "",
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *sqsQueueQueueLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The dead-letter queue is not updated as a part of the link update.
	return &provider.LinkUpdateResourceOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *sqsQueueQueueLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// There are no intermediary resources to update
	// for the SQS queue to queue link.
	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData:                   core.MappingNodeFields(),
	}, nil
}
