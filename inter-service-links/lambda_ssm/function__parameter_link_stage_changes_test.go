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

type FunctionParameterLinkStageChangesSuite struct {
	suite.Suite
}

const fpStageEnvVarName = "DB_HOST_PARAM"

func functionParameterStageLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service],
) provider.Link {
	build := FunctionParameterLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iammock.CreateIamServiceMock() },
		ec2mock.CreateEc2ServiceMockFactory(),
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service],
	) provider.Link {
		return build(FunctionToParameterLinkDeps(deps))
	}
}

func callerFunctionWithEnvVarAnnotation() provider.ResourceInfo {
	return provider.ResourceInfo{
		ResourceName: "apiFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"aws.lambda.ssm.dbHost.envVarName": core.MappingNodeFromString(fpStageEnvVarName),
					},
				},
			},
		},
	}
}

func parameterResourceBChanges(withNewField bool) *provider.Changes {
	changes := &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceName: "dbHost",
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"name": core.MappingNodeFromString(testParameterName),
						"arn":  core.MappingNodeFromString(testParameterARN),
					},
				},
			},
		},
	}
	if withNewField {
		changes.NewFields = []provider.FieldChange{
			{
				FieldPath: "spec.name",
				NewValue:  core.MappingNodeFromString(testParameterName),
			},
		}
	}
	return changes
}

func (s *FunctionParameterLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		stageFunctionParameterEnvVarChangeTestCase(),
		stageFunctionParameterNoChangesTestCase(),
		stageFunctionParameterDisableEnvVarsTestCase(),
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		functionParameterStageLinkFactory(),
		&s.Suite,
	)
}

func stageFunctionParameterEnvVarChangeTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name: "stages env var change for the linked parameter",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerFunctionWithEnvVarAnnotation(),
			},
			ResourceBChanges: parameterResourceBChanges(true),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data:   map[string]*core.MappingNode{},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				NewFields: []*provider.FieldChange{
					{
						FieldPath: "apiFunction.environmentVariables[\"" + fpStageEnvVarName + "\"]",
						NewValue:  core.MappingNodeFromString(testParameterName),
					},
				},
				FieldChangesKnownOnDeploy: []string{"apiFunctionExecutionRole"},
			},
		},
	}
}

func stageFunctionParameterNoChangesTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name: "stages no changes when the env var already matches",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerFunctionWithEnvVarAnnotation(),
			},
			ResourceBChanges: parameterResourceBChanges(false),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"apiFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									fpStageEnvVarName: core.MappingNodeFromString(testParameterName),
								},
							},
						},
					},
				},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				UnchangedFields: []string{
					"apiFunction.environmentVariables[\"" + fpStageEnvVarName + "\"]",
				},
			},
		},
	}
}

func stageFunctionParameterDisableEnvVarsTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	callerInfo := provider.ResourceInfo{
		ResourceName: "apiFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"aws.lambda.ssm.dbHost.envVarName":      core.MappingNodeFromString(fpStageEnvVarName),
						"aws.lambda.ssm.dbHost.populateEnvVars": core.MappingNodeFromBool(false),
					},
				},
			},
		},
	}

	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name: "removes the env var when env var population is disabled",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerInfo,
			},
			ResourceBChanges: parameterResourceBChanges(false),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"apiFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									fpStageEnvVarName: core.MappingNodeFromString(testParameterName),
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
					"apiFunction.environmentVariables[\"" + fpStageEnvVarName + "\"]",
				},
			},
		},
	}
}

func TestFunctionParameterLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(FunctionParameterLinkStageChangesSuite))
}
