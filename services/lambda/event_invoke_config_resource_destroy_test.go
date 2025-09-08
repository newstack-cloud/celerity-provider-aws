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

type LambdaEventInvokeConfigResourceDestroySuite struct {
	suite.Suite
}

func (s *LambdaEventInvokeConfigResourceDestroySuite) Test_destroy_lambda_event_invoke_config() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-east-1"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		destroyEventInvokeConfigTestCase(providerCtx, loader),
		destroyEventInvokeConfigFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		EventInvokeConfigResource,
		&s.Suite,
	)
}

func destroyEventInvokeConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteFunctionEventInvokeConfigOutput(&lambda.DeleteFunctionEventInvokeConfigOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"qualifier":    core.MappingNodeFromString("$LATEST"),
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "destroy event invoke config",
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
				SpecData: specData,
			},
		},
		DestroyActionsCalled: map[string]any{
			"DeleteFunctionEventInvokeConfig": &lambda.DeleteFunctionEventInvokeConfigInput{
				FunctionName: aws.String("test-function"),
				Qualifier:    aws.String("$LATEST"),
			},
		},
	}
}

func destroyEventInvokeConfigFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteFunctionEventInvokeConfigError(fmt.Errorf("failed to delete event invoke config")),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"qualifier":    core.MappingNodeFromString("$LATEST"),
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "destroy event invoke config failure",
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
				SpecData: specData,
			},
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteFunctionEventInvokeConfig": &lambda.DeleteFunctionEventInvokeConfigInput{
				FunctionName: aws.String("test-function"),
				Qualifier:    aws.String("$LATEST"),
			},
		},
	}
}

func TestLambdaEventInvokeConfigResourceDestroy(t *testing.T) {
	suite.Run(t, new(LambdaEventInvokeConfigResourceDestroySuite))
}
