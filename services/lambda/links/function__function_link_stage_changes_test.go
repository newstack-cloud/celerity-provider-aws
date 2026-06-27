//go:build unit

package lambdalinks

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FunctionFunctionLinkStageChangesSuite struct {
	suite.Suite
}

const ffEnvVarName = "INVOKE_CALLEE"

func functionFunctionLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, lambdaservice.Service],
) provider.Link {
	build := FunctionFunctionLink(
		iammock.CreateIamServiceMockFactory(),
		ec2mock.CreateEc2ServiceMockFactory(),
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, lambdaservice.Service],
	) provider.Link {
		return build(FunctionToFunctionLinkDeps(deps))
	}
}

func callerResolvedResource(spec *core.MappingNode) *provider.ResolvedResource {
	return &provider.ResolvedResource{
		Spec: spec,
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"aws.lambda.invoke.calleeFunction.envVarName": core.MappingNodeFromString(
						ffEnvVarName,
					),
				},
			},
		},
	}
}

func (s *FunctionFunctionLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		stageFunctionFunctionEnvVarChangeTestCase(),
		stageFunctionFunctionKnownOnDeployTestCase(),
		stageFunctionFunctionNoChangesTestCase(),
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		functionFunctionLinkFactory(),
		&s.Suite,
	)
}

func stageFunctionFunctionEnvVarChangeTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	calleeARN := "arn:aws:lambda:us-west-2:123456789012:function:callee"

	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name: "stages env var change for the linked callee function",
		Input: &provider.LinkStageChangesInput{
			// ResourceAChanges represents the caller function which sources
			// the callee function ARN into its environment variables.
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "callerFunction",
					ResourceWithResolvedSubs: callerResolvedResource(
						&core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"arn": core.MappingNodeFromString(calleeARN),
							},
						},
					),
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.arn",
						NewValue:  core.MappingNodeFromString(calleeARN),
					},
				},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "calleeFunction",
				},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data:   map[string]*core.MappingNode{},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				NewFields: []*provider.FieldChange{
					{
						FieldPath: "callerFunction.environmentVariables[\"" + ffEnvVarName + "\"]",
						NewValue:  core.MappingNodeFromString(calleeARN),
					},
				},
				// The caller function is new (it has NewFields), so the
				// execution-role permission data for this link is only
				// known at deploy time.
				FieldChangesKnownOnDeploy: []string{"callerFunctionExecutionRole"},
			},
		},
	}
}

func stageFunctionFunctionKnownOnDeployTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name: "stages env var change not known until deployment",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "callerFunction",
					ResourceWithResolvedSubs: callerResolvedResource(
						&core.MappingNode{
							Fields: map[string]*core.MappingNode{},
						},
					),
				},
				FieldChangesKnownOnDeploy: []string{
					"spec.arn",
				},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "calleeFunction",
				},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data:   map[string]*core.MappingNode{},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				FieldChangesKnownOnDeploy: []string{
					"callerFunction.environmentVariables[\"" + ffEnvVarName + "\"]",
				},
			},
		},
	}
}

func stageFunctionFunctionNoChangesTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	calleeARN := "arn:aws:lambda:us-west-2:123456789012:function:callee"

	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name: "stages no changes when the env var already matches",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "callerFunction",
					ResourceWithResolvedSubs: callerResolvedResource(
						&core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"arn": core.MappingNodeFromString(calleeARN),
							},
						},
					),
				},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "calleeFunction",
				},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"callerFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									ffEnvVarName: core.MappingNodeFromString(calleeARN),
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
					"callerFunction.environmentVariables[\"" + ffEnvVarName + "\"]",
				},
			},
		},
	}
}

func TestFunctionFunctionLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(FunctionFunctionLinkStageChangesSuite))
}
