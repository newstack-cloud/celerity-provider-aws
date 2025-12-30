//go:build unit

package lambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type LambdaEventSourceMappingResourceCreateSuite struct {
	suite.Suite
}

func (s *LambdaEventSourceMappingResourceCreateSuite) Test_create_lambda_event_source_mapping() {
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

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		createBasicEventSourceMappingTestCase(providerCtx, loader),
		createKinesisEventSourceMappingTestCase(providerCtx, loader),
		createKafkaEventSourceMappingTestCase(providerCtx, loader),
		createEventSourceMappingWithDestinationConfigTestCase(providerCtx, loader),
		createEventSourceMappingWithTagsTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	eventSourceMappingResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return EventSourceMappingResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		eventSourceMappingResourceWrapper,
		&s.Suite,
	)
}

func createBasicEventSourceMappingTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	uuid := "test-uuid-123"
	eventSourceMappingArn := "arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-uuid-123"
	functionArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateEventSourceMappingOutput(&lambda.CreateEventSourceMappingOutput{
			UUID:                  aws.String(uuid),
			EventSourceMappingArn: aws.String(eventSourceMappingArn),
			FunctionArn:           aws.String(functionArn),
			State:                 aws.String("Creating"),
			EventSourceArn:        aws.String("arn:aws:sqs:us-west-2:123456789012:test-queue"),
			BatchSize:             aws.Int32(10),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":   core.MappingNodeFromString("test-function"),
			"eventSourceArn": core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:test-queue"),
			"batchSize":      core.MappingNodeFromInt(10),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create basic SQS event source mapping",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
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
			ResourceID: "test-esm-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-esm-id",
					ResourceName: "TestEventSourceMapping",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventSourceMapping",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.eventSourceArn",
					},
					{
						FieldPath: "spec.batchSize",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.id":                    core.MappingNodeFromString(uuid),
				"spec.eventSourceMappingArn": core.MappingNodeFromString(eventSourceMappingArn),
				"spec.functionArn":           core.MappingNodeFromString(functionArn),
				"spec.state":                 core.MappingNodeFromString("Creating"),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateEventSourceMapping": &lambda.CreateEventSourceMappingInput{
				FunctionName:   aws.String("test-function"),
				EventSourceArn: aws.String("arn:aws:sqs:us-west-2:123456789012:test-queue"),
				BatchSize:      aws.Int32(10),
			},
		},
	}
}

func createKinesisEventSourceMappingTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	uuid := "test-kinesis-uuid"
	eventSourceMappingArn := "arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-kinesis-uuid"
	functionArn := "arn:aws:lambda:us-west-2:123456789012:function:test-kinesis-function"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateEventSourceMappingOutput(&lambda.CreateEventSourceMappingOutput{
			UUID:                           aws.String(uuid),
			EventSourceMappingArn:          aws.String(eventSourceMappingArn),
			FunctionArn:                    aws.String(functionArn),
			State:                          aws.String("Creating"),
			EventSourceArn:                 aws.String("arn:aws:kinesis:us-west-2:123456789012:stream/test-stream"),
			BatchSize:                      aws.Int32(100),
			StartingPosition:               types.EventSourcePositionLatest,
			MaximumBatchingWindowInSeconds: aws.Int32(5),
			MaximumRecordAgeInSeconds:      aws.Int32(604800),
			MaximumRetryAttempts:           aws.Int32(3),
			BisectBatchOnFunctionError:     aws.Bool(true),
			ParallelizationFactor:          aws.Int32(2),
			FunctionResponseTypes:          []types.FunctionResponseType{types.FunctionResponseTypeReportBatchItemFailures},
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":                   core.MappingNodeFromString("test-kinesis-function"),
			"eventSourceArn":                 core.MappingNodeFromString("arn:aws:kinesis:us-west-2:123456789012:stream/test-stream"),
			"batchSize":                      core.MappingNodeFromInt(100),
			"startingPosition":               core.MappingNodeFromString("LATEST"),
			"maximumBatchingWindowInSeconds": core.MappingNodeFromInt(5),
			"maximumRecordAgeInSeconds":      core.MappingNodeFromInt(604800),
			"maximumRetryAttempts":           core.MappingNodeFromInt(3),
			"bisectBatchOnFunctionError":     core.MappingNodeFromBool(true),
			"parallelizationFactor":          core.MappingNodeFromInt(2),
			"functionResponseTypes": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("ReportBatchItemFailures"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create kinesis event source mapping with maximum configuration",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
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
			ResourceID: "test-kinesis-esm-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-kinesis-esm-id",
					ResourceName: "TestKinesisEventSourceMapping",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventSourceMapping",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.functionName"},
					{FieldPath: "spec.eventSourceArn"},
					{FieldPath: "spec.batchSize"},
					{FieldPath: "spec.startingPosition"},
					{FieldPath: "spec.maximumBatchingWindowInSeconds"},
					{FieldPath: "spec.maximumRecordAgeInSeconds"},
					{FieldPath: "spec.maximumRetryAttempts"},
					{FieldPath: "spec.bisectBatchOnFunctionError"},
					{FieldPath: "spec.parallelizationFactor"},
					{FieldPath: "spec.functionResponseTypes"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.id":                    core.MappingNodeFromString(uuid),
				"spec.eventSourceMappingArn": core.MappingNodeFromString(eventSourceMappingArn),
				"spec.functionArn":           core.MappingNodeFromString(functionArn),
				"spec.state":                 core.MappingNodeFromString("Creating"),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateEventSourceMapping": &lambda.CreateEventSourceMappingInput{
				FunctionName:                   aws.String("test-kinesis-function"),
				EventSourceArn:                 aws.String("arn:aws:kinesis:us-west-2:123456789012:stream/test-stream"),
				BatchSize:                      aws.Int32(100),
				StartingPosition:               types.EventSourcePositionLatest,
				MaximumBatchingWindowInSeconds: aws.Int32(5),
				MaximumRecordAgeInSeconds:      aws.Int32(604800),
				MaximumRetryAttempts:           aws.Int32(3),
				BisectBatchOnFunctionError:     aws.Bool(true),
				ParallelizationFactor:          aws.Int32(2),
				FunctionResponseTypes:          []types.FunctionResponseType{types.FunctionResponseTypeReportBatchItemFailures},
			},
		},
	}
}

func createKafkaEventSourceMappingTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	uuid := "test-kafka-uuid"
	eventSourceMappingArn := "arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-kafka-uuid"
	functionArn := "arn:aws:lambda:us-west-2:123456789012:function:test-kafka-function"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateEventSourceMappingOutput(&lambda.CreateEventSourceMappingOutput{
			UUID:                  aws.String(uuid),
			EventSourceMappingArn: aws.String(eventSourceMappingArn),
			FunctionArn:           aws.String(functionArn),
			State:                 aws.String("Creating"),
			BatchSize:             aws.Int32(50),
			Topics:                []string{"test-topic-1", "test-topic-2"},
			FilterCriteria: &types.FilterCriteria{
				Filters: []types.Filter{
					{
						Pattern: aws.String(`{"eventType": ["order"]}`),
					},
				},
			},
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-kafka-function"),
			"batchSize":    core.MappingNodeFromInt(50),
			"topics": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("test-topic-1"),
					core.MappingNodeFromString("test-topic-2"),
				},
			},
			"filterCriteria": {
				Fields: map[string]*core.MappingNode{
					"filters": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"pattern": core.MappingNodeFromString(`{"eventType": ["order"]}`),
								},
							},
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create kafka event source mapping with filter criteria",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
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
			ResourceID: "test-kafka-esm-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-kafka-esm-id",
					ResourceName: "TestKafkaEventSourceMapping",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventSourceMapping",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.functionName"},
					{FieldPath: "spec.batchSize"},
					{FieldPath: "spec.topics"},
					{FieldPath: "spec.filterCriteria"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.id":                    core.MappingNodeFromString(uuid),
				"spec.eventSourceMappingArn": core.MappingNodeFromString(eventSourceMappingArn),
				"spec.functionArn":           core.MappingNodeFromString(functionArn),
				"spec.state":                 core.MappingNodeFromString("Creating"),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateEventSourceMapping": &lambda.CreateEventSourceMappingInput{
				FunctionName: aws.String("test-kafka-function"),
				BatchSize:    aws.Int32(50),
				Topics:       []string{"test-topic-1", "test-topic-2"},
				FilterCriteria: &types.FilterCriteria{
					Filters: []types.Filter{
						{
							Pattern: aws.String(`{"eventType": ["order"]}`),
						},
					},
				},
			},
		},
	}
}

func createEventSourceMappingWithDestinationConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	uuid := "test-dest-uuid"
	eventSourceMappingArn := "arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-dest-uuid"
	functionArn := "arn:aws:lambda:us-west-2:123456789012:function:test-dest-function"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateEventSourceMappingOutput(&lambda.CreateEventSourceMappingOutput{
			UUID:                  aws.String(uuid),
			EventSourceMappingArn: aws.String(eventSourceMappingArn),
			FunctionArn:           aws.String(functionArn),
			State:                 aws.String("Creating"),
			EventSourceArn:        aws.String("arn:aws:dynamodb:us-west-2:123456789012:table/test-table/stream/2023-01-01T00:00:00.000"),
			BatchSize:             aws.Int32(20),
			StartingPosition:      types.EventSourcePositionTrimHorizon,
			DestinationConfig: &types.DestinationConfig{
				OnFailure: &types.OnFailure{
					Destination: aws.String("arn:aws:sqs:us-west-2:123456789012:dlq"),
				},
			},
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":     core.MappingNodeFromString("test-dest-function"),
			"eventSourceArn":   core.MappingNodeFromString("arn:aws:dynamodb:us-west-2:123456789012:table/test-table/stream/2023-01-01T00:00:00.000"),
			"batchSize":        core.MappingNodeFromInt(20),
			"startingPosition": core.MappingNodeFromString("TRIM_HORIZON"),
			"destinationConfig": {
				Fields: map[string]*core.MappingNode{
					"onFailure": {
						Fields: map[string]*core.MappingNode{
							"destination": core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:dlq"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create event source mapping with destination config",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
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
			ResourceID: "test-dest-esm-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-dest-esm-id",
					ResourceName: "TestDestEventSourceMapping",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventSourceMapping",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.functionName"},
					{FieldPath: "spec.eventSourceArn"},
					{FieldPath: "spec.batchSize"},
					{FieldPath: "spec.startingPosition"},
					{FieldPath: "spec.destinationConfig"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.id":                    core.MappingNodeFromString(uuid),
				"spec.eventSourceMappingArn": core.MappingNodeFromString(eventSourceMappingArn),
				"spec.functionArn":           core.MappingNodeFromString(functionArn),
				"spec.state":                 core.MappingNodeFromString("Creating"),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateEventSourceMapping": &lambda.CreateEventSourceMappingInput{
				FunctionName:     aws.String("test-dest-function"),
				EventSourceArn:   aws.String("arn:aws:dynamodb:us-west-2:123456789012:table/test-table/stream/2023-01-01T00:00:00.000"),
				BatchSize:        aws.Int32(20),
				StartingPosition: types.EventSourcePositionTrimHorizon,
				DestinationConfig: &types.DestinationConfig{
					OnFailure: &types.OnFailure{
						Destination: aws.String("arn:aws:sqs:us-west-2:123456789012:dlq"),
					},
				},
			},
		},
	}
}

func createEventSourceMappingWithTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	uuid := "test-tagged-uuid"
	eventSourceMappingArn := "arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-tagged-uuid"
	functionArn := "arn:aws:lambda:us-west-2:123456789012:function:test-tagged-function"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateEventSourceMappingOutput(&lambda.CreateEventSourceMappingOutput{
			UUID:                  aws.String(uuid),
			EventSourceMappingArn: aws.String(eventSourceMappingArn),
			FunctionArn:           aws.String(functionArn),
			State:                 aws.String("Creating"),
			EventSourceArn:        aws.String("arn:aws:sqs:us-west-2:123456789012:test-tagged-queue"),
			BatchSize:             aws.Int32(15),
		}),
		lambdamock.WithTagResourceOutput(&lambda.TagResourceOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":   core.MappingNodeFromString("test-tagged-function"),
			"eventSourceArn": core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:test-tagged-queue"),
			"batchSize":      core.MappingNodeFromInt(15),
			"tags": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("test"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Team"),
							"value": core.MappingNodeFromString("platform"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Purpose"),
							"value": core.MappingNodeFromString("event-processing"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create event source mapping with tags",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
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
			ResourceID: "test-tagged-esm-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-tagged-esm-id",
					ResourceName: "TestTaggedEventSourceMapping",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventSourceMapping",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.functionName"},
					{FieldPath: "spec.eventSourceArn"},
					{FieldPath: "spec.batchSize"},
					{FieldPath: "spec.tags"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.id":                    core.MappingNodeFromString(uuid),
				"spec.eventSourceMappingArn": core.MappingNodeFromString(eventSourceMappingArn),
				"spec.functionArn":           core.MappingNodeFromString(functionArn),
				"spec.state":                 core.MappingNodeFromString("Creating"),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateEventSourceMapping": &lambda.CreateEventSourceMappingInput{
				FunctionName:   aws.String("test-tagged-function"),
				EventSourceArn: aws.String("arn:aws:sqs:us-west-2:123456789012:test-tagged-queue"),
				BatchSize:      aws.Int32(15),
			},
			"TagResource": &lambda.TagResourceInput{
				Resource: aws.String(eventSourceMappingArn),
				Tags: map[string]string{
					"Environment": "test",
					"Team":        "platform",
					"Purpose":     "event-processing",
				},
			},
		},
	}
}

func TestLambdaEventSourceMappingResourceCreate(t *testing.T) {
	suite.Run(t, new(LambdaEventSourceMappingResourceCreateSuite))
}
