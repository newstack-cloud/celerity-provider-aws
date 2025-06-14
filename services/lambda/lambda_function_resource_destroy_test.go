package lambda

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity-provider-aws/internal/testutils"
	"github.com/newstack-cloud/celerity-provider-aws/utils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/blueprint/state"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/plugintestutils"
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

	testCases := []plugintestutils.ResourceDestroyTestCase[*aws.Config, Service]{
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
) plugintestutils.ResourceDestroyTestCase[*aws.Config, Service] {
	service := createLambdaServiceMock(
		WithDeleteFunctionOutput(&lambda.DeleteFunctionOutput{}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, Service]{
		Name: "successfully deletes function",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		Input: &provider.ResourceDestroyInput{
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
		ExpectError: false,
	}
}

func createFailingDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, Service] {
	service := createLambdaServiceMock(
		WithDeleteFunctionError(errors.New("failed to delete function")),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, Service]{
		Name: "fails to delete function",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		Input: &provider.ResourceDestroyInput{
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
		ExpectError: true,
	}
}

func TestLambdaFunctionResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionResourceDestroySuite))
}
