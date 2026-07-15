//go:build unit

package ssm

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ssmmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ssm_mock"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type SSMParameterTreeResourceDestroySuite struct {
	suite.Suite
}

func (s *SSMParameterTreeResourceDestroySuite) Test_destroy() {
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
		destroyTreeTestCase(providerCtx, loader),
		destroyTreeTolerateNotFoundTestCase(providerCtx, loader),
		destroyTreeErrorTestCase(providerCtx, loader),
		destroyTreeMissingPathTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		ParameterTreeResource,
		&s.Suite,
	)
}

func destroyTreeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithDeleteParameterOutput(&ssm.DeleteParameterOutput{}),
	)

	// The state records "logLevel" in values, "apiToken" in secureValues, and an
	// "orphanKey" only present in the computed parameters metadata; all three are owned
	// and must be deleted, in sorted order.
	stateSpec := treeStateSpecData(
		map[string]string{"logLevel": "info"},
		map[string]string{"apiToken": "super-secret"},
	)
	stateSpec.Fields["parameters"].Fields["orphanKey"] = &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"valueHash": core.MappingNodeFromString(parameterTreeValueHash("orphaned")),
		},
	}

	deleteForKey := func(key string) func(arg any) bool {
		return func(arg any) bool {
			in, ok := arg.(*ssm.DeleteParameterInput)
			return ok && aws.ToString(in.Name) == testTreePath+"/"+key
		}
	}

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ssmservice.Service]{
		Name:             "deletes every owned parameter in sorted order",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: destroyInput(providerCtx, stateSpec),
		DestroyActionsCalled: map[string]any{
			"DeleteParameter": []any{
				deleteForKey("apiToken"),
				deleteForKey("logLevel"),
				deleteForKey("orphanKey"),
			},
		},
		ExpectError: false,
	}
}

func destroyTreeTolerateNotFoundTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithDeleteParameterError(&ssmtypes.ParameterNotFound{}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ssmservice.Service]{
		Name:             "tolerates parameters already deleted out-of-band",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: destroyInput(
			providerCtx,
			treeStateSpecData(map[string]string{"logLevel": "info"}, nil),
		),
		ExpectError: false,
	}
}

func destroyTreeErrorTestCase(
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
		Input: destroyInput(
			providerCtx,
			treeStateSpecData(map[string]string{"logLevel": "info"}, nil),
		),
		ExpectError: true,
	}
}

func destroyTreeMissingPathTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock()

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, ssmservice.Service]{
		Name:             "returns error when path is missing from state",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: destroyInput(providerCtx, &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		}),
		DestroyActionsNotCalled: []string{"DeleteParameter"},
		ExpectError:             true,
	}
}

func TestSSMParameterTreeResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(SSMParameterTreeResourceDestroySuite))
}
