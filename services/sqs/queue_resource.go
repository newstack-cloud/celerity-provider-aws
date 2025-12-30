package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// QueueResource returns a resource implementation for an AWS SQS Queue.
func QueueResource(
	sqsServiceFactory pluginutils.ServiceFactory[*aws.Config, sqsservice.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := examples.ReadFile("examples/resources/sqs_queue_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/sqs_queue_complete.md")
	jsoncExample, _ := examples.ReadFile("examples/resources/sqs_queue_jsonc.md")
	fifoExample, _ := examples.ReadFile("examples/resources/sqs_queue_fifo.md")

	sqsQueueActions := &sqsQueueResourceActions{
		sqsServiceFactory:                  sqsServiceFactory,
		resourceGroupTaggingServiceFactory: resourceGroupTaggingServiceFactory,
		awsConfigStore:                     awsConfigStore,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/sqs/queue",
		Label:            "Amazon SQS Queue",
		PlainTextSummary: "A resource for managing an Amazon SQS queue.",
		FormattedDescription: "The resource type used to define an [SQS queue](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-getting-started.html) " +
			"that is deployed to AWS.",
		Schema:         sqsQueueResourceSchema(),
		IDField:        "arn",
		TaggingSupport: provider.TaggingSupportFull,
		// An SQS queue is commonly used as an event source for Lambda functions and other services
		CommonTerminal: false,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
			string(jsoncExample),
			string(fifoExample),
		},
		ResourceCanLinkTo:    []string{"aws/sqs/queue", "aws/lambda/function", "aws/lambda/version", "aws/lambda/alias"},
		GetExternalStateFunc: sqsQueueActions.GetExternalState,
		CreateFunc:           sqsQueueActions.Create,
		UpdateFunc:           sqsQueueActions.Update,
		DestroyFunc:          sqsQueueActions.Destroy,
		StabilisedFunc:       sqsQueueActions.Stabilised,
	}
}

type sqsQueueResourceActions struct {
	sqsServiceFactory                  pluginutils.ServiceFactory[*aws.Config, sqsservice.Service]
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service]
	awsConfigStore                     pluginutils.ServiceConfigStore[*aws.Config]
}

func (s *sqsQueueResourceActions) getSQSService(
	ctx context.Context,
	providerContext provider.Context,
) (sqsservice.Service, error) {
	awsConfig, err := s.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return s.sqsServiceFactory(awsConfig, providerContext), nil
}

func (s *sqsQueueResourceActions) getResourceGroupTaggingService(
	ctx context.Context,
	providerContext provider.Context,
) (resgrouptagservice.Service, error) {
	awsConfig, err := s.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return s.resourceGroupTaggingServiceFactory(awsConfig, providerContext), nil
}
