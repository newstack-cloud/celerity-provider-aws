//go:build unit

package lambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaLayerVersionPermissionsResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *LambdaLayerVersionPermissionsResourceGetExternalStateSuite) Test_get_external_state_lambda_layer_version_permissions() {
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

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		getExternalStateBasicLayerVersionPermissionTestCase(providerCtx, loader),
		getExternalStateLayerVersionPermissionNotFoundTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		LayerVersionPermissionResource,
		&s.Suite,
	)
}

func getExternalStateBasicLayerVersionPermissionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	// Create test data for layer version permission get external state
	currentResourceSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"),
			"statementId":     core.MappingNodeFromString("test-statement"),
			"action":          core.MappingNodeFromString("lambda:GetLayerVersion"),
			"principal":       core.MappingNodeFromString("123456789012"),
		},
	}

	// Expected output with computed ID
	expectedResourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"),
			"statementId":     core.MappingNodeFromString("test-statement"),
			"action":          core.MappingNodeFromString("lambda:GetLayerVersion"),
			"principal":       core.MappingNodeFromString("123456789012"),
			"id":              core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1#test-statement"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "get external state basic layer version permission",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetLayerVersionPolicyOutput(&lambda.GetLayerVersionPolicyOutput{
				Policy:     aws.String(`{"Version":"2012-10-17","Statement":{"Sid":"test-statement","Effect":"Allow","Principal":"123456789012","Action":"lambda:GetLayerVersion"}}`),
				RevisionId: aws.String("revision-123"),
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-layer-version-permission-id",
			CurrentResourceSpec: currentResourceSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: expectedResourceSpecState,
		},
	}
}

func getExternalStateLayerVersionPermissionNotFoundTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "get external state layer version permission not found",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetLayerVersionPolicyError(&smithy.GenericAPIError{
				Code:    "ResourceNotFoundException",
				Message: "Layer version not found",
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-layer-version-permission-id",
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"layerVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:nonexistent-layer:1"),
					"statementId":     core.MappingNodeFromString("test-statement"),
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{Fields: make(map[string]*core.MappingNode)},
		},
	}
}

func TestLambdaLayerVersionPermissionsResourceGetExternalState(t *testing.T) {
	suite.Run(t, new(LambdaLayerVersionPermissionsResourceGetExternalStateSuite))
}
