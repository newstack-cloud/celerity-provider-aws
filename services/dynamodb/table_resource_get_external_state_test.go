//go:build unit

package dynamodb

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dynamodbmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/dynamodb_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type DynamoDBTableResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *DynamoDBTableResourceGetExternalStateSuite) Test_get_external_state_dynamodb_table() {
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

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, dynamodbservice.Service]{
		getExternalStateBasicTableTestCase(providerCtx, loader),
		getExternalStateTableWithStreamsTestCase(providerCtx, loader),
		getExternalStateTableWithTTLTestCase(providerCtx, loader),
		getExternalStateTableWithPITRTestCase(providerCtx, loader),
		getExternalStateTableNotFoundTestCase(providerCtx, loader),
		getExternalStateErrorTestCase(providerCtx, loader),
	}

	tableResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, dynamodbservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return TableResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		tableResourceWrapper,
		&s.Suite,
	)
}

func getExternalStateBasicTableTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/test-table"
	tableID := "12345678-1234-1234-1234-123456789012"

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "successfully gets basic table state",
		ServiceFactory: dynamodbmock.CreateDynamoDBServiceMockFactory(
			dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
				Table: &types.TableDescription{
					TableArn:    aws.String(tableARN),
					TableId:     aws.String(tableID),
					TableName:   aws.String("test-table"),
					TableStatus: types.TableStatusActive,
					BillingModeSummary: &types.BillingModeSummary{
						BillingMode: types.BillingModePayPerRequest,
					},
					AttributeDefinitions: []types.AttributeDefinition{
						{
							AttributeName: aws.String("pk"),
							AttributeType: types.ScalarAttributeTypeS,
						},
					},
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("pk"),
							KeyType:       types.KeyTypeHash,
						},
					},
				},
			}),
			dynamodbmock.WithDescribeTimeToLiveOutput(&dynamodb.DescribeTimeToLiveOutput{
				TimeToLiveDescription: &types.TimeToLiveDescription{
					TimeToLiveStatus: types.TimeToLiveStatusDisabled,
				},
			}),
			dynamodbmock.WithDescribeContinuousBackupsOutput(&dynamodb.DescribeContinuousBackupsOutput{
				ContinuousBackupsDescription: &types.ContinuousBackupsDescription{
					PointInTimeRecoveryDescription: &types.PointInTimeRecoveryDescription{
						PointInTimeRecoveryStatus: types.PointInTimeRecoveryStatusDisabled,
					},
				},
			}),
			dynamodbmock.WithListTagsOfResourceOutput(&dynamodb.ListTagsOfResourceOutput{
				Tags: []types.Tag{},
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
					"arn":       core.MappingNodeFromString(tableARN),
					"tableName": core.MappingNodeFromString("test-table"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":     core.MappingNodeFromString(tableARN),
					"tableId": core.MappingNodeFromString(tableID),
					"tableName": core.MappingNodeFromString("test-table"),
					"billingMode": core.MappingNodeFromString("PAY_PER_REQUEST"),
					"attributeDefinitions": core.MappingNodeItems(
						core.MappingNodeFields(
							"attributeName", core.MappingNodeFromString("pk"),
							"attributeType", core.MappingNodeFromString("S"),
						),
					),
					"keySchema": core.MappingNodeItems(
						core.MappingNodeFields(
							"attributeName", core.MappingNodeFromString("pk"),
							"keyType", core.MappingNodeFromString("HASH"),
						),
					),
				},
			},
		},
	}
}

func getExternalStateTableWithStreamsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/stream-table"
	tableID := "22222222-2222-2222-2222-222222222222"
	streamARN := "arn:aws:dynamodb:us-west-2:123456789012:table/stream-table/stream/2024-01-01T00:00:00.000"
	streamLabel := "2024-01-01T00:00:00.000"

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "gets table state with streams enabled",
		ServiceFactory: dynamodbmock.CreateDynamoDBServiceMockFactory(
			dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
				Table: &types.TableDescription{
					TableArn:          aws.String(tableARN),
					TableId:           aws.String(tableID),
					TableName:         aws.String("stream-table"),
					TableStatus:       types.TableStatusActive,
					LatestStreamArn:   aws.String(streamARN),
					LatestStreamLabel: aws.String(streamLabel),
					BillingModeSummary: &types.BillingModeSummary{
						BillingMode: types.BillingModePayPerRequest,
					},
					StreamSpecification: &types.StreamSpecification{
						StreamEnabled:  aws.Bool(true),
						StreamViewType: types.StreamViewTypeNewAndOldImages,
					},
					AttributeDefinitions: []types.AttributeDefinition{
						{
							AttributeName: aws.String("pk"),
							AttributeType: types.ScalarAttributeTypeS,
						},
					},
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("pk"),
							KeyType:       types.KeyTypeHash,
						},
					},
				},
			}),
			dynamodbmock.WithDescribeTimeToLiveOutput(&dynamodb.DescribeTimeToLiveOutput{
				TimeToLiveDescription: &types.TimeToLiveDescription{
					TimeToLiveStatus: types.TimeToLiveStatusDisabled,
				},
			}),
			dynamodbmock.WithDescribeContinuousBackupsOutput(&dynamodb.DescribeContinuousBackupsOutput{
				ContinuousBackupsDescription: &types.ContinuousBackupsDescription{
					PointInTimeRecoveryDescription: &types.PointInTimeRecoveryDescription{
						PointInTimeRecoveryStatus: types.PointInTimeRecoveryStatusDisabled,
					},
				},
			}),
			dynamodbmock.WithListTagsOfResourceOutput(&dynamodb.ListTagsOfResourceOutput{
				Tags: []types.Tag{},
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
					"arn":       core.MappingNodeFromString(tableARN),
					"tableName": core.MappingNodeFromString("stream-table"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":               core.MappingNodeFromString(tableARN),
					"tableId":           core.MappingNodeFromString(tableID),
					"tableName":         core.MappingNodeFromString("stream-table"),
					"billingMode":       core.MappingNodeFromString("PAY_PER_REQUEST"),
					"latestStreamArn":   core.MappingNodeFromString(streamARN),
					"latestStreamLabel": core.MappingNodeFromString(streamLabel),
					"streamSpecification": core.MappingNodeFields(
						"streamEnabled", core.MappingNodeFromBool(true),
						"streamViewType", core.MappingNodeFromString("NEW_AND_OLD_IMAGES"),
					),
					"attributeDefinitions": core.MappingNodeItems(
						core.MappingNodeFields(
							"attributeName", core.MappingNodeFromString("pk"),
							"attributeType", core.MappingNodeFromString("S"),
						),
					),
					"keySchema": core.MappingNodeItems(
						core.MappingNodeFields(
							"attributeName", core.MappingNodeFromString("pk"),
							"keyType", core.MappingNodeFromString("HASH"),
						),
					),
				},
			},
		},
	}
}

func getExternalStateTableWithTTLTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/ttl-table"
	tableID := "33333333-3333-3333-3333-333333333333"

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "gets table state with TTL enabled",
		ServiceFactory: dynamodbmock.CreateDynamoDBServiceMockFactory(
			dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
				Table: &types.TableDescription{
					TableArn:    aws.String(tableARN),
					TableId:     aws.String(tableID),
					TableName:   aws.String("ttl-table"),
					TableStatus: types.TableStatusActive,
					BillingModeSummary: &types.BillingModeSummary{
						BillingMode: types.BillingModePayPerRequest,
					},
					AttributeDefinitions: []types.AttributeDefinition{
						{
							AttributeName: aws.String("pk"),
							AttributeType: types.ScalarAttributeTypeS,
						},
					},
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("pk"),
							KeyType:       types.KeyTypeHash,
						},
					},
				},
			}),
			dynamodbmock.WithDescribeTimeToLiveOutput(&dynamodb.DescribeTimeToLiveOutput{
				TimeToLiveDescription: &types.TimeToLiveDescription{
					TimeToLiveStatus: types.TimeToLiveStatusEnabled,
					AttributeName:    aws.String("expiresAt"),
				},
			}),
			dynamodbmock.WithDescribeContinuousBackupsOutput(&dynamodb.DescribeContinuousBackupsOutput{
				ContinuousBackupsDescription: &types.ContinuousBackupsDescription{
					PointInTimeRecoveryDescription: &types.PointInTimeRecoveryDescription{
						PointInTimeRecoveryStatus: types.PointInTimeRecoveryStatusDisabled,
					},
				},
			}),
			dynamodbmock.WithListTagsOfResourceOutput(&dynamodb.ListTagsOfResourceOutput{
				Tags: []types.Tag{},
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
					"arn":       core.MappingNodeFromString(tableARN),
					"tableName": core.MappingNodeFromString("ttl-table"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":         core.MappingNodeFromString(tableARN),
					"tableId":     core.MappingNodeFromString(tableID),
					"tableName":   core.MappingNodeFromString("ttl-table"),
					"billingMode": core.MappingNodeFromString("PAY_PER_REQUEST"),
					"timeToLiveSpecification": core.MappingNodeFields(
						"enabled", core.MappingNodeFromBool(true),
						"attributeName", core.MappingNodeFromString("expiresAt"),
					),
					"attributeDefinitions": core.MappingNodeItems(
						core.MappingNodeFields(
							"attributeName", core.MappingNodeFromString("pk"),
							"attributeType", core.MappingNodeFromString("S"),
						),
					),
					"keySchema": core.MappingNodeItems(
						core.MappingNodeFields(
							"attributeName", core.MappingNodeFromString("pk"),
							"keyType", core.MappingNodeFromString("HASH"),
						),
					),
				},
			},
		},
	}
}

func getExternalStateTableWithPITRTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/pitr-table"
	tableID := "44444444-4444-4444-4444-444444444444"

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "gets table state with PITR enabled",
		ServiceFactory: dynamodbmock.CreateDynamoDBServiceMockFactory(
			dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
				Table: &types.TableDescription{
					TableArn:    aws.String(tableARN),
					TableId:     aws.String(tableID),
					TableName:   aws.String("pitr-table"),
					TableStatus: types.TableStatusActive,
					BillingModeSummary: &types.BillingModeSummary{
						BillingMode: types.BillingModePayPerRequest,
					},
					AttributeDefinitions: []types.AttributeDefinition{
						{
							AttributeName: aws.String("pk"),
							AttributeType: types.ScalarAttributeTypeS,
						},
					},
					KeySchema: []types.KeySchemaElement{
						{
							AttributeName: aws.String("pk"),
							KeyType:       types.KeyTypeHash,
						},
					},
				},
			}),
			dynamodbmock.WithDescribeTimeToLiveOutput(&dynamodb.DescribeTimeToLiveOutput{
				TimeToLiveDescription: &types.TimeToLiveDescription{
					TimeToLiveStatus: types.TimeToLiveStatusDisabled,
				},
			}),
			dynamodbmock.WithDescribeContinuousBackupsOutput(&dynamodb.DescribeContinuousBackupsOutput{
				ContinuousBackupsDescription: &types.ContinuousBackupsDescription{
					PointInTimeRecoveryDescription: &types.PointInTimeRecoveryDescription{
						PointInTimeRecoveryStatus: types.PointInTimeRecoveryStatusEnabled,
					},
				},
			}),
			dynamodbmock.WithListTagsOfResourceOutput(&dynamodb.ListTagsOfResourceOutput{
				Tags: []types.Tag{},
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
					"arn":       core.MappingNodeFromString(tableARN),
					"tableName": core.MappingNodeFromString("pitr-table"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":         core.MappingNodeFromString(tableARN),
					"tableId":     core.MappingNodeFromString(tableID),
					"tableName":   core.MappingNodeFromString("pitr-table"),
					"billingMode": core.MappingNodeFromString("PAY_PER_REQUEST"),
					"pointInTimeRecoverySpecification": core.MappingNodeFields(
						"pointInTimeRecoveryEnabled", core.MappingNodeFromBool(true),
					),
					"attributeDefinitions": core.MappingNodeItems(
						core.MappingNodeFields(
							"attributeName", core.MappingNodeFromString("pk"),
							"attributeType", core.MappingNodeFromString("S"),
						),
					),
					"keySchema": core.MappingNodeItems(
						core.MappingNodeFields(
							"attributeName", core.MappingNodeFromString("pk"),
							"keyType", core.MappingNodeFromString("HASH"),
						),
					),
				},
			},
		},
	}
}

func getExternalStateTableNotFoundTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/not-found-table"

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "returns empty state when table not found",
		ServiceFactory: dynamodbmock.CreateDynamoDBServiceMockFactory(
			dynamodbmock.WithDescribeTableError(&types.ResourceNotFoundException{
				Message: aws.String("Table not found"),
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
					"arn":       core.MappingNodeFromString(tableARN),
					"tableName": core.MappingNodeFromString("not-found-table"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: nil,
		},
	}
}

func getExternalStateErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/error-table"

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "returns error on describe table failure",
		ServiceFactory: dynamodbmock.CreateDynamoDBServiceMockFactory(
			dynamodbmock.WithDescribeTableError(errors.New("internal server error")),
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
					"arn":       core.MappingNodeFromString(tableARN),
					"tableName": core.MappingNodeFromString("error-table"),
				},
			},
		},
		ExpectError: true,
	}
}

func TestDynamoDBTableResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(DynamoDBTableResourceGetExternalStateSuite))
}
