//go:build unit

package dynamodb

import (
	"fmt"
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
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type DynamoDBTableResourceCreateSuite struct {
	suite.Suite
}

func (s *DynamoDBTableResourceCreateSuite) Test_create_dynamodb_table() {
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

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service]{
		createBasicTableTestCase(providerCtx, loader),
		createTableWithProvisionedThroughputTestCase(providerCtx, loader),
		createTableWithGSITestCase(providerCtx, loader),
		createTableWithStreamsTestCase(providerCtx, loader),
		createTableWithTTLTestCase(providerCtx, loader),
		createTableWithPITRTestCase(providerCtx, loader),
		createTableFailureTestCase(providerCtx, loader),
	}

	tableResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, dynamodbservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return TableResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		tableResourceWrapper,
		&s.Suite,
	)
}

func createBasicTableTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/test-table"
	tableID := "12345678-1234-1234-1234-123456789012"

	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithCreateTableOutput(&dynamodb.CreateTableOutput{
			TableDescription: &types.TableDescription{
				TableArn:    aws.String(tableARN),
				TableId:     aws.String(tableID),
				TableName:   aws.String("test-table"),
				TableStatus: types.TableStatusActive,
			},
		}),
		dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
			Table: &types.TableDescription{
				TableArn:    aws.String(tableARN),
				TableId:     aws.String(tableID),
				TableName:   aws.String("test-table"),
				TableStatus: types.TableStatusActive,
			},
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("test-table"),
			"attributeDefinitions": core.MappingNodeItems(
				core.MappingNodeFields(
					"attributeName", core.MappingNodeFromString("id"),
					"attributeType", core.MappingNodeFromString("S"),
				),
			),
			"keySchema": core.MappingNodeItems(
				core.MappingNodeFields(
					"attributeName", core.MappingNodeFromString("id"),
					"keyType", core.MappingNodeFromString("HASH"),
				),
			),
			"billingMode": core.MappingNodeFromString("PAY_PER_REQUEST"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "create basic DynamoDB table",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
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
			ResourceID: "test-table-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-table-id",
					ResourceName: "TestTable",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/dynamodb/table",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.tableName"},
					{FieldPath: "spec.attributeDefinitions"},
					{FieldPath: "spec.keySchema"},
					{FieldPath: "spec.billingMode"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":     core.MappingNodeFromString(tableARN),
				"spec.tableId": core.MappingNodeFromString(tableID),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateTable": &dynamodb.CreateTableInput{
				TableName: aws.String("test-table"),
				AttributeDefinitions: []types.AttributeDefinition{
					{
						AttributeName: aws.String("id"),
						AttributeType: types.ScalarAttributeTypeS,
					},
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("id"),
						KeyType:       types.KeyTypeHash,
					},
				},
				BillingMode: types.BillingModePayPerRequest,
				Tags:        []types.Tag{},
			},
		},
	}
}

func createTableWithProvisionedThroughputTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/provisioned-table"
	tableID := "22222222-2222-2222-2222-222222222222"

	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithCreateTableOutput(&dynamodb.CreateTableOutput{
			TableDescription: &types.TableDescription{
				TableArn:    aws.String(tableARN),
				TableId:     aws.String(tableID),
				TableName:   aws.String("provisioned-table"),
				TableStatus: types.TableStatusActive,
			},
		}),
		dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
			Table: &types.TableDescription{
				TableArn:    aws.String(tableARN),
				TableId:     aws.String(tableID),
				TableName:   aws.String("provisioned-table"),
				TableStatus: types.TableStatusActive,
			},
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("provisioned-table"),
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
			"billingMode": core.MappingNodeFromString("PROVISIONED"),
			"provisionedThroughput": core.MappingNodeFields(
				"readCapacityUnits", core.MappingNodeFromInt(5),
				"writeCapacityUnits", core.MappingNodeFromInt(5),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "create DynamoDB table with provisioned throughput",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
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
			ResourceID: "provisioned-table-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "provisioned-table-id",
					ResourceName: "ProvisionedTable",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/dynamodb/table",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.tableName"},
					{FieldPath: "spec.attributeDefinitions"},
					{FieldPath: "spec.keySchema"},
					{FieldPath: "spec.billingMode"},
					{FieldPath: "spec.provisionedThroughput"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":     core.MappingNodeFromString(tableARN),
				"spec.tableId": core.MappingNodeFromString(tableID),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateTable": &dynamodb.CreateTableInput{
				TableName: aws.String("provisioned-table"),
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
				BillingMode: types.BillingModeProvisioned,
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
				Tags: []types.Tag{},
			},
		},
	}
}

func createTableWithGSITestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/gsi-table"
	tableID := "33333333-3333-3333-3333-333333333333"

	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithCreateTableOutput(&dynamodb.CreateTableOutput{
			TableDescription: &types.TableDescription{
				TableArn:    aws.String(tableARN),
				TableId:     aws.String(tableID),
				TableName:   aws.String("gsi-table"),
				TableStatus: types.TableStatusActive,
			},
		}),
		dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
			Table: &types.TableDescription{
				TableArn:    aws.String(tableARN),
				TableId:     aws.String(tableID),
				TableName:   aws.String("gsi-table"),
				TableStatus: types.TableStatusActive,
			},
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("gsi-table"),
			"attributeDefinitions": core.MappingNodeItems(
				core.MappingNodeFields(
					"attributeName", core.MappingNodeFromString("pk"),
					"attributeType", core.MappingNodeFromString("S"),
				),
				core.MappingNodeFields(
					"attributeName", core.MappingNodeFromString("gsi_pk"),
					"attributeType", core.MappingNodeFromString("S"),
				),
			),
			"keySchema": core.MappingNodeItems(
				core.MappingNodeFields(
					"attributeName", core.MappingNodeFromString("pk"),
					"keyType", core.MappingNodeFromString("HASH"),
				),
			),
			"billingMode": core.MappingNodeFromString("PAY_PER_REQUEST"),
			"globalSecondaryIndexes": core.MappingNodeItems(
				core.MappingNodeFields(
					"indexName", core.MappingNodeFromString("gsi-index"),
					"keySchema", core.MappingNodeItems(
						core.MappingNodeFields(
							"attributeName", core.MappingNodeFromString("gsi_pk"),
							"keyType", core.MappingNodeFromString("HASH"),
						),
					),
					"projection", core.MappingNodeFields(
						"projectionType", core.MappingNodeFromString("ALL"),
					),
				),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "create DynamoDB table with GSI",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
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
			ResourceID: "gsi-table-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "gsi-table-id",
					ResourceName: "GSITable",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/dynamodb/table",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.tableName"},
					{FieldPath: "spec.attributeDefinitions"},
					{FieldPath: "spec.keySchema"},
					{FieldPath: "spec.billingMode"},
					{FieldPath: "spec.globalSecondaryIndexes"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":     core.MappingNodeFromString(tableARN),
				"spec.tableId": core.MappingNodeFromString(tableID),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateTable": &dynamodb.CreateTableInput{
				TableName: aws.String("gsi-table"),
				AttributeDefinitions: []types.AttributeDefinition{
					{
						AttributeName: aws.String("pk"),
						AttributeType: types.ScalarAttributeTypeS,
					},
					{
						AttributeName: aws.String("gsi_pk"),
						AttributeType: types.ScalarAttributeTypeS,
					},
				},
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("pk"),
						KeyType:       types.KeyTypeHash,
					},
				},
				BillingMode: types.BillingModePayPerRequest,
				GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
					{
						IndexName: aws.String("gsi-index"),
						KeySchema: []types.KeySchemaElement{
							{
								AttributeName: aws.String("gsi_pk"),
								KeyType:       types.KeyTypeHash,
							},
						},
						Projection: &types.Projection{
							ProjectionType: types.ProjectionTypeAll,
						},
					},
				},
				Tags: []types.Tag{},
			},
		},
	}
}

func createTableWithStreamsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/stream-table"
	tableID := "44444444-4444-4444-4444-444444444444"
	streamARN := "arn:aws:dynamodb:us-west-2:123456789012:table/stream-table/stream/2024-01-01T00:00:00.000"
	streamLabel := "2024-01-01T00:00:00.000"

	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithCreateTableOutput(&dynamodb.CreateTableOutput{
			TableDescription: &types.TableDescription{
				TableArn:          aws.String(tableARN),
				TableId:           aws.String(tableID),
				TableName:         aws.String("stream-table"),
				TableStatus:       types.TableStatusActive,
				LatestStreamArn:   aws.String(streamARN),
				LatestStreamLabel: aws.String(streamLabel),
			},
		}),
		dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
			Table: &types.TableDescription{
				TableArn:          aws.String(tableARN),
				TableId:           aws.String(tableID),
				TableName:         aws.String("stream-table"),
				TableStatus:       types.TableStatusActive,
				LatestStreamArn:   aws.String(streamARN),
				LatestStreamLabel: aws.String(streamLabel),
			},
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("stream-table"),
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
			"billingMode": core.MappingNodeFromString("PAY_PER_REQUEST"),
			"streamSpecification": core.MappingNodeFields(
				"streamEnabled", core.MappingNodeFromBool(true),
				"streamViewType", core.MappingNodeFromString("NEW_AND_OLD_IMAGES"),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "create DynamoDB table with streams",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
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
			ResourceID: "stream-table-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "stream-table-id",
					ResourceName: "StreamTable",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/dynamodb/table",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.tableName"},
					{FieldPath: "spec.attributeDefinitions"},
					{FieldPath: "spec.keySchema"},
					{FieldPath: "spec.billingMode"},
					{FieldPath: "spec.streamSpecification"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":               core.MappingNodeFromString(tableARN),
				"spec.tableId":           core.MappingNodeFromString(tableID),
				"spec.latestStreamArn":   core.MappingNodeFromString(streamARN),
				"spec.latestStreamLabel": core.MappingNodeFromString(streamLabel),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateTable": &dynamodb.CreateTableInput{
				TableName: aws.String("stream-table"),
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
				BillingMode: types.BillingModePayPerRequest,
				StreamSpecification: &types.StreamSpecification{
					StreamEnabled:  aws.Bool(true),
					StreamViewType: types.StreamViewTypeNewAndOldImages,
				},
				Tags: []types.Tag{},
			},
		},
	}
}

func createTableWithTTLTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/ttl-table"
	tableID := "55555555-5555-5555-5555-555555555555"

	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithCreateTableOutput(&dynamodb.CreateTableOutput{
			TableDescription: &types.TableDescription{
				TableArn:    aws.String(tableARN),
				TableId:     aws.String(tableID),
				TableName:   aws.String("ttl-table"),
				TableStatus: types.TableStatusActive,
			},
		}),
		dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
			Table: &types.TableDescription{
				TableArn:    aws.String(tableARN),
				TableId:     aws.String(tableID),
				TableName:   aws.String("ttl-table"),
				TableStatus: types.TableStatusActive,
			},
		}),
		dynamodbmock.WithUpdateTimeToLiveOutput(&dynamodb.UpdateTimeToLiveOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("ttl-table"),
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
			"billingMode": core.MappingNodeFromString("PAY_PER_REQUEST"),
			"timeToLiveSpecification": core.MappingNodeFields(
				"enabled", core.MappingNodeFromBool(true),
				"attributeName", core.MappingNodeFromString("expiresAt"),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "create DynamoDB table with TTL",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
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
			ResourceID: "ttl-table-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "ttl-table-id",
					ResourceName: "TTLTable",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/dynamodb/table",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.tableName"},
					{FieldPath: "spec.attributeDefinitions"},
					{FieldPath: "spec.keySchema"},
					{FieldPath: "spec.billingMode"},
					{FieldPath: "spec.timeToLiveSpecification"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":     core.MappingNodeFromString(tableARN),
				"spec.tableId": core.MappingNodeFromString(tableID),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateTable": &dynamodb.CreateTableInput{
				TableName: aws.String("ttl-table"),
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
				BillingMode: types.BillingModePayPerRequest,
				Tags:        []types.Tag{},
			},
			"UpdateTimeToLive": &dynamodb.UpdateTimeToLiveInput{
				TableName: aws.String("ttl-table"),
				TimeToLiveSpecification: &types.TimeToLiveSpecification{
					Enabled:       aws.Bool(true),
					AttributeName: aws.String("expiresAt"),
				},
			},
		},
	}
}

func createTableWithPITRTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/pitr-table"
	tableID := "66666666-6666-6666-6666-666666666666"

	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithCreateTableOutput(&dynamodb.CreateTableOutput{
			TableDescription: &types.TableDescription{
				TableArn:    aws.String(tableARN),
				TableId:     aws.String(tableID),
				TableName:   aws.String("pitr-table"),
				TableStatus: types.TableStatusActive,
			},
		}),
		dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
			Table: &types.TableDescription{
				TableArn:    aws.String(tableARN),
				TableId:     aws.String(tableID),
				TableName:   aws.String("pitr-table"),
				TableStatus: types.TableStatusActive,
			},
		}),
		dynamodbmock.WithUpdateContinuousBackupsOutput(&dynamodb.UpdateContinuousBackupsOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("pitr-table"),
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
			"billingMode": core.MappingNodeFromString("PAY_PER_REQUEST"),
			"pointInTimeRecoverySpecification": core.MappingNodeFields(
				"pointInTimeRecoveryEnabled", core.MappingNodeFromBool(true),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "create DynamoDB table with PITR",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
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
			ResourceID: "pitr-table-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "pitr-table-id",
					ResourceName: "PITRTable",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/dynamodb/table",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{FieldPath: "spec.tableName"},
					{FieldPath: "spec.attributeDefinitions"},
					{FieldPath: "spec.keySchema"},
					{FieldPath: "spec.billingMode"},
					{FieldPath: "spec.pointInTimeRecoverySpecification"},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":     core.MappingNodeFromString(tableARN),
				"spec.tableId": core.MappingNodeFromString(tableID),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateTable": &dynamodb.CreateTableInput{
				TableName: aws.String("pitr-table"),
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
				BillingMode: types.BillingModePayPerRequest,
				Tags:        []types.Tag{},
			},
			"UpdateContinuousBackups": &dynamodb.UpdateContinuousBackupsInput{
				TableName: aws.String("pitr-table"),
				PointInTimeRecoverySpecification: &types.PointInTimeRecoverySpecification{
					PointInTimeRecoveryEnabled: aws.Bool(true),
				},
			},
		},
	}
}

func createTableFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service] {
	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithCreateTableError(fmt.Errorf("failed to create table")),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("failed-table"),
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
			"billingMode": core.MappingNodeFromString("PAY_PER_REQUEST"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "create DynamoDB table failure",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
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
			ResourceID: "failed-table-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "failed-table-id",
					ResourceName: "FailedTable",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/dynamodb/table",
						},
						Spec: specData,
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"CreateTable": &dynamodb.CreateTableInput{
				TableName: aws.String("failed-table"),
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
				BillingMode: types.BillingModePayPerRequest,
				Tags:        []types.Tag{},
			},
		},
	}
}

func TestDynamoDBTableResourceCreateSuite(t *testing.T) {
	suite.Run(t, new(DynamoDBTableResourceCreateSuite))
}
