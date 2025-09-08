//go:build unit

package lambda

import (
	"errors"
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

type LambdaFunctionResourceDestroySuite struct {
	suite.Suite
}

func (s *LambdaFunctionResourceDestroySuite) Test_destroy() {
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
		createSuccesfulDestroyTestCase(providerCtx, loader),
		createFailingDestroyTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		FunctionResource,
		&s.Suite,
	)
}

func createSuccesfulDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteFunctionOutput(&lambda.DeleteFunctionOutput{}),
	)

	expectedFunctionARN := "arn:aws:lambda:us-east-1:123456789012:function:test-function"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully deletes function",
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
						"arn": core.MappingNodeFromString(expectedFunctionARN),
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"DeleteFunction": &lambda.DeleteFunctionInput{
				FunctionName: aws.String(expectedFunctionARN),
			},
		},
	}
}

func createFailingDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteFunctionError(errors.New("failed to delete function")),
	)

	expectedFunctionARN := "arn:aws:lambda:us-east-1:123456789012:function:test-function"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "fails to delete function",
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
						"arn": core.MappingNodeFromString(expectedFunctionARN),
					},
				},
			},
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteFunction": &lambda.DeleteFunctionInput{
				FunctionName: aws.String(expectedFunctionARN),
			},
		},
	}
}

func TestLambdaFunctionResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionResourceDestroySuite))
}
