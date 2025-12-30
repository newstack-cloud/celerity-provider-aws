//go:build unit

package dynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	dynamodbmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/dynamodb_mock"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type DynamoDBTableResourceUpdateSuite struct {
	suite.Suite
}

func (s *DynamoDBTableResourceUpdateSuite) Test_update_dynamodb_table() {
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
		updateBillingModeTestCase(providerCtx, loader),
		updateStreamSpecificationTestCase(providerCtx, loader),
		updateNoChangesTestCase(providerCtx, loader),
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

func updateBillingModeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/test-table"
	tableID := "12345678-1234-1234-1234-123456789012"

	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithUpdateTableOutput(&dynamodb.UpdateTableOutput{
			TableDescription: &types.TableDescription{
				TableArn:    aws.String(tableARN),
				TableId:     aws.String(tableID),
				TableName:   aws.String("test-table"),
				TableStatus: types.TableStatusUpdating,
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

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("test-table"),
			"arn":       core.MappingNodeFromString(tableARN),
			"tableId":   core.MappingNodeFromString(tableID),
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

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("test-table"),
			"arn":       core.MappingNodeFromString(tableARN),
			"tableId":   core.MappingNodeFromString(tableID),
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
				"readCapacityUnits", core.MappingNodeFromInt(10),
				"writeCapacityUnits", core.MappingNodeFromInt(5),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "update billing mode to provisioned",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-table-id",
						Name:       "TestTable",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/dynamodb/table",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{FieldPath: "spec.billingMode"},
					{FieldPath: "spec.provisionedThroughput.readCapacityUnits"},
					{FieldPath: "spec.provisionedThroughput.writeCapacityUnits"},
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
			"UpdateTable": &dynamodb.UpdateTableInput{
				TableName:   aws.String("test-table"),
				BillingMode: types.BillingModeProvisioned,
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(10),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
	}
}

func updateStreamSpecificationTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/stream-table"
	tableID := "22222222-2222-2222-2222-222222222222"
	streamARN := "arn:aws:dynamodb:us-west-2:123456789012:table/stream-table/stream/2024-01-01T00:00:00.000"
	streamLabel := "2024-01-01T00:00:00.000"

	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithUpdateTableOutput(&dynamodb.UpdateTableOutput{
			TableDescription: &types.TableDescription{
				TableArn:          aws.String(tableARN),
				TableId:           aws.String(tableID),
				TableName:         aws.String("stream-table"),
				TableStatus:       types.TableStatusUpdating,
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

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("stream-table"),
			"arn":       core.MappingNodeFromString(tableARN),
			"tableId":   core.MappingNodeFromString(tableID),
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

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("stream-table"),
			"arn":       core.MappingNodeFromString(tableARN),
			"tableId":   core.MappingNodeFromString(tableID),
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
		Name: "enable DynamoDB streams",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "stream-table-id",
						Name:       "StreamTable",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/dynamodb/table",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{FieldPath: "spec.streamSpecification.streamEnabled"},
					{FieldPath: "spec.streamSpecification.streamViewType"},
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
			"UpdateTable": &dynamodb.UpdateTableInput{
				TableName: aws.String("stream-table"),
				StreamSpecification: &types.StreamSpecification{
					StreamEnabled:  aws.Bool(true),
					StreamViewType: types.StreamViewTypeNewAndOldImages,
				},
			},
		},
	}
}

func updateNoChangesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, dynamodbservice.Service] {
	tableARN := "arn:aws:dynamodb:us-west-2:123456789012:table/no-change-table"
	tableID := "33333333-3333-3333-3333-333333333333"

	service := dynamodbmock.CreateDynamoDBServiceMock()

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("no-change-table"),
			"arn":       core.MappingNodeFromString(tableARN),
			"tableId":   core.MappingNodeFromString(tableID),
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

	// Same spec data - no changes
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("no-change-table"),
			"arn":       core.MappingNodeFromString(tableARN),
			"tableId":   core.MappingNodeFromString(tableID),
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
		Name: "no changes to table",
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
			ResourceID: "no-change-table-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "no-change-table-id",
					ResourceName: "NoChangeTable",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "no-change-table-id",
						Name:       "NoChangeTable",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/dynamodb/table",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":     core.MappingNodeFromString(tableARN),
				"spec.tableId": core.MappingNodeFromString(tableID),
			},
		},
		SaveActionsCalled: map[string]any{},
	}
}

func TestDynamoDBTableResourceUpdateSuite(t *testing.T) {
	suite.Run(t, new(DynamoDBTableResourceUpdateSuite))
}
