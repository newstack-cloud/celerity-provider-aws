//go:build unit

package lambdassm

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FunctionParameterTreeLinkStageChangesSuite struct {
	suite.Suite
}

func functionParameterTreeStageLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service],
) provider.Link {
	build := FunctionParameterTreeLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iammock.CreateIamServiceMock() },
		ec2mock.CreateEc2ServiceMockFactory(),
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service],
	) provider.Link {
		return build(FunctionToParameterLinkDeps(deps))
	}
}

func parameterTreeResourceBChanges(withNewField bool) *provider.Changes {
	changes := &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceName: "config-store",
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"path": core.MappingNodeFromString(fptPath),
						"values": core.MappingNodeFields(
							"logLevel", core.MappingNodeFromString("info"),
						),
					},
				},
			},
		},
	}
	if withNewField {
		changes.NewFields = []provider.FieldChange{
			{
				FieldPath: "spec.path",
				NewValue:  core.MappingNodeFromString(fptPath),
			},
		}
	}
	return changes
}

func (s *FunctionParameterTreeLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		stageParameterTreeDefaultEnvVarNameTestCase(),
		stageParameterTreeDisableEnvVarsTestCase(),
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		functionParameterTreeStageLinkFactory(),
		&s.Suite,
	)
}

// The staged field path must use the same sanitised default name the deploy path writes,
// keeping the SSM_PARAMETER_PATH_ prefix shared with the parameter path link.
func stageParameterTreeDefaultEnvVarNameTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name: "stages env var change using the default name when no annotation is set",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerFunctionWithPathAnnotations(map[string]*core.MappingNode{}),
			},
			ResourceBChanges: parameterTreeResourceBChanges(true),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data:   map[string]*core.MappingNode{},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				NewFields: []*provider.FieldChange{
					{
						FieldPath: "apiFunction.environmentVariables[\"" + fptDefaultEnvVarName + "\"]",
						NewValue:  core.MappingNodeFromString(fptPath),
					},
				},
				FieldChangesKnownOnDeploy: []string{"apiFunctionExecutionRole"},
			},
		},
	}
}

func stageParameterTreeDisableEnvVarsTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name: "removes the env var when env var population is disabled",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerFunctionWithPathAnnotations(map[string]*core.MappingNode{
					"aws.lambda.ssm.config-store.envVarName":      core.MappingNodeFromString("APP_CONFIG_STORE_PATH"),
					"aws.lambda.ssm.config-store.populateEnvVars": core.MappingNodeFromBool(false),
				}),
			},
			ResourceBChanges: parameterTreeResourceBChanges(false),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"apiFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									"APP_CONFIG_STORE_PATH": core.MappingNodeFromString(fptPath),
								},
							},
						},
					},
				},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				RemovedFields: []string{
					"apiFunction.environmentVariables[\"APP_CONFIG_STORE_PATH\"]",
				},
			},
		},
	}
}

func TestFunctionParameterTreeLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(FunctionParameterTreeLinkStageChangesSuite))
}
