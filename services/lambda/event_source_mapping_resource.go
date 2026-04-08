package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// EventSourceMappingResource returns a resource implementation for an AWS Lambda Event Source Mapping.
func EventSourceMappingResource(
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := examples.ReadFile("examples/resources/lambda_event_source_mapping_basic.md")
	kinesisExample, _ := examples.ReadFile("examples/resources/lambda_event_source_mapping_kinesis.md")
	dynamodbExample, _ := examples.ReadFile("examples/resources/lambda_event_source_mapping_dynamodb.md")
	kafkaExample, _ := examples.ReadFile("examples/resources/lambda_event_source_mapping_kafka.md")
	completeExample, _ := examples.ReadFile("examples/resources/lambda_event_source_mapping_complete.md")
	jsoncExample, _ := examples.ReadFile("examples/resources/lambda_event_source_mapping_jsonc.md")
	documentdbExample, _ := examples.ReadFile("examples/resources/lambda_event_source_mapping_documentdb.md")
	mqExample, _ := examples.ReadFile("examples/resources/lambda_event_source_mapping_mq.md")

	lambdaEventSourceMappingActions := &lambdaEventSourceMappingResourceActions{
		lambdaServiceFactory:               lambdaServiceFactory,
		resourceGroupTaggingServiceFactory: resourceGroupTaggingServiceFactory,
		awsConfigStore:                     awsConfigStore,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/lambda/eventSourceMapping",
		Label:            "AWS Lambda Event Source Mapping",
		PlainTextSummary: "A resource for managing an AWS Lambda event source mapping.",
		FormattedDescription: "The resource type used to define a [Lambda event source mapping](https://docs.aws.amazon.com/lambda/latest/dg/invocation-eventsourcemapping.html) " +
			"that is deployed to AWS.",
		Schema:         lambdaEventSourceMappingResourceSchema(),
		IDField:        "id",
		TaggingSupport: provider.TaggingSupportFull,
		CommonTerminal: true,
		FormattedExamples: []string{
			string(basicExample),
			string(kinesisExample),
			string(dynamodbExample),
			string(kafkaExample),
			string(completeExample),
			string(jsoncExample),
			string(documentdbExample),
			string(mqExample),
		},
		GetExternalStateFunc: lambdaEventSourceMappingActions.GetExternalState,
		CreateFunc:           lambdaEventSourceMappingActions.Create,
		UpdateFunc:           lambdaEventSourceMappingActions.Update,
		DestroyFunc:          lambdaEventSourceMappingActions.Destroy,
		StabilisedFunc:       lambdaEventSourceMappingActions.Stabilised,
	}
}

type lambdaEventSourceMappingResourceActions struct {
	lambdaServiceFactory               pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service]
	awsConfigStore                     pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaEventSourceMappingResourceActions) getLambdaService(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.lambdaServiceFactory(awsConfig, providerContext), nil
}

func (l *lambdaEventSourceMappingResourceActions) getResourceGroupTaggingService(
	ctx context.Context,
	providerContext provider.Context,
) (resgrouptagservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.resourceGroupTaggingServiceFactory(awsConfig, providerContext), nil
}
