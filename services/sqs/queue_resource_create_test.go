//go:build unit

package sqs

import (
	"encoding/json"
	"fmt"
	"testing"

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
	"github.com/stretchr/testify/suite"
)

type SQSQueueResourceCreateSuite struct {
	suite.Suite
}

func (s *SQSQueueResourceCreateSuite) Test_create_sqs_queue() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service]{
		createBasicQueueTestCase(providerCtx, loader),
		createAdvancedQueueTestCase(providerCtx, loader),
		createFIFOQueueTestCase(providerCtx, loader),
		createSQSFailureTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	queueResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, sqsservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return QueueResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		queueResourceWrapper,
		&s.Suite,
	)
}

func createBasicQueueTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service] {
	queueARN := "arn:aws:sqs:us-west-2:123456789012:test-queue"
	queueURL := "https://sqs.us-west-2.amazonaws.com/123456789012/test-queue"
	tags := map[string]string{
		"Environment": "test",
		"Project":     "test-project",
	}

	// Create the mock service directly
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

	// Use shared helper for the basic test case structure
	config := CreateUnitTestCaseConfig("test-queue", queueARN, queueURL, tags, providerCtx)
	testCase := CreateBasicQueueTestCase(config)

	// Override with unit test specific configurations
	testCase.ServiceFactory = func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
		return service
	}
	testCase.ServiceMockCalls = &service.MockCalls
	testCase.ExpectedOutput = &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.arn":      core.MappingNodeFromString(queueARN),
			"spec.queueUrl": core.MappingNodeFromString(queueURL),
		},
	}
	testCase.SaveActionsCalled = map[string]any{
		"CreateQueue":        CreateExpectedCreateQueueInput("test-queue", tags),
		"GetQueueAttributes": CreateExpectedGetQueueAttributesInput(queueURL),
	}
	testCase.ConfigStore = utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	return testCase
}

func createAdvancedQueueTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service] {
	queueARN := "arn:aws:sqs:us-west-2:123456789012:test-advanced-queue"
	queueURL := "https://sqs.us-west-2.amazonaws.com/123456789012/test-advanced-queue"

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

	specData := core.MappingNodeFields(
		"queueName",
		core.MappingNodeFromString("test-advanced-queue"),
		"delaySeconds",
		core.MappingNodeFromInt(30),
		"visibilityTimeout",
		core.MappingNodeFromInt(30),
		"maximumMessageSize",
		core.MappingNodeFromInt(262144),
		"messageRetentionPeriod",
		core.MappingNodeFromInt(345600),
		"receiveMessageWaitTimeSeconds",
		core.MappingNodeFromInt(20),
		"redrivePolicy",
		core.MappingNodeFields(
			"deadLetterTargetArn",
			core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:test-dead-letter-queue"),
			"maxReceiveCount",
			core.MappingNodeFromInt(10),
		),
		"redriveAllowPolicy",
		core.MappingNodeFields(
			"redrivePermission",
			core.MappingNodeFromString("allowAll"),
			"sourceQueueArns",
			core.MappingNodeFromStringSlice([]string{"arn:aws:sqs:us-west-2:123456789012:test-source-queue"}),
		),
		"kmsMasterKeyId",
		core.MappingNodeFromString("arn:aws:kms:us-west-2:123456789012:key/test-key"),
		"kmsDataKeyReusePeriodSeconds",
		core.MappingNodeFromInt(300),
		"sqsManagedSseEnabled",
		core.MappingNodeFromBool(true),
		"policy",
		core.MappingNodeFields(
			"Version",
			core.MappingNodeFromString("2012-10-17"),
			"Statement",
			core.MappingNodeItems(
				core.MappingNodeFields(
					"Effect",
					core.MappingNodeFromString("Allow"),
					"Principal",
					core.MappingNodeFields(
						"AWS",
						core.MappingNodeFromString("arn:aws:iam::123456789012:user/test-user"),
					),
					"Action",
					core.MappingNodeFromStringSlice([]string{"sqs:SendMessage", "sqs:ReceiveMessage"}),
					"Resource",
					core.MappingNodeFromString("arn:aws:sqs:*:*:test-queue"),
				),
			),
		),
		"tags",
		core.MappingNodeItems(
			core.MappingNodeFields(
				"key",
				core.MappingNodeFromString("Environment"),
			),
			core.MappingNodeFields(
				"key",
				core.MappingNodeFromString("Project"),
				"value",
				core.MappingNodeFromString("test-project"),
			),
			core.MappingNodeFields(
				"key",
				core.MappingNodeFromString("Environment"),
				"value",
				core.MappingNodeFromString("test"),
			),
		),
	)

	return plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service]{
		Name: "create queue with advanced configuration",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-advanced-queue-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-advanced-queue-id",
					ResourceName: "TestAdvancedQueue",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/sqs/queue",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.queueName",
					},
					{
						FieldPath: "spec.delaySeconds",
					},
					{
						FieldPath: "spec.visibilityTimeout",
					},
					{
						FieldPath: "spec.maximumMessageSize",
					},
					{
						FieldPath: "spec.messageRetentionPeriod",
					},
					{
						FieldPath: "spec.receiveMessageWaitTimeSeconds",
					},
					{
						FieldPath: "spec.redrivePolicy.deadLetterTargetArn",
					},
					{
						FieldPath: "spec.redrivePolicy.maxReceiveCount",
					},
					{
						FieldPath: "spec.redriveAllowPolicy.redrivePermission",
					},
					{
						FieldPath: "spec.redriveAllowPolicy.sourceQueueArns",
					},
					{
						FieldPath: "spec.kmsMasterKeyId",
					},
					{
						FieldPath: "spec.kmsDataKeyReusePeriodSeconds",
					},
					{
						FieldPath: "spec.sqsManagedSseEnabled",
					},
					{
						FieldPath: "spec.policy.Version",
					},
					{
						FieldPath: "spec.policy.Statement[0].Effect",
					},
					{
						FieldPath: "spec.policy.Statement[0].Principal.AWS",
					},
					{
						FieldPath: "spec.policy.Statement[0].Action",
					},
					{
						FieldPath: "spec.policy.Statement[0].Resource",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":      core.MappingNodeFromString(queueARN),
				"spec.queueUrl": core.MappingNodeFromString(queueURL),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateQueue": expectCreateQueueAdvancedConfigInput,
			"GetQueueAttributes": &sqs.GetQueueAttributesInput{
				QueueUrl: aws.String(queueURL),
				AttributeNames: []types.QueueAttributeName{
					types.QueueAttributeNameQueueArn,
				},
			},
		},
	}
}

func expectCreateQueueAdvancedConfigInput(actual any) (plugintestutils.EqualityCheckValues, error) {
	input, ok := actual.(*sqs.CreateQueueInput)
	if !ok {
		return plugintestutils.EqualityCheckValues{}, fmt.Errorf("input is not an *sqs.CreateQueueInput")
	}

	// Unmarshal the queue policy document for comparison
	var actualDoc map[string]any
	if err := json.Unmarshal([]byte(input.Attributes["Policy"]), &actualDoc); err != nil {
		return plugintestutils.EqualityCheckValues{}, err
	}
	expectedDoc := map[string]any{
		"Version": "2012-10-17",
		"Statement": []any{
			map[string]any{
				"Principal": map[string]any{
					"AWS": "arn:aws:iam::123456789012:user/test-user",
				},
				"Effect":   "Allow",
				"Action":   []any{"sqs:SendMessage", "sqs:ReceiveMessage"},
				"Resource": "arn:aws:sqs:*:*:test-queue",
			},
		},
	}

	// Unmarshal the redrive policy document for comparison
	var actualRedriveDoc map[string]any
	if err := json.Unmarshal([]byte(input.Attributes["RedrivePolicy"]), &actualRedriveDoc); err != nil {
		return plugintestutils.EqualityCheckValues{}, err
	}
	expectedRedriveDoc := map[string]any{
		"deadLetterTargetArn": "arn:aws:sqs:us-west-2:123456789012:test-dead-letter-queue",
		"maxReceiveCount":     10.0,
	}

	// Unmarshal the redrive allow policy document for comparison
	var actualRedriveAllowDoc map[string]any
	if err := json.Unmarshal([]byte(input.Attributes["RedriveAllowPolicy"]), &actualRedriveAllowDoc); err != nil {
		return plugintestutils.EqualityCheckValues{}, err
	}
	expectedRedriveAllowDoc := map[string]any{
		"redrivePermission": "allowAll",
		"sourceQueueArns":   []any{"arn:aws:sqs:us-west-2:123456789012:test-source-queue"},
	}

	expectedInput := &sqs.CreateQueueInput{
		QueueName: aws.String("test-advanced-queue"),
		Tags: map[string]string{
			"Environment": "test",
			"Project":     "test-project",
		},
	}
	actualInput := &sqs.CreateQueueInput{
		QueueName: input.QueueName,
		Tags:      input.Tags,
	}
	expectedMap := map[string]any{
		"QueueName":          *expectedInput.QueueName,
		"Tags":               expectedInput.Tags,
		"PolicyDocument":     expectedDoc,
		"RedrivePolicy":      expectedRedriveDoc,
		"RedriveAllowPolicy": expectedRedriveAllowDoc,
	}
	actualMap := map[string]any{
		"QueueName":          *actualInput.QueueName,
		"Tags":               actualInput.Tags,
		"PolicyDocument":     actualDoc,
		"RedrivePolicy":      actualRedriveDoc,
		"RedriveAllowPolicy": actualRedriveAllowDoc,
	}
	return plugintestutils.EqualityCheckValues{
		Expected: expectedMap,
		Actual:   actualMap,
	}, nil
}

func createFIFOQueueTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service] {
	queueARN := "arn:aws:sqs:us-west-2:123456789012:test-queue.fifo"
	queueURL := "https://sqs.us-west-2.amazonaws.com/123456789012/test-queue.fifo"

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

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"queueName": core.MappingNodeFromString("test-queue.fifo"),
			"tags": core.MappingNodeItems(
				core.MappingNodeFields(
					"key",
					core.MappingNodeFromString("Environment"),
					"value",
					core.MappingNodeFromString("test"),
				),
				core.MappingNodeFields(
					"key",
					core.MappingNodeFromString("Project"),
					"value",
					core.MappingNodeFromString("test-project"),
				),
			),
			"fifoQueue":                 core.MappingNodeFromBool(true),
			"fifoThroughput":            core.MappingNodeFromInt(100),
			"contentBasedDeduplication": core.MappingNodeFromBool(true),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service]{
		Name: "create FIFO queue",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-fifo-queue-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-fifo-queue-id",
					ResourceName: "TestFIFOQueue",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/sqs/queue",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.queueName",
					},
					{
						FieldPath: "spec.fifoQueue",
					},
					{
						FieldPath: "spec.fifoThroughput",
					},
					{
						FieldPath: "spec.contentBasedDeduplication",
					},
					{
						FieldPath: "spec.tags",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":      core.MappingNodeFromString(queueARN),
				"spec.queueUrl": core.MappingNodeFromString(queueURL),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateQueue": &sqs.CreateQueueInput{
				QueueName: aws.String("test-queue.fifo"),
				Attributes: map[string]string{
					"FifoQueue":                 "true",
					"ContentBasedDeduplication": "true",
				},
				Tags: map[string]string{
					"Environment": "test",
					"Project":     "test-project",
				},
			},
			"GetQueueAttributes": &sqs.GetQueueAttributesInput{
				QueueUrl: aws.String(queueURL),
				AttributeNames: []types.QueueAttributeName{
					types.QueueAttributeNameQueueArn,
				},
			},
		},
	}
}

func createSQSFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service] {
	service := sqsmock.CreateSQSServiceMock(
		sqsmock.WithCreateQueueError(fmt.Errorf("failed to create queue")),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"queueName": core.MappingNodeFromString("test-queue"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, sqsservice.Service]{
		Name: "create SQS queue failure",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
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
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"CreateQueue": &sqs.CreateQueueInput{
				QueueName: aws.String("test-queue"),
			},
		},
	}
}

func TestSQSQueueResourceCreateSuite(t *testing.T) {
	suite.Run(t, new(SQSQueueResourceCreateSuite))
}
