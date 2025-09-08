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

type LambdaLayerVersionResourceDestroySuite struct {
	suite.Suite
}

func (s *LambdaLayerVersionResourceDestroySuite) Test_destroy_lambda_layer_version() {
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
		destroyLayerVersionSuccessTestCase(providerCtx, loader),
		destroyLayerVersionFailureTestCase(providerCtx, loader),
		destroyLayerVersionInvalidArnTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		LayerVersionResource,
		&s.Suite,
	)
}

func destroyLayerVersionSuccessTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	layerVersionArn := "arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteLayerVersionOutput(&lambda.DeleteLayerVersionOutput{}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "destroy layer version success",
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
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"layerVersionArn": core.MappingNodeFromString(layerVersionArn),
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"DeleteLayerVersion": &lambda.DeleteLayerVersionInput{
				LayerName:     aws.String("test-layer"),
				VersionNumber: aws.Int64(1),
			},
		},
	}
}

func destroyLayerVersionFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	layerVersionArn := "arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteLayerVersionError(fmt.Errorf("failed to delete layer version")),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "destroy layer version failure",
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
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"layerVersionArn": core.MappingNodeFromString(layerVersionArn),
					},
				},
			},
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteLayerVersion": &lambda.DeleteLayerVersionInput{
				LayerName:     aws.String("test-layer"),
				VersionNumber: aws.Int64(1),
			},
		},
	}
}

func destroyLayerVersionInvalidArnTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	invalidArn := "invalid-arn-format"

	service := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "destroy layer version with invalid ARN",
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
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"layerVersionArn": core.MappingNodeFromString(invalidArn),
					},
				},
			},
		},
		ExpectError: true,
		// No destroy actions should be called due to ARN parsing failure
	}
}

func TestLambdaLayerVersionResourceDestroy(t *testing.T) {
	suite.Run(t, new(LambdaLayerVersionResourceDestroySuite))
}
