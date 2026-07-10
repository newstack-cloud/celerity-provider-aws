//go:build unit

package lambdasecretsmanager

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FunctionSecretLinkStageChangesSuite struct {
	suite.Suite
}

const fsStageEnvVarName = "DB_CREDENTIALS_SECRET"

func functionSecretStageLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
) provider.Link {
	build := FunctionSecretLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iammock.CreateIamServiceMock() },
		ec2mock.CreateEc2ServiceMockFactory(),
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
	) provider.Link {
		return build(FunctionToSecretLinkDeps(deps))
	}
}

func callerFunctionWithEnvVarAnnotation() provider.ResourceInfo {
	return provider.ResourceInfo{
		ResourceName: "apiFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"aws.lambda.secretsmanager.dbCredentials.envVarName": core.MappingNodeFromString(fsStageEnvVarName),
					},
				},
			},
		},
	}
}

func secretResourceBChanges(withNewField bool) *provider.Changes {
	changes := &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceName: "dbCredentials",
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"id": core.MappingNodeFromString(testSecretARN),
					},
				},
			},
		},
	}
	if withNewField {
		changes.NewFields = []provider.FieldChange{
			{
				FieldPath: "spec.id",
				NewValue:  core.MappingNodeFromString(testSecretARN),
			},
		}
	}
	return changes
}

func (s *FunctionSecretLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		stageFunctionSecretEnvVarChangeTestCase(),
		stageFunctionSecretDefaultEnvVarNameTestCase(),
		stageFunctionSecretNoChangesTestCase(),
		stageFunctionSecretDisableEnvVarsTestCase(),
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		functionSecretStageLinkFactory(),
		&s.Suite,
	)
}

func stageFunctionSecretEnvVarChangeTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name: "stages env var change for the linked secret",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerFunctionWithEnvVarAnnotation(),
			},
			ResourceBChanges: secretResourceBChanges(true),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data:   map[string]*core.MappingNode{},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				NewFields: []*provider.FieldChange{
					{
						FieldPath: "apiFunction.environmentVariables[\"" + fsStageEnvVarName + "\"]",
						NewValue:  core.MappingNodeFromString(testSecretARN),
					},
				},
				FieldChangesKnownOnDeploy: []string{"apiFunctionExecutionRole"},
			},
		},
	}
}

// Regression: when no envVarName annotation is set, the staged field path must use the
// same default name the deploy path writes, not an empty name.
func stageFunctionSecretDefaultEnvVarNameTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name: "stages env var change using the default name when no annotation is set",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "apiFunction",
				},
			},
			ResourceBChanges: secretResourceBChanges(true),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data:   map[string]*core.MappingNode{},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				NewFields: []*provider.FieldChange{
					{
						FieldPath: "apiFunction.environmentVariables[\"SECRET_dbCredentials\"]",
						NewValue:  core.MappingNodeFromString(testSecretARN),
					},
				},
				FieldChangesKnownOnDeploy: []string{"apiFunctionExecutionRole"},
			},
		},
	}
}

func stageFunctionSecretNoChangesTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name: "stages no changes when the env var already matches",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerFunctionWithEnvVarAnnotation(),
			},
			ResourceBChanges: secretResourceBChanges(false),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"apiFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									fsStageEnvVarName: core.MappingNodeFromString(testSecretARN),
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
					"apiFunction.environmentVariables[\"" + fsStageEnvVarName + "\"]",
				},
			},
		},
	}
}

func stageFunctionSecretDisableEnvVarsTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	callerInfo := provider.ResourceInfo{
		ResourceName: "apiFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"aws.lambda.secretsmanager.dbCredentials.envVarName":      core.MappingNodeFromString(fsStageEnvVarName),
						"aws.lambda.secretsmanager.dbCredentials.populateEnvVars": core.MappingNodeFromBool(false),
					},
				},
			},
		},
	}

	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name: "removes the env var when env var population is disabled",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerInfo,
			},
			ResourceBChanges: secretResourceBChanges(false),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"apiFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									fsStageEnvVarName: core.MappingNodeFromString(testSecretARN),
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
					"apiFunction.environmentVariables[\"" + fsStageEnvVarName + "\"]",
				},
			},
		},
	}
}

func TestFunctionSecretLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(FunctionSecretLinkStageChangesSuite))
}
