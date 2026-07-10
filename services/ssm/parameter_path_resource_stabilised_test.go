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

type SSMParameterPathResourceStabilisedSuite struct {
	suite.Suite
}

func (s *SSMParameterPathResourceStabilisedSuite) Test_stabilised() {
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

	testCases := []plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, ssmservice.Service]{
		{
			Name:           "a parameter path is always stabilised",
			ServiceFactory: func(*aws.Config, provider.Context) ssmservice.Service { return service },
			ConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
				utils.AWSConfigCacheKey,
			),
			Input: &provider.ResourceHasStabilisedInput{
				ProviderContext: providerCtx,
				ResourceSpec:    parameterPathSpecData(),
			},
			ExpectedOutput: &provider.ResourceHasStabilisedOutput{
				Stabilised: true,
			},
			ExpectError: false,
		},
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		ParameterPathResource,
		&s.Suite,
	)
}

func TestSSMParameterPathResourceStabilisedSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterPathResourceStabilisedSuite))
}
