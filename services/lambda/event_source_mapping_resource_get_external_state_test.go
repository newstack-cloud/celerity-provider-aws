//go:build unit

package lambda

import (
	"errors"
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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type LambdaEventSourceMappingResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *LambdaEventSourceMappingResourceGetExternalStateSuite) Test_get_external_state() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			pluginutils.SessionIDKey: core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		createBasicEventSourceMappingStateTestCase(providerCtx, loader),
		createAllOptionalConfigsEventSourceMappingTestCase(providerCtx, loader),
		createGetEventSourceMappingErrorTestCase(providerCtx, loader),
		createNoUUIDTestCase(providerCtx, loader),
		createWithFilterCriteriaTestCase(providerCtx, loader),
		createWithDestinationConfigTestCase(providerCtx, loader),
		createWithSourceAccessConfigurationsTestCase(providerCtx, loader),
		createWithComplexConfigurationsTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	eventSourceMappingResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return EventSourceMappingResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		eventSourceMappingResourceWrapper,
		&s.Suite,
	)
}

func TestLambdaEventSourceMappingResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(LambdaEventSourceMappingResourceGetExternalStateSuite))
}

// Test case generator functions below.

func createBasicEventSourceMappingStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets basic event source mapping state",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetEventSourceMappingOutput(&lambda.GetEventSourceMappingOutput{
				UUID:                           aws.String("test-uuid-123"),
				EventSourceMappingArn:          aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-uuid-123"),
				FunctionArn:                    aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				State:                          aws.String("Enabled"),
				EventSourceArn:                 aws.String("arn:aws:kinesis:us-west-2:123456789012:stream/test-stream"),
				BatchSize:                      aws.Int32(100),
				StartingPosition:               types.EventSourcePositionLatest,
				MaximumBatchingWindowInSeconds: aws.Int32(5),
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id": core.MappingNodeFromString("test-uuid-123"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id":                             core.MappingNodeFromString("test-uuid-123"),
					"eventSourceMappingArn":          core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-uuid-123"),
					"functionArn":                    core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
					"enabled":                        core.MappingNodeFromBool(true),
					"eventSourceArn":                 core.MappingNodeFromString("arn:aws:kinesis:us-west-2:123456789012:stream/test-stream"),
					"batchSize":                      core.MappingNodeFromInt(100),
					"state":                          core.MappingNodeFromString("Enabled"),
					"startingPosition":               core.MappingNodeFromString("LATEST"),
					"maximumBatchingWindowInSeconds": core.MappingNodeFromInt(5),
				},
			},
		},
		ExpectError: false,
	}
}

func createAllOptionalConfigsEventSourceMappingTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets event source mapping state with all optional configurations",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetEventSourceMappingOutput(&lambda.GetEventSourceMappingOutput{
				UUID:                           aws.String("test-uuid-456"),
				EventSourceMappingArn:          aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-uuid-456"),
				FunctionArn:                    aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				State:                          aws.String("Enabled"),
				EventSourceArn:                 aws.String("arn:aws:kinesis:us-west-2:123456789012:stream/test-stream"),
				BatchSize:                      aws.Int32(200),
				StartingPosition:               types.EventSourcePositionTrimHorizon,
				MaximumBatchingWindowInSeconds: aws.Int32(10),
				MaximumRecordAgeInSeconds:      aws.Int32(3600),
				MaximumRetryAttempts:           aws.Int32(3),
				BisectBatchOnFunctionError:     aws.Bool(true),
				ParallelizationFactor:          aws.Int32(2),
				TumblingWindowInSeconds:        aws.Int32(300),
				KMSKeyArn:                      aws.String("arn:aws:kms:us-west-2:123456789012:key/test-key"),
				FunctionResponseTypes:          []types.FunctionResponseType{types.FunctionResponseTypeReportBatchItemFailures},
				Topics:                         []string{"test-topic-1", "test-topic-2"},
				Queues:                         []string{"test-queue-1", "test-queue-2"},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id": core.MappingNodeFromString("test-uuid-456"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id":                             core.MappingNodeFromString("test-uuid-456"),
					"eventSourceMappingArn":          core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-uuid-456"),
					"functionArn":                    core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
					"enabled":                        core.MappingNodeFromBool(true),
					"eventSourceArn":                 core.MappingNodeFromString("arn:aws:kinesis:us-west-2:123456789012:stream/test-stream"),
					"batchSize":                      core.MappingNodeFromInt(200),
					"state":                          core.MappingNodeFromString("Enabled"),
					"startingPosition":               core.MappingNodeFromString("TRIM_HORIZON"),
					"maximumBatchingWindowInSeconds": core.MappingNodeFromInt(10),
					"maximumRecordAgeInSeconds":      core.MappingNodeFromInt(3600),
					"maximumRetryAttempts":           core.MappingNodeFromInt(3),
					"bisectBatchOnFunctionError":     core.MappingNodeFromBool(true),
					"parallelizationFactor":          core.MappingNodeFromInt(2),
					"tumblingWindowInSeconds":        core.MappingNodeFromInt(300),
					"kmsKeyArn":                      core.MappingNodeFromString("arn:aws:kms:us-west-2:123456789012:key/test-key"),
					"functionResponseTypes": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("ReportBatchItemFailures"),
						},
					},
					"topics": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("test-topic-1"),
							core.MappingNodeFromString("test-topic-2"),
						},
					},
					"queues": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("test-queue-1"),
							core.MappingNodeFromString("test-queue-2"),
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createGetEventSourceMappingErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "handles get event source mapping error",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetEventSourceMappingError(errors.New("failed to get event source mapping")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id": core.MappingNodeFromString("test-uuid-123"),
				},
			},
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createNoUUIDTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name:           "returns empty state when no UUID is present",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					// No UUID field present
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{},
			},
		},
		ExpectError: false,
	}
}

func createWithFilterCriteriaTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets event source mapping with filter criteria",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetEventSourceMappingOutput(&lambda.GetEventSourceMappingOutput{
				UUID:                  aws.String("test-uuid-filter"),
				EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-uuid-filter"),
				FunctionArn:           aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				State:                 aws.String("Enabled"),
				FilterCriteria: &types.FilterCriteria{
					Filters: []types.Filter{
						{
							Pattern: aws.String("{\"source\":[\"aws.s3\"]}"),
						},
						{
							Pattern: aws.String("{\"detail-type\":[\"Object Created\"]}"),
						},
					},
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id": core.MappingNodeFromString("test-uuid-filter"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id":                    core.MappingNodeFromString("test-uuid-filter"),
					"eventSourceMappingArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-uuid-filter"),
					"functionArn":           core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
					"enabled":               core.MappingNodeFromBool(true),
					"state":                 core.MappingNodeFromString("Enabled"),
					"filterCriteria": {
						Fields: map[string]*core.MappingNode{
							"filters": {
								Items: []*core.MappingNode{
									{
										Fields: map[string]*core.MappingNode{
											"pattern": core.MappingNodeFromString("{\"source\":[\"aws.s3\"]}"),
										},
									},
									{
										Fields: map[string]*core.MappingNode{
											"pattern": core.MappingNodeFromString("{\"detail-type\":[\"Object Created\"]}"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createWithDestinationConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets event source mapping with destination configuration",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetEventSourceMappingOutput(&lambda.GetEventSourceMappingOutput{
				UUID:                  aws.String("test-uuid-dest"),
				EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-uuid-dest"),
				FunctionArn:           aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				State:                 aws.String("Enabled"),
				DestinationConfig: &types.DestinationConfig{
					OnFailure: &types.OnFailure{
						Destination: aws.String("arn:aws:sqs:us-west-2:123456789012:dlq-queue"),
					},
					OnSuccess: &types.OnSuccess{
						Destination: aws.String("arn:aws:sqs:us-west-2:123456789012:success-queue"),
					},
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id": core.MappingNodeFromString("test-uuid-dest"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id":                    core.MappingNodeFromString("test-uuid-dest"),
					"eventSourceMappingArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-uuid-dest"),
					"functionArn":           core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
					"enabled":               core.MappingNodeFromBool(true),
					"state":                 core.MappingNodeFromString("Enabled"),
					"destinationConfig": {
						Fields: map[string]*core.MappingNode{
							"onFailure": {
								Fields: map[string]*core.MappingNode{
									"destination": core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:dlq-queue"),
								},
							},
							"onSuccess": {
								Fields: map[string]*core.MappingNode{
									"destination": core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:success-queue"),
								},
							},
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createWithSourceAccessConfigurationsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets event source mapping with source access configurations",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetEventSourceMappingOutput(&lambda.GetEventSourceMappingOutput{
				UUID:                  aws.String("test-uuid-sac"),
				EventSourceMappingArn: aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-uuid-sac"),
				FunctionArn:           aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				State:                 aws.String("Enabled"),
				SourceAccessConfigurations: []types.SourceAccessConfiguration{
					{
						Type: types.SourceAccessTypeVpcSubnet,
						URI:  aws.String("subnet-12345678"),
					},
					{
						Type: types.SourceAccessTypeVpcSecurityGroup,
						URI:  aws.String("sg-12345678"),
					},
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id": core.MappingNodeFromString("test-uuid-sac"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id":                    core.MappingNodeFromString("test-uuid-sac"),
					"eventSourceMappingArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-uuid-sac"),
					"functionArn":           core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
					"enabled":               core.MappingNodeFromBool(true),
					"state":                 core.MappingNodeFromString("Enabled"),
					"sourceAccessConfigurations": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"type": core.MappingNodeFromString("VPC_SUBNET"),
									"uri":  core.MappingNodeFromString("subnet-12345678"),
								},
							},
							{
								Fields: map[string]*core.MappingNode{
									"type": core.MappingNodeFromString("VPC_SECURITY_GROUP"),
									"uri":  core.MappingNodeFromString("sg-12345678"),
								},
							},
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createWithComplexConfigurationsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets event source mapping with complex configurations",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetEventSourceMappingOutput(&lambda.GetEventSourceMappingOutput{
				UUID:                           aws.String("test-uuid-complex"),
				EventSourceMappingArn:          aws.String("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-uuid-complex"),
				FunctionArn:                    aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				State:                          aws.String("Disabled"),
				EventSourceArn:                 aws.String("arn:aws:kinesis:us-west-2:123456789012:stream/complex-stream"),
				BatchSize:                      aws.Int32(500),
				StartingPosition:               types.EventSourcePositionAtTimestamp,
				MaximumBatchingWindowInSeconds: aws.Int32(15),
				MaximumRecordAgeInSeconds:      aws.Int32(7200),
				MaximumRetryAttempts:           aws.Int32(5),
				BisectBatchOnFunctionError:     aws.Bool(false),
				ParallelizationFactor:          aws.Int32(3),
				TumblingWindowInSeconds:        aws.Int32(600),
				KMSKeyArn:                      aws.String("arn:aws:kms:us-west-2:123456789012:key/complex-key"),
				FunctionResponseTypes:          []types.FunctionResponseType{types.FunctionResponseTypeReportBatchItemFailures},
				Topics:                         []string{"complex-topic"},
				Queues:                         []string{"complex-queue"},
				FilterCriteria: &types.FilterCriteria{
					Filters: []types.Filter{
						{
							Pattern: aws.String("{\"source\":[\"aws.kinesis\"]}"),
						},
					},
				},
				DestinationConfig: &types.DestinationConfig{
					OnFailure: &types.OnFailure{
						Destination: aws.String("arn:aws:sqs:us-west-2:123456789012:complex-dlq"),
					},
				},
				SourceAccessConfigurations: []types.SourceAccessConfiguration{
					{
						Type: types.SourceAccessTypeVpcSubnet,
						URI:  aws.String("subnet-complex"),
					},
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id": core.MappingNodeFromString("test-uuid-complex"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"id":                             core.MappingNodeFromString("test-uuid-complex"),
					"eventSourceMappingArn":          core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-uuid-complex"),
					"functionArn":                    core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
					"enabled":                        core.MappingNodeFromBool(false),
					"eventSourceArn":                 core.MappingNodeFromString("arn:aws:kinesis:us-west-2:123456789012:stream/complex-stream"),
					"batchSize":                      core.MappingNodeFromInt(500),
					"state":                          core.MappingNodeFromString("Disabled"),
					"startingPosition":               core.MappingNodeFromString("AT_TIMESTAMP"),
					"maximumBatchingWindowInSeconds": core.MappingNodeFromInt(15),
					"maximumRecordAgeInSeconds":      core.MappingNodeFromInt(7200),
					"maximumRetryAttempts":           core.MappingNodeFromInt(5),
					"bisectBatchOnFunctionError":     core.MappingNodeFromBool(false),
					"parallelizationFactor":          core.MappingNodeFromInt(3),
					"tumblingWindowInSeconds":        core.MappingNodeFromInt(600),
					"kmsKeyArn":                      core.MappingNodeFromString("arn:aws:kms:us-west-2:123456789012:key/complex-key"),
					"functionResponseTypes": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("ReportBatchItemFailures"),
						},
					},
					"topics": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("complex-topic"),
						},
					},
					"queues": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("complex-queue"),
						},
					},
					"filterCriteria": {
						Fields: map[string]*core.MappingNode{
							"filters": {
								Items: []*core.MappingNode{
									{
										Fields: map[string]*core.MappingNode{
											"pattern": core.MappingNodeFromString("{\"source\":[\"aws.kinesis\"]}"),
										},
									},
								},
							},
						},
					},
					"destinationConfig": {
						Fields: map[string]*core.MappingNode{
							"onFailure": {
								Fields: map[string]*core.MappingNode{
									"destination": core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:complex-dlq"),
								},
							},
						},
					},
					"sourceAccessConfigurations": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"type": core.MappingNodeFromString("VPC_SUBNET"),
									"uri":  core.MappingNodeFromString("subnet-complex"),
								},
							},
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}
