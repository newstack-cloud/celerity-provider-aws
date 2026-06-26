//go:build unit

package cloudcontrol

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	resgrouptagtypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	resgrouptagservice "github.com/newstack-cloud/bluelink-provider-aws/services/resgrouptag/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// A small SQS-Queue-shaped CCResourceConfig used to exercise
// the generic engine without depending on codegen output.
func testQueueConfig() CCResourceConfig {
	return CCResourceConfig{
		BlueprintType:    "aws/sqs/queue",
		CFNType:          "AWS::SQS::Queue",
		Label:            "Amazon SQS Queue",
		PlainTextSummary: "A test SQS queue backed by Cloud Control.",
		Schema: &provider.ResourceDefinitionsSchema{
			Type:  provider.ResourceDefinitionsSchemaTypeObject,
			Label: "AWS SQS Queue",
			Attributes: map[string]*provider.ResourceDefinitionsSchema{
				"queueName": {
					Type:         provider.ResourceDefinitionsSchemaTypeString,
					Label:        "Queue Name",
					MustRecreate: true,
					Nullable:     true,
				},
				"visibilityTimeout": {
					Type:     provider.ResourceDefinitionsSchemaTypeInteger,
					Label:    "Visibility Timeout",
					Nullable: true,
				},
				"tags": {
					Type:     provider.ResourceDefinitionsSchemaTypeArray,
					Label:    "Tags",
					Nullable: true,
					Items: &provider.ResourceDefinitionsSchema{
						Type:  provider.ResourceDefinitionsSchemaTypeObject,
						Label: "Tag",
						Attributes: map[string]*provider.ResourceDefinitionsSchema{
							"key":   {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Key"},
							"value": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Value"},
						},
					},
				},
				"queueUrl": {
					Type:     provider.ResourceDefinitionsSchemaTypeString,
					Label:    "Queue URL",
					Computed: true,
				},
				"arn": {
					Type:     provider.ResourceDefinitionsSchemaTypeString,
					Label:    "ARN",
					Computed: true,
				},
			},
		},
		Meta: CCResourceMeta{
			PrimaryIdentifierField: "queueUrl",
			ComputedFields:         []string{"queueUrl", "arn"},
			CreateOnlyFields:       []string{"queueName"},
			TagPropertyName:        "tags",
			TagShape:               TagShapeKeyValueList,
		},
	}
}

// Wraps CCResource for the deploy/stabilised test harnesses, which
// only supply the primary service factory and config store.
func newTestResource(
	serviceFactory pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service],
	configStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	return CCResource(
		testQueueConfig(),
		serviceFactory,
		mockResourceGroupTaggingServiceFactory,
		configStore,
	)
}

type mockResourceGroupTaggingService struct {
	output *resourcegroupstaggingapi.GetResourcesOutput
}

func (m *mockResourceGroupTaggingService) GetResources(
	ctx context.Context,
	input *resourcegroupstaggingapi.GetResourcesInput,
	optFns ...func(*resourcegroupstaggingapi.Options),
) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
	if m.output != nil {
		return m.output, nil
	}
	return &resourcegroupstaggingapi.GetResourcesOutput{
		ResourceTagMappingList: []resgrouptagtypes.ResourceTagMapping{},
	}, nil
}

func mockResourceGroupTaggingServiceFactory(
	config *aws.Config,
	ctx provider.Context,
) resgrouptagservice.Service {
	return &mockResourceGroupTaggingService{}
}

func newAWSConfigStore(loader *testutils.MockAWSConfigLoader) pluginutils.ServiceConfigStore[*aws.Config] {
	return utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)
}
