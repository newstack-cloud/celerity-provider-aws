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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type SSMParameterPathResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *SSMParameterPathResourceGetExternalStateSuite) Test_get_external_state() {
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

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, ssmservice.Service]{
		{
			Name:           "echoes the path from the current resource spec",
			ServiceFactory: func(*aws.Config, provider.Context) ssmservice.Service { return service },
			ConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
				utils.AWSConfigCacheKey,
			),
			Input: &provider.ResourceGetExternalStateInput{
				ProviderContext:     providerCtx,
				CurrentResourceSpec: parameterPathSpecData(),
			},
			ExpectedOutput: &provider.ResourceGetExternalStateOutput{
				ResourceSpecState: parameterPathSpecData(),
			},
			ExpectError: false,
		},
		{
			Name:           "returns error when path is missing",
			ServiceFactory: func(*aws.Config, provider.Context) ssmservice.Service { return service },
			ConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
				utils.AWSConfigCacheKey,
			),
			Input: &provider.ResourceGetExternalStateInput{
				ProviderContext: providerCtx,
				CurrentResourceSpec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{},
				},
			},
			ExpectError: true,
		},
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		ParameterPathResource,
		&s.Suite,
	)
}

func TestSSMParameterPathResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterPathResourceGetExternalStateSuite))
}
