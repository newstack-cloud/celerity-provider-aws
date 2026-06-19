package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// TableResource returns a resource implementation for an AWS DynamoDB Table.
func TableResource(
	dynamodbServiceFactory pluginutils.ServiceFactory[*aws.Config, dynamodbservice.Service],
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := examples.ReadFile("examples/resources/dynamodb_table_basic.md")
	completeExample, _ := examples.ReadFile("examples/resources/dynamodb_table_complete.md")

	tableActions := &dynamodbTableResourceActions{
		dynamodbServiceFactory:             dynamodbServiceFactory,
		resourceGroupTaggingServiceFactory: resourceGroupTaggingServiceFactory,
		awsConfigStore:                     awsConfigStore,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/dynamodb/table",
		Label:            "Amazon DynamoDB Table",
		PlainTextSummary: "A resource for managing an Amazon DynamoDB table.",
		FormattedDescription: "The resource type used to define a [DynamoDB table](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Introduction.html) " +
			"that is deployed to AWS.",
		Schema:         dynamodbTableResourceSchema(),
		IDField:        "arn",
		TaggingSupport: provider.TaggingSupportFull,
		// A DynamoDB table typically links out to Lambda functions that process events from the DynamoDB stream
		CommonTerminal: false,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
		},
		GetExternalStateFunc: tableActions.GetExternalState,
		CreateFunc:           tableActions.Create,
		UpdateFunc:           tableActions.Update,
		DestroyFunc:          tableActions.Destroy,
		StabilisedFunc:       tableActions.Stabilised,
	}
}

type dynamodbTableResourceActions struct {
	dynamodbServiceFactory             pluginutils.ServiceFactory[*aws.Config, dynamodbservice.Service]
	resourceGroupTaggingServiceFactory pluginutils.ServiceFactory[*aws.Config, resgrouptagservice.Service]
	awsConfigStore                     pluginutils.ServiceConfigStore[*aws.Config]
}

func (a *dynamodbTableResourceActions) getDynamoDBService(
	ctx context.Context,
	providerContext provider.Context,
) (dynamodbservice.Service, error) {
	awsConfig, err := a.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return a.dynamodbServiceFactory(awsConfig, providerContext), nil
}

func (a *dynamodbTableResourceActions) getResourceGroupTaggingService(
	ctx context.Context,
	providerContext provider.Context,
) (resgrouptagservice.Service, error) {
	awsConfig, err := a.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return a.resourceGroupTaggingServiceFactory(awsConfig, providerContext), nil
}
