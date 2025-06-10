package lambda

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/celerity-provider-aws/internal/testutils"
	"github.com/newstack-cloud/celerity-provider-aws/utils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type LambdaFunctionResourceStabilisedSuite struct {
	suite.Suite
}

type stabilisedTestCase struct {
	name                 string
	lambdaServiceFactory func(awsConfig *aws.Config, providerContext provider.Context) Service
	awsConfigStore       *utils.AWSConfigStore
	input                *provider.ResourceHasStabilisedInput
	expectedOutput       *provider.ResourceHasStabilisedOutput
	expectError          bool
}

func (s *LambdaFunctionResourceStabilisedSuite) Test_stabilised() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := testutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []stabilisedTestCase{
		{
			name: "returns stabilised when function is successfully updated",
			lambdaServiceFactory: createLambdaServiceMockFactory(
				WithGetFunctionOutput(&lambda.GetFunctionOutput{
					Configuration: &types.FunctionConfiguration{
						LastUpdateStatus: types.LastUpdateStatusSuccessful,
					},
				}),
			),
			awsConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
			),
			input: &provider.ResourceHasStabilisedInput{
				ProviderContext: providerCtx,
				ResourceSpec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn": core.MappingNodeFromString(
							"arn:aws:lambda:us-east-1:123456789012:function:test-function",
						),
					},
				},
			},
			expectedOutput: &provider.ResourceHasStabilisedOutput{
				Stabilised: true,
			},
			expectError: false,
		},
		{
			name: "returns not stabilised when function is still updating",
			lambdaServiceFactory: createLambdaServiceMockFactory(
				WithGetFunctionOutput(&lambda.GetFunctionOutput{
					Configuration: &types.FunctionConfiguration{
						LastUpdateStatus: types.LastUpdateStatusInProgress,
					},
				}),
			),
			awsConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
			),
			input: &provider.ResourceHasStabilisedInput{
				ProviderContext: providerCtx,
				ResourceSpec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn": core.MappingNodeFromString(
							"arn:aws:lambda:us-east-1:123456789012:function:test-function",
						),
					},
				},
			},
			expectedOutput: &provider.ResourceHasStabilisedOutput{
				Stabilised: false,
			},
			expectError: false,
		},
		{
			name: "handles get function error",
			lambdaServiceFactory: createLambdaServiceMockFactory(
				WithGetFunctionError(errors.New("failed to get function")),
			),
			awsConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
			),
			input: &provider.ResourceHasStabilisedInput{
				ProviderContext: providerCtx,
				ResourceSpec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn": core.MappingNodeFromString(
							"arn:aws:lambda:us-east-1:123456789012:function:test-function",
						),
					},
				},
			},
			expectedOutput: nil,
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			lambdaActions := lambdaFunctionResourceActions{
				lambdaServiceFactory: tc.lambdaServiceFactory,
				awsConfigStore:       tc.awsConfigStore,
			}

			output, err := lambdaActions.StabilisedFunc(context.Background(), tc.input)

			if tc.expectError {
				s.Error(err)
				s.Nil(output)
			} else {
				s.NoError(err)
				s.Equal(tc.expectedOutput, output)
			}
		})
	}
}

func TestLambdaFunctionResourceStabilisedSuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionResourceStabilisedSuite))
}
