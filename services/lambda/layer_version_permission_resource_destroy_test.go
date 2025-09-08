//go:build unit

package lambda

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaLayerVersionPermissionsResourceDestroySuite struct {
	suite.Suite
}

func (s *LambdaLayerVersionPermissionsResourceDestroySuite) Test_destroy_lambda_layer_version_permissions() {
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

	testCases := []plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		destroyBasicLayerVersionPermissionTestCase(providerCtx, loader),
		destroyLayerVersionPermissionFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		LayerVersionPermissionResource,
		&s.Suite,
	)
}

func destroyBasicLayerVersionPermissionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithRemoveLayerVersionPermissionOutput(&lambda.RemoveLayerVersionPermissionOutput{}),
	)

	// Create test data for layer version permission destruction
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"),
			"statementId":     core.MappingNodeFromString("test-statement"),
			"action":          core.MappingNodeFromString("lambda:GetLayerVersion"),
			"principal":       core.MappingNodeFromString("123456789012"),
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "destroy basic layer version permission",
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
		Input: &provider.ResourceDestroyInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-layer-version-permission-id",
			ResourceState: &state.ResourceState{
				SpecData: specData,
			},
			ProviderContext: providerCtx,
		},
		DestroyActionsCalled: map[string]any{
			"RemoveLayerVersionPermission": &lambda.RemoveLayerVersionPermissionInput{
				LayerName:     aws.String("test-layer"),
				VersionNumber: aws.Int64(1),
				StatementId:   aws.String("test-statement"),
			},
		},
	}
}

func destroyLayerVersionPermissionFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithRemoveLayerVersionPermissionError(fmt.Errorf("failed to remove layer version permission")),
	)

	// Create test data for layer version permission destruction
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"),
			"statementId":     core.MappingNodeFromString("test-statement"),
			"action":          core.MappingNodeFromString("lambda:GetLayerVersion"),
			"principal":       core.MappingNodeFromString("123456789012"),
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "destroy layer version permission failure",
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
		Input: &provider.ResourceDestroyInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-layer-version-permission-id",
			ResourceState: &state.ResourceState{
				SpecData: specData,
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"RemoveLayerVersionPermission": &lambda.RemoveLayerVersionPermissionInput{
				LayerName:     aws.String("test-layer"),
				VersionNumber: aws.Int64(1),
				StatementId:   aws.String("test-statement"),
			},
		},
	}
}

func TestLambdaLayerVersionPermissionsResourceDestroy(t *testing.T) {
	suite.Run(t, new(LambdaLayerVersionPermissionsResourceDestroySuite))
}
