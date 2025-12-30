package sqs

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	sqsmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/sqs_mock"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// TestCaseConfig holds configuration for creating test cases.
type TestCaseConfig struct {
	QueueName       string
	QueueARN        string
	QueueURL        string
	Tags            map[string]string
	IsIntegration   bool
	ServiceFactory  func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service
	ConfigStore     pluginutils.ServiceConfigStore[*aws.Config]
	ProviderContext provider.Context
}

// CreateBasicQueueSpecData creates the basic queue spec data used in tests.
func CreateBasicQueueSpecData(queueName string, tags map[string]string) *core.MappingNode {
	tagItems := make([]*core.MappingNode, 0, len(tags))
	for key, value := range tags {
		tagItems = append(tagItems, core.MappingNodeFields(
			"key",
			core.MappingNodeFromString(key),
			"value",
			core.MappingNodeFromString(value),
		))
	}

	return core.MappingNodeFields(
		"queueName",
		core.MappingNodeFromString(queueName),
		"tags",
		core.MappingNodeItems(tagItems...),
	)
}

// CreateBasicQueueTestCase creates a basic queue test case with the given configuration.
func CreateBasicQueueTestCase(config TestCaseConfig) plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service] {
	specData := CreateBasicQueueSpecData(config.QueueName, config.Tags)

	// Create field changes for all fields in the spec
	fieldChanges := []provider.FieldChange{
		{FieldPath: "spec.queueName"},
	}

	// Add tag field changes
	for range config.Tags {
		fieldChanges = append(fieldChanges,
			provider.FieldChange{FieldPath: "spec.tags[0].key"},
			provider.FieldChange{FieldPath: "spec.tags[0].value"},
		)
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service]{
		Name:           "create basic queue",
		ServiceFactory: config.ServiceFactory,
		ConfigStore:    config.ConfigStore,
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-queue-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-queue-id",
					ResourceName: "TestQueue",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/sqs/queue",
						},
						Spec: specData,
					},
				},
				NewFields: fieldChanges,
			},
			ProviderContext: config.ProviderContext,
		},
	}
}

// CreateUnitTestCaseConfig creates configuration for unit tests.
func CreateUnitTestCaseConfig(
	queueName, queueARN, queueURL string,
	tags map[string]string,
	providerContext provider.Context,
) TestCaseConfig {
	service := sqsmock.CreateSQSServiceMock(
		sqsmock.WithCreateQueueOutput(
			&sqs.CreateQueueOutput{
				QueueUrl: aws.String(queueURL),
			},
		),
		sqsmock.WithGetQueueAttributesOutput(
			&sqs.GetQueueAttributesOutput{
				Attributes: map[string]string{
					"QueueArn": queueARN,
				},
			},
		),
	)

	return TestCaseConfig{
		QueueName: queueName,
		QueueARN:  queueARN,
		QueueURL:  queueURL,
		Tags:      tags,
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			&testutils.MockAWSConfigLoader{},
			utils.AWSConfigCacheKey,
		),
		ProviderContext: providerContext,
	}
}

// CreateUnitTestCaseWithService creates a unit test case with the service for mock calls.
func CreateUnitTestCaseWithService(queueName, queueARN, queueURL string, tags map[string]string) (plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service], sqsservice.Service) {
	service := sqsmock.CreateSQSServiceMock(
		sqsmock.WithCreateQueueOutput(
			&sqs.CreateQueueOutput{
				QueueUrl: aws.String(queueURL),
			},
		),
		sqsmock.WithGetQueueAttributesOutput(
			&sqs.GetQueueAttributesOutput{
				Attributes: map[string]string{
					"QueueArn": queueARN,
				},
			},
		),
	)

	config := TestCaseConfig{
		QueueName: queueName,
		QueueARN:  queueARN,
		QueueURL:  queueURL,
		Tags:      tags,
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			&testutils.MockAWSConfigLoader{},
			utils.AWSConfigCacheKey,
		),
	}

	testCase := CreateBasicQueueTestCase(config)
	return testCase, service
}

// CreateIntegrationTestCaseConfig creates configuration for integration tests.
func CreateIntegrationTestCaseConfig(queueName string, tags map[string]string, loader utils.AWSConfigLoader, providerContext provider.Context) TestCaseConfig {
	return TestCaseConfig{
		QueueName:      queueName,
		Tags:           tags,
		ServiceFactory: sqsservice.NewService,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		ProviderContext: providerContext,
	}
}

// CreateExpectedCreateQueueInput creates the expected CreateQueue input for unit tests.
func CreateExpectedCreateQueueInput(queueName string, tags map[string]string) *sqs.CreateQueueInput {
	return &sqs.CreateQueueInput{
		QueueName:  aws.String(queueName),
		Attributes: map[string]string{},
		Tags:       tags,
	}
}

// CreateExpectedGetQueueAttributesInput creates the expected GetQueueAttributes input.
func CreateExpectedGetQueueAttributesInput(queueURL string) *sqs.GetQueueAttributesInput {
	return &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(queueURL),
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameQueueArn,
		},
	}
}
