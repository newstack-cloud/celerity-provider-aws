//go:build unit

package dynamodb

import (
	"context"
	"errors"
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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type DynamoDBTableDataSourceSuite struct {
	suite.Suite
}

type DynamoDBDataSourceFetchTestCase struct {
	Name                 string
	ServiceFactory       func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service
	ConfigStore          pluginutils.ServiceConfigStore[*aws.Config]
	Input                *provider.DataSourceFetchInput
	ExpectedOutput       *provider.DataSourceFetchOutput
	ExpectError          bool
	ExpectedErrorMessage string
}

func (s *DynamoDBTableDataSourceSuite) Test_fetch() {
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

	testCases := []DynamoDBDataSourceFetchTestCase{
		createBasicTableDataSourceTestCase(providerCtx, loader),
		createTableDataSourceWithStreamsTestCase(providerCtx, loader),
		createTableDataSourceWithTTLTestCase(providerCtx, loader),
		createTableDataSourceWithPITRTestCase(providerCtx, loader),
		createTableDataSourceNotFoundTestCase(providerCtx, loader),
		createTableDataSourceFetchErrorTestCase(providerCtx, loader),
		createTableDataSourceMissingFilterTestCase(providerCtx, loader),
	}

	for _, tc := range testCases {
		s.Run(tc.Name, func() {
			dataSource := TableDataSource(tc.ServiceFactory, tc.ConfigStore)

			output, err := dataSource.Fetch(context.Background(), tc.Input)

			if tc.ExpectError {
				s.Error(err)
				if tc.ExpectedErrorMessage != "" {
					s.ErrorContains(err, tc.ExpectedErrorMessage)
				}
				s.Nil(output)
			} else {
				s.NoError(err)
				s.NotNil(output)
				s.Equal(tc.ExpectedOutput.Data, output.Data)
			}
		})
	}
}

func TestDynamoDBTableDataSourceSuite(t *testing.T) {
	suite.Run(t, new(DynamoDBTableDataSourceSuite))
}

func createBasicTableDataSourceTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DynamoDBDataSourceFetchTestCase {
	tableName := "test-table"
	tableArn := "arn:aws:dynamodb:us-west-2:123456789012:table/test-table"
	tableId := "abc123"

	return DynamoDBDataSourceFetchTestCase{
		Name: "successfully fetches basic table data",
		ServiceFactory: dynamodbmock.CreateDynamoDBServiceMockFactory(
			dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
				Table: &types.TableDescription{
					TableName:   aws.String(tableName),
					TableArn:    aws.String(tableArn),
					TableId:     aws.String(tableId),
					TableStatus: types.TableStatusActive,
					KeySchema: []types.KeySchemaElement{
						{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
					},
					AttributeDefinitions: []types.AttributeDefinition{
						{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
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
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("tableName", tableName),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"tableName":   core.MappingNodeFromString(tableName),
				"arn":         core.MappingNodeFromString(tableArn),
				"tableId":     core.MappingNodeFromString(tableId),
				"tableStatus": core.MappingNodeFromString("ACTIVE"),
				"keySchema": core.MappingNodeItems(
					core.MappingNodeFields(
						"attributeName", core.MappingNodeFromString("id"),
						"keyType", core.MappingNodeFromString("HASH"),
					),
				),
				"attributeDefinitions": core.MappingNodeItems(
					core.MappingNodeFields(
						"attributeName", core.MappingNodeFromString("id"),
						"attributeType", core.MappingNodeFromString("S"),
					),
				),
				"timeToLiveSpecification.enabled": core.MappingNodeFromBool(false),
				"pointInTimeRecoverySpecification.pointInTimeRecoveryEnabled": core.MappingNodeFromBool(false),
			},
		},
		ExpectError: false,
	}
}

func createTableDataSourceWithStreamsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DynamoDBDataSourceFetchTestCase {
	tableName := "stream-table"
	tableArn := "arn:aws:dynamodb:us-west-2:123456789012:table/stream-table"
	streamArn := "arn:aws:dynamodb:us-west-2:123456789012:table/stream-table/stream/2024-01-01T00:00:00.000"
	streamLabel := "2024-01-01T00:00:00.000"

	return DynamoDBDataSourceFetchTestCase{
		Name: "successfully fetches table with streams enabled",
		ServiceFactory: dynamodbmock.CreateDynamoDBServiceMockFactory(
			dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
				Table: &types.TableDescription{
					TableName:   aws.String(tableName),
					TableArn:    aws.String(tableArn),
					TableId:     aws.String("stream123"),
					TableStatus: types.TableStatusActive,
					StreamSpecification: &types.StreamSpecification{
						StreamEnabled:  aws.Bool(true),
						StreamViewType: types.StreamViewTypeNewAndOldImages,
					},
					LatestStreamArn:   aws.String(streamArn),
					LatestStreamLabel: aws.String(streamLabel),
					KeySchema: []types.KeySchemaElement{
						{AttributeName: aws.String("pk"), KeyType: types.KeyTypeHash},
					},
					AttributeDefinitions: []types.AttributeDefinition{
						{AttributeName: aws.String("pk"), AttributeType: types.ScalarAttributeTypeS},
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
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("tableName", tableName),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"tableName":                         core.MappingNodeFromString(tableName),
				"arn":                               core.MappingNodeFromString(tableArn),
				"tableId":                           core.MappingNodeFromString("stream123"),
				"tableStatus":                       core.MappingNodeFromString("ACTIVE"),
				"streamSpecification.streamEnabled": core.MappingNodeFromBool(true),
				"streamSpecification.streamViewType": core.MappingNodeFromString("NEW_AND_OLD_IMAGES"),
				"latestStreamArn":                   core.MappingNodeFromString(streamArn),
				"latestStreamLabel":                 core.MappingNodeFromString(streamLabel),
				"keySchema": core.MappingNodeItems(
					core.MappingNodeFields(
						"attributeName", core.MappingNodeFromString("pk"),
						"keyType", core.MappingNodeFromString("HASH"),
					),
				),
				"attributeDefinitions": core.MappingNodeItems(
					core.MappingNodeFields(
						"attributeName", core.MappingNodeFromString("pk"),
						"attributeType", core.MappingNodeFromString("S"),
					),
				),
				"timeToLiveSpecification.enabled": core.MappingNodeFromBool(false),
				"pointInTimeRecoverySpecification.pointInTimeRecoveryEnabled": core.MappingNodeFromBool(false),
			},
		},
		ExpectError: false,
	}
}

func createTableDataSourceWithTTLTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DynamoDBDataSourceFetchTestCase {
	tableName := "ttl-table"

	return DynamoDBDataSourceFetchTestCase{
		Name: "successfully fetches table with TTL enabled",
		ServiceFactory: dynamodbmock.CreateDynamoDBServiceMockFactory(
			dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
				Table: &types.TableDescription{
					TableName:   aws.String(tableName),
					TableArn:    aws.String("arn:aws:dynamodb:us-west-2:123456789012:table/ttl-table"),
					TableId:     aws.String("ttl123"),
					TableStatus: types.TableStatusActive,
					KeySchema: []types.KeySchemaElement{
						{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
					},
					AttributeDefinitions: []types.AttributeDefinition{
						{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
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
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("tableName", tableName),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"tableName":   core.MappingNodeFromString(tableName),
				"arn":         core.MappingNodeFromString("arn:aws:dynamodb:us-west-2:123456789012:table/ttl-table"),
				"tableId":     core.MappingNodeFromString("ttl123"),
				"tableStatus": core.MappingNodeFromString("ACTIVE"),
				"keySchema": core.MappingNodeItems(
					core.MappingNodeFields(
						"attributeName", core.MappingNodeFromString("id"),
						"keyType", core.MappingNodeFromString("HASH"),
					),
				),
				"attributeDefinitions": core.MappingNodeItems(
					core.MappingNodeFields(
						"attributeName", core.MappingNodeFromString("id"),
						"attributeType", core.MappingNodeFromString("S"),
					),
				),
				"timeToLiveSpecification.enabled":       core.MappingNodeFromBool(true),
				"timeToLiveSpecification.attributeName": core.MappingNodeFromString("expiresAt"),
				"pointInTimeRecoverySpecification.pointInTimeRecoveryEnabled": core.MappingNodeFromBool(false),
			},
		},
		ExpectError: false,
	}
}

func createTableDataSourceWithPITRTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DynamoDBDataSourceFetchTestCase {
	tableName := "pitr-table"

	return DynamoDBDataSourceFetchTestCase{
		Name: "successfully fetches table with PITR enabled",
		ServiceFactory: dynamodbmock.CreateDynamoDBServiceMockFactory(
			dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
				Table: &types.TableDescription{
					TableName:   aws.String(tableName),
					TableArn:    aws.String("arn:aws:dynamodb:us-west-2:123456789012:table/pitr-table"),
					TableId:     aws.String("pitr123"),
					TableStatus: types.TableStatusActive,
					KeySchema: []types.KeySchemaElement{
						{AttributeName: aws.String("id"), KeyType: types.KeyTypeHash},
					},
					AttributeDefinitions: []types.AttributeDefinition{
						{AttributeName: aws.String("id"), AttributeType: types.ScalarAttributeTypeS},
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
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("tableName", tableName),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"tableName":   core.MappingNodeFromString(tableName),
				"arn":         core.MappingNodeFromString("arn:aws:dynamodb:us-west-2:123456789012:table/pitr-table"),
				"tableId":     core.MappingNodeFromString("pitr123"),
				"tableStatus": core.MappingNodeFromString("ACTIVE"),
				"keySchema": core.MappingNodeItems(
					core.MappingNodeFields(
						"attributeName", core.MappingNodeFromString("id"),
						"keyType", core.MappingNodeFromString("HASH"),
					),
				),
				"attributeDefinitions": core.MappingNodeItems(
					core.MappingNodeFields(
						"attributeName", core.MappingNodeFromString("id"),
						"attributeType", core.MappingNodeFromString("S"),
					),
				),
				"timeToLiveSpecification.enabled": core.MappingNodeFromBool(false),
				"pointInTimeRecoverySpecification.pointInTimeRecoveryEnabled": core.MappingNodeFromBool(true),
			},
		},
		ExpectError: false,
	}
}

func createTableDataSourceNotFoundTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DynamoDBDataSourceFetchTestCase {
	return DynamoDBDataSourceFetchTestCase{
		Name: "returns error when table not found",
		ServiceFactory: dynamodbmock.CreateDynamoDBServiceMockFactory(
			dynamodbmock.WithDescribeTableError(&types.ResourceNotFoundException{
				Message: aws.String("Requested resource not found: Table: nonexistent-table not found"),
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("tableName", "nonexistent-table"),
			},
		},
		ExpectedOutput:       nil,
		ExpectError:          true,
		ExpectedErrorMessage: "not found",
	}
}

func createTableDataSourceFetchErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DynamoDBDataSourceFetchTestCase {
	return DynamoDBDataSourceFetchTestCase{
		Name: "returns error on describe table failure",
		ServiceFactory: dynamodbmock.CreateDynamoDBServiceMockFactory(
			dynamodbmock.WithDescribeTableError(errors.New("service unavailable")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("tableName", "error-table"),
			},
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func createTableDataSourceMissingFilterTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DynamoDBDataSourceFetchTestCase {
	return DynamoDBDataSourceFetchTestCase{
		Name: "returns error when table name or ARN filter is missing",
		ServiceFactory: dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: &provider.ResolvedDataSourceFilters{
					Filters: []*provider.ResolvedDataSourceFilter{},
				},
			},
		},
		ExpectedOutput:       nil,
		ExpectError:          true,
		ExpectedErrorMessage: "table name or ARN is required",
	}
}
