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
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type SSMParameterPathResourceUpdateSuite struct {
	suite.Suite
}

func (s *SSMParameterPathResourceUpdateSuite) Test_update() {
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

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		{
			Name:             "updates the parameter path without any AWS calls",
			ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
			ServiceMockCalls: &service.MockCalls,
			ConfigStore: utils.NewAWSConfigStore(
				[]string{},
				utils.AWSConfigFromProviderContext,
				loader,
				utils.AWSConfigCacheKey,
			),
			Input: parameterPathUpdateInput(providerCtx),
			ExpectedOutput: &provider.ResourceDeployOutput{
				ComputedFieldValues: map[string]*core.MappingNode{},
			},
			SaveActionsNotCalled: allSSMServiceMethods,
			ExpectError:          false,
		},
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		ParameterPathResource,
		&s.Suite,
	)
}

func parameterPathUpdateInput(providerCtx provider.Context) *provider.ResourceDeployInput {
	return &provider.ResourceDeployInput{
		InstanceID: "test-instance-id",
		ResourceID: "test-resource-id",
		Changes: &provider.Changes{
			AppliedResourceInfo: provider.ResourceInfo{
				ResourceID:   "test-resource-id",
				ResourceName: "TestParameterPath",
				InstanceID:   "test-instance-id",
				CurrentResourceState: &state.ResourceState{
					ResourceID: "test-resource-id",
					Name:       "TestParameterPath",
					InstanceID: "test-instance-id",
					SpecData:   parameterPathSpecData(),
				},
				ResourceWithResolvedSubs: &provider.ResolvedResource{
					Type: &schema.ResourceTypeWrapper{
						Value: "aws/ssm/parameterPath",
					},
					Spec: parameterPathSpecData(),
				},
			},
		},
		ProviderContext: providerCtx,
	}
}

func TestSSMParameterPathResourceUpdateSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterPathResourceUpdateSuite))
}
