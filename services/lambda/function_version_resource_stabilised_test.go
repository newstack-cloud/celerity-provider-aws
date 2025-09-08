//go:build unit

package lambda

import (
	"errors"
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

type LambdaFunctionVersionResourceStabilisedSuite struct {
	suite.Suite
}

func (s *LambdaFunctionVersionResourceStabilisedSuite) Test_stabilised() {
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
		createSuccessfulFunctionVersionStabilisedTestCase(providerCtx, loader),
		createFailingFunctionVersionStabilisedTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		FunctionVersionResource,
		&s.Suite,
	)
}

func createSuccessfulFunctionVersionStabilisedTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &types.FunctionConfiguration{
				State: types.StateActive,
			},
		}),
	)

	expectedFunctionARN := "arn:aws:lambda:us-east-1:123456789012:function:test-function"
	expectedVersion := "1"

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully stabilises function version",
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
			ResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionArn": core.MappingNodeFromString(expectedFunctionARN),
					"version":     core.MappingNodeFromString(expectedVersion),
				},
			},
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
		ExpectError: false,
	}
}

func createFailingFunctionVersionStabilisedTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionError(errors.New("failed to get function")),
	)

	expectedFunctionARN := "arn:aws:lambda:us-east-1:123456789012:function:test-function"
	expectedVersion := "1"

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, lambdaservice.Service]{
		Name: "fails to stabilise function version",
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
			ResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionArn": core.MappingNodeFromString(expectedFunctionARN),
					"version":     core.MappingNodeFromString(expectedVersion),
				},
			},
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func TestLambdaFunctionVersionResourceStabilisedSuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionVersionResourceStabilisedSuite))
}
