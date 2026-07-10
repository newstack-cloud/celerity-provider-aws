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

type FunctionParameterPathLinkStageChangesSuite struct {
	suite.Suite
}

func functionParameterPathStageLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service],
) provider.Link {
	build := FunctionParameterPathLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iammock.CreateIamServiceMock() },
		ec2mock.CreateEc2ServiceMockFactory(),
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service],
	) provider.Link {
		return build(FunctionToParameterLinkDeps(deps))
	}
}

func callerFunctionWithPathAnnotations(annotations map[string]*core.MappingNode) provider.ResourceInfo {
	return provider.ResourceInfo{
		ResourceName: "apiFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{Fields: annotations},
			},
		},
	}
}

func parameterPathResourceBChanges(withNewField bool) *provider.Changes {
	changes := &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceName: "app-config",
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"path": core.MappingNodeFromString(fppPath),
					},
				},
			},
		},
	}
	if withNewField {
		changes.NewFields = []provider.FieldChange{
			{
				FieldPath: "spec.path",
				NewValue:  core.MappingNodeFromString(fppPath),
			},
		}
	}
	return changes
}

func (s *FunctionParameterPathLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		stageParameterPathEnvVarChangeTestCase(),
		stageParameterPathDefaultEnvVarNameTestCase(),
		stageParameterPathDisableEnvVarsTestCase(),
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		functionParameterPathStageLinkFactory(),
		&s.Suite,
	)
}

func stageParameterPathEnvVarChangeTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name: "stages env var change for the linked parameter path with a custom name",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerFunctionWithPathAnnotations(map[string]*core.MappingNode{
					"aws.lambda.ssm.app-config.envVarName": core.MappingNodeFromString(fppCustomEnvVarName),
				}),
			},
			ResourceBChanges: parameterPathResourceBChanges(true),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data:   map[string]*core.MappingNode{},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				NewFields: []*provider.FieldChange{
					{
						FieldPath: "apiFunction.environmentVariables[\"" + fppCustomEnvVarName + "\"]",
						NewValue:  core.MappingNodeFromString(fppPath),
					},
				},
				FieldChangesKnownOnDeploy: []string{"apiFunctionExecutionRole"},
			},
		},
	}
}

// Regression: when no envVarName annotation is set, the staged field path must use the
// same sanitised default name the deploy path writes.
func stageParameterPathDefaultEnvVarNameTestCase() plugintestutils.LinkChangeStagingTestCase[
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
			ResourceBChanges: parameterPathResourceBChanges(true),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data:   map[string]*core.MappingNode{},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				NewFields: []*provider.FieldChange{
					{
						FieldPath: "apiFunction.environmentVariables[\"" + fppDefaultEnvVarName + "\"]",
						NewValue:  core.MappingNodeFromString(fppPath),
					},
				},
				FieldChangesKnownOnDeploy: []string{"apiFunctionExecutionRole"},
			},
		},
	}
}

func stageParameterPathDisableEnvVarsTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name: "removes the env var when env var population is disabled",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerFunctionWithPathAnnotations(map[string]*core.MappingNode{
					"aws.lambda.ssm.app-config.envVarName":      core.MappingNodeFromString(fppCustomEnvVarName),
					"aws.lambda.ssm.app-config.populateEnvVars": core.MappingNodeFromBool(false),
				}),
			},
			ResourceBChanges: parameterPathResourceBChanges(false),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"apiFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									fppCustomEnvVarName: core.MappingNodeFromString(fppPath),
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
					"apiFunction.environmentVariables[\"" + fppCustomEnvVarName + "\"]",
				},
			},
		},
	}
}

func TestFunctionParameterPathLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(FunctionParameterPathLinkStageChangesSuite))
}
