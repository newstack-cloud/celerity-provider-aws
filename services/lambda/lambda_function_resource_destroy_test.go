package lambda

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity-provider-aws/internal/testutils"
	"github.com/newstack-cloud/celerity-provider-aws/utils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/blueprint/state"
	"github.com/stretchr/testify/suite"
)

type LambdaFunctionResourceDestroySuite struct {
	suite.Suite
}

func (s *LambdaFunctionResourceDestroySuite) Test_destroy() {
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

	testCases := []destroyTestCase{
		{
			name: "successfully deletes function",
			lambdaServiceFactory: createLambdaServiceMockFactory(
				WithDeleteFunctionOutput(&lambda.DeleteFunctionOutput{}),
			),
			awsConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
			),
			input: &provider.ResourceDestroyInput{
				ProviderContext: providerCtx,
				ResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(
								"arn:aws:lambda:us-east-1:123456789012:function:test-function",
							),
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "handles delete function error",
			lambdaServiceFactory: createLambdaServiceMockFactory(
				WithDeleteFunctionError(errors.New("failed to delete function")),
			),
			awsConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
			),
			input: &provider.ResourceDestroyInput{
				ProviderContext: providerCtx,
				ResourceState: &state.ResourceState{
					SpecData: &core.MappingNode{
						Fields: map[string]*core.MappingNode{
							"arn": core.MappingNodeFromString(
								"arn:aws:lambda:us-east-1:123456789012:function:test-function",
							),
						},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			lambdaActions := lambdaFunctionResourceActions{
				lambdaServiceFactory: tc.lambdaServiceFactory,
				awsConfigStore:       tc.awsConfigStore,
			}

			err := lambdaActions.DestroyFunc(context.Background(), tc.input)

			if tc.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
			}
		})
	}
}

func TestLambdaFunctionResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionResourceDestroySuite))
}
