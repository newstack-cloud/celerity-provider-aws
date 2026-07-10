//go:build unit

package ssm

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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

type SSMParameterPathResourceDestroySuite struct {
	suite.Suite
}

func (s *SSMParameterPathResourceDestroySuite) Test_destroy() {
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

	service := ssmmock.CreateSSMServiceMock()

	testCases := []plugintestutils.ResourceDestroyTestCase[*aws.Config, ssmservice.Service]{
		{
			Name:             "destroys the parameter path without any AWS calls",
			ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
			ServiceMockCalls: &service.MockCalls,
			ConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
				utils.AWSConfigCacheKey,
			),
			Input: &provider.ResourceDestroyInput{
				InstanceID: "test-instance-id",
				ResourceID: "test-resource-id",
				ResourceState: &state.ResourceState{
					ResourceID: "test-resource-id",
					Name:       "TestParameterPath",
					InstanceID: "test-instance-id",
					SpecData:   parameterPathSpecData(),
				},
				ProviderContext: providerCtx,
			},
			ExpectError:             false,
			DestroyActionsNotCalled: allSSMServiceMethods,
		},
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		ParameterPathResource,
		&s.Suite,
	)
}

func TestSSMParameterPathResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(SSMParameterPathResourceDestroySuite))
}
