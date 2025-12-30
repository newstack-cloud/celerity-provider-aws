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
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type DynamoDBTableResourceDestroySuite struct {
	suite.Suite
}

func (s *DynamoDBTableResourceDestroySuite) Test_destroy_dynamodb_table() {
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

	testCases := []plugintestutils.ResourceDestroyTestCase[*aws.Config, dynamodbservice.Service]{
		destroyTableSuccessTestCase(providerCtx, loader),
		destroyTableFailureTestCase(providerCtx, loader),
	}

	tableResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, dynamodbservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return TableResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		tableResourceWrapper,
		&s.Suite,
	)
}

func destroyTableSuccessTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, dynamodbservice.Service] {
	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithDeleteTableOutput(&dynamodb.DeleteTableOutput{
			TableDescription: &types.TableDescription{
				TableName:   aws.String("test-table"),
				TableStatus: types.TableStatusDeleting,
			},
		}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "destroy DynamoDB table successfully",
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
		Input: &provider.ResourceDestroyInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-table-id",
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"tableName": core.MappingNodeFromString("test-table"),
						"arn":       core.MappingNodeFromString("arn:aws:dynamodb:us-west-2:123456789012:table/test-table"),
					},
				},
			},
			ProviderContext: providerCtx,
		},
		DestroyActionsCalled: map[string]any{
			"DeleteTable": &dynamodb.DeleteTableInput{
				TableName: aws.String("test-table"),
			},
		},
	}
}

func destroyTableFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, dynamodbservice.Service] {
	service := dynamodbmock.CreateDynamoDBServiceMock(
		dynamodbmock.WithDeleteTableError(fmt.Errorf("failed to delete table")),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, dynamodbservice.Service]{
		Name: "destroy DynamoDB table failure",
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
		Input: &provider.ResourceDestroyInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-table-id",
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"tableName": core.MappingNodeFromString("test-table"),
						"arn":       core.MappingNodeFromString("arn:aws:dynamodb:us-west-2:123456789012:table/test-table"),
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteTable": &dynamodb.DeleteTableInput{
				TableName: aws.String("test-table"),
			},
		},
	}
}

func TestDynamoDBTableResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(DynamoDBTableResourceDestroySuite))
}
