//go:build unit

package lambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaFunctionUrlResourceStabilisedSuite struct {
	suite.Suite
}

func (s *LambdaFunctionUrlResourceStabilisedSuite) Test_stabilised() {
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

	testCases := []plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service]{
		stabilisedFunctionUrlExistsTestCase(providerCtx, loader),
		stabilisedFunctionUrlNotExistsTestCase(providerCtx, loader),
		stabilisedFunctionUrlWithQualifierTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		FunctionUrlResource,
		&s.Suite,
	)
}

func stabilisedFunctionUrlExistsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{
			FunctionUrl: aws.String("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			AuthType:    types.FunctionUrlAuthTypeNone,
		}),
	)

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service]{
		Name: "function URL exists and is stabilised",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
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

func stabilisedFunctionUrlNotExistsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionUrlConfigError(&types.ResourceNotFoundException{
			Message: aws.String("Function URL not found"),
		}),
	)

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service]{
		Name: "function URL does not exist and is not stabilised",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
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

func stabilisedFunctionUrlWithQualifierTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{
			FunctionUrl: aws.String("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			AuthType:    types.FunctionUrlAuthTypeAwsIam,
		}),
	)

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"qualifier":   core.MappingNodeFromString("PROD"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service]{
		Name: "function URL with qualifier exists and is stabilised",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
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

func TestLambdaFunctionUrlResourceStabilised(t *testing.T) {
	suite.Run(t, new(LambdaFunctionUrlResourceStabilisedSuite))
}
