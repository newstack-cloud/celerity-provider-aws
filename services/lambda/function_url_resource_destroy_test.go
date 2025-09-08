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

type LambdaFunctionUrlResourceDestroySuite struct {
	suite.Suite
}

func (s *LambdaFunctionUrlResourceDestroySuite) Test_destroy() {
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
		destroyBasicFunctionUrlTestCase(providerCtx, loader),
		destroyFunctionUrlWithQualifierTestCase(providerCtx, loader),
		destroyFunctionUrlFailureTestCase(providerCtx, loader),
		destroyFunctionUrlWithNoIDTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		FunctionUrlResource,
		&s.Suite,
	)
}

func destroyBasicFunctionUrlTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteFunctionUrlConfigOutput(&lambda.DeleteFunctionUrlConfigOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully deletes function URL",
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
			"DeleteFunctionUrlConfig": &lambda.DeleteFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			},
		},
	}
}

func destroyFunctionUrlWithQualifierTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteFunctionUrlConfigOutput(&lambda.DeleteFunctionUrlConfigOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"qualifier":   core.MappingNodeFromString("PROD"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully deletes function URL with qualifier",
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
			"DeleteFunctionUrlConfig": &lambda.DeleteFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				Qualifier:    aws.String("PROD"),
			},
		},
	}
}

func destroyFunctionUrlFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteFunctionUrlConfigError(fmt.Errorf("function URL not found")),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "fails to delete function URL",
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
			"DeleteFunctionUrlConfig": &lambda.DeleteFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			},
		},
	}
}

func destroyFunctionUrlWithNoIDTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteFunctionUrlConfigOutput(&lambda.DeleteFunctionUrlConfigOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{},
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, lambdaservice.Service]{
		Name: "handles destroy with no function arn without panicking",
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
		ExpectError:          true,
		DestroyActionsCalled: map[string]any{},
	}
}

func TestLambdaFunctionUrlResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionUrlResourceDestroySuite))
}
