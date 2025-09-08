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

type LambdaFunctionResourceStabilisedSuite struct {
	suite.Suite
}

func (s *LambdaFunctionResourceStabilisedSuite) Test_stabilised() {
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
		{
			Name: "returns stabilised when function is successfully updated",
			ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
				lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
					Configuration: &types.FunctionConfiguration{
						State: types.StateActive,
					},
				}),
			),
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
						"arn": core.MappingNodeFromString(
							"arn:aws:lambda:us-east-1:123456789012:function:test-function",
						),
					},
				},
			},
			ExpectedOutput: &provider.ResourceHasStabilisedOutput{
				Stabilised: true,
			},
			ExpectError: false,
		},
		{
			Name: "returns not stabilised when function is still updating",
			ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
				lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
					Configuration: &types.FunctionConfiguration{
						State: types.StatePending,
					},
				}),
			),
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
						"arn": core.MappingNodeFromString(
							"arn:aws:lambda:us-east-1:123456789012:function:test-function",
						),
					},
				},
			},
			ExpectedOutput: &provider.ResourceHasStabilisedOutput{
				Stabilised: false,
			},
			ExpectError: false,
		},
		{
			Name: "handles get function error",
			ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
				lambdamock.WithGetFunctionError(errors.New("failed to get function")),
			),
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
						"arn": core.MappingNodeFromString(
							"arn:aws:lambda:us-east-1:123456789012:function:test-function",
						),
					},
				},
			},
			ExpectedOutput: nil,
			ExpectError:    true,
		},
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		FunctionResource,
		&s.Suite,
	)
}

func TestLambdaFunctionResourceStabilisedSuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionResourceStabilisedSuite))
}
