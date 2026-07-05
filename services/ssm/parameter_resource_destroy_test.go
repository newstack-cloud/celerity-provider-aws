//go:build unit

package ssm

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ssmmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ssm_mock"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type SSMParameterResourceDestroySuite struct {
	suite.Suite
}

func (s *SSMParameterResourceDestroySuite) Test_destroy() {
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

	testCases := []plugintestutils.ResourceDestroyTestCase[*aws.Config, ssmservice.Service]{
		destroyParameterTestCase(providerCtx, loader),
		destroyParameterErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		ParameterResource,
		&s.Suite,
	)
}

func destroyParameterTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithDeleteParameterOutput(&ssm.DeleteParameterOutput{}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ssmservice.Service]{
		Name:             "deletes the parameter",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input:       destroyInput(providerCtx, currentParameterState()),
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"DeleteParameter": func(arg any) bool {
				in, ok := arg.(*ssm.DeleteParameterInput)
				return ok && aws.ToString(in.Name) == testParameterName
			},
		},
	}
}

func destroyParameterErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithDeleteParameterError(errTestDeleteParameter),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ssmservice.Service]{
		Name:             "returns error when DeleteParameter fails",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input:       destroyInput(providerCtx, currentParameterState()),
		ExpectError: true,
	}
}

func destroyInput(
	providerCtx provider.Context,
	currentSpec *core.MappingNode,
) *provider.ResourceDestroyInput {
	return &provider.ResourceDestroyInput{
		InstanceID: "test-instance-id",
		ResourceID: "test-resource-id",
		ResourceState: &state.ResourceState{
			ResourceID: "test-resource-id",
			Name:       "TestParameter",
			InstanceID: "test-instance-id",
			SpecData:   currentSpec,
		},
		ProviderContext: providerCtx,
	}
}

func TestSSMParameterResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(SSMParameterResourceDestroySuite))
}
