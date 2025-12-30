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

type DynamoDBTableResourceStabilisedSuite struct {
	suite.Suite
}

func (s *DynamoDBTableResourceStabilisedSuite) Test_stabilised_dynamodb_table() {
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

	testCases := []plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, dynamodbservice.Service]{
		stabilisedTableActiveTestCase(providerCtx, loader),
		stabilisedTableCreatingTestCase(providerCtx, loader),
		stabilisedTableUpdatingTestCase(providerCtx, loader),
		stabilisedTableWithActiveGSITestCase(providerCtx, loader),
		stabilisedTableWithCreatingGSITestCase(providerCtx, loader),
		stabilisedTableErrorTestCase(providerCtx, loader),
	}

	tableResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, dynamodbservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return TableResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		tableResourceWrapper,
		&s.Suite,
	)
}

func stabilisedTableActiveTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, dynamodbservice.Service] {
	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
			Table: &types.TableDescription{
				TableName:   aws.String("test-table"),
				TableStatus: types.TableStatusActive,
			},
		}),
	)

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("test-table"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "table in ACTIVE status returns stabilised=true",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpecState,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func stabilisedTableCreatingTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, dynamodbservice.Service] {
	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
			Table: &types.TableDescription{
				TableName:   aws.String("test-table"),
				TableStatus: types.TableStatusCreating,
			},
		}),
	)

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("test-table"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "table in CREATING status returns stabilised=false",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpecState,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: false,
		},
	}
}

func stabilisedTableUpdatingTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, dynamodbservice.Service] {
	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
			Table: &types.TableDescription{
				TableName:   aws.String("test-table"),
				TableStatus: types.TableStatusUpdating,
			},
		}),
	)

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("test-table"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "table in UPDATING status returns stabilised=false",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpecState,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: false,
		},
	}
}

func stabilisedTableWithActiveGSITestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, dynamodbservice.Service] {
	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
			Table: &types.TableDescription{
				TableName:   aws.String("test-table"),
				TableStatus: types.TableStatusActive,
				GlobalSecondaryIndexes: []types.GlobalSecondaryIndexDescription{
					{
						IndexName:   aws.String("gsi-index"),
						IndexStatus: types.IndexStatusActive,
					},
				},
			},
		}),
	)

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("test-table"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "table with ACTIVE GSI returns stabilised=true",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpecState,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func stabilisedTableWithCreatingGSITestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, dynamodbservice.Service] {
	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithDescribeTableOutput(&dynamodb.DescribeTableOutput{
			Table: &types.TableDescription{
				TableName:   aws.String("test-table"),
				TableStatus: types.TableStatusActive,
				GlobalSecondaryIndexes: []types.GlobalSecondaryIndexDescription{
					{
						IndexName:   aws.String("gsi-index"),
						IndexStatus: types.IndexStatusCreating,
					},
				},
			},
		}),
	)

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("test-table"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "table with CREATING GSI returns stabilised=false",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpecState,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: false,
		},
	}
}

func stabilisedTableErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, dynamodbservice.Service] {
	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithDescribeTableError(errors.New("failed to describe table")),
	)

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"tableName": core.MappingNodeFromString("test-table"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "describe table error returns error",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpecState,
		},
		ExpectError: true,
	}
}

func TestDynamoDBTableResourceStabilisedSuite(t *testing.T) {
	suite.Run(t, new(DynamoDBTableResourceStabilisedSuite))
}
