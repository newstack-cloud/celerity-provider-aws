//go:build unit

package lambdakms

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	kmsservice "github.com/newstack-cloud/bluelink-provider-aws/services/kms/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FunctionKeyLinkStageChangesSuite struct {
	suite.Suite
}

const fkStageEnvVarName = "DATA_ENCRYPTION_KEY"

func functionKeyStageLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
) provider.Link {
	build := FunctionKeyLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iammock.CreateIamServiceMock() },
		ec2mock.CreateEc2ServiceMockFactory(),
		func(c *aws.Config, pc provider.Context) kmsservice.Service { return defaultKMSMock() },
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
	) provider.Link {
		return build(FunctionToKeyLinkDeps(deps))
	}
}

func callerFunctionWithEnvVarAnnotation() provider.ResourceInfo {
	return provider.ResourceInfo{
		ResourceName: "encryptFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"aws.lambda.kms.dataKey.envVarName": core.MappingNodeFromString(fkStageEnvVarName),
					},
				},
			},
		},
	}
}

func keyResourceBChanges(withNewField bool) *provider.Changes {
	changes := &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceName: "dataKey",
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn": core.MappingNodeFromString(testKeyARN),
					},
				},
			},
		},
	}
	if withNewField {
		changes.NewFields = []provider.FieldChange{
			{
				FieldPath: "spec.arn",
				NewValue:  core.MappingNodeFromString(testKeyARN),
			},
		}
	}
	return changes
}

func (s *FunctionKeyLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		stageFunctionKeyEnvVarChangeTestCase(),
		stageFunctionKeyNoChangesTestCase(),
		stageFunctionKeyDisableEnvVarsTestCase(),
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		functionKeyStageLinkFactory(),
		&s.Suite,
	)
}

func stageFunctionKeyEnvVarChangeTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name: "stages env var change for the linked key",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerFunctionWithEnvVarAnnotation(),
			},
			ResourceBChanges: keyResourceBChanges(true),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data:   map[string]*core.MappingNode{},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				NewFields: []*provider.FieldChange{
					{
						FieldPath: "encryptFunction.environmentVariables[\"" + fkStageEnvVarName + "\"]",
						NewValue:  core.MappingNodeFromString(testKeyARN),
					},
				},
				FieldChangesKnownOnDeploy: []string{"encryptFunctionExecutionRole"},
			},
		},
	}
}

func stageFunctionKeyNoChangesTestCase() plugintestutils.LinkChangeStagingTestCase[
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
			ResourceBChanges: keyResourceBChanges(false),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"encryptFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									fkStageEnvVarName: core.MappingNodeFromString(testKeyARN),
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
					"encryptFunction.environmentVariables[\"" + fkStageEnvVarName + "\"]",
				},
			},
		},
	}
}

func stageFunctionKeyDisableEnvVarsTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	callerInfo := provider.ResourceInfo{
		ResourceName: "encryptFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"aws.lambda.kms.dataKey.envVarName":      core.MappingNodeFromString(fkStageEnvVarName),
						"aws.lambda.kms.dataKey.populateEnvVars": core.MappingNodeFromBool(false),
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
			ResourceBChanges: keyResourceBChanges(false),
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"encryptFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									fkStageEnvVarName: core.MappingNodeFromString(testKeyARN),
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
					"encryptFunction.environmentVariables[\"" + fkStageEnvVarName + "\"]",
				},
			},
		},
	}
}

// Builds a function resource that manages a KMS grant. Env var
// population is disabled so the staged changes isolate the grant side effect.
func callerFunctionWithGrant(manage bool, accessLevel string) provider.ResourceInfo {
	fields := map[string]*core.MappingNode{
		"aws.lambda.kms.dataKey.populateEnvVars": core.MappingNodeFromBool(false),
		"aws.lambda.kms.dataKey.manageKeyGrant":  core.MappingNodeFromBool(manage),
	}
	if accessLevel != "" {
		fields["aws.lambda.kms.dataKey.accessLevel"] = core.MappingNodeFromString(accessLevel)
	}
	return provider.ResourceInfo{
		ResourceName: "encryptFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{Fields: fields},
			},
		},
	}
}

func expectedGrantNode(operations ...string) *core.MappingNode {
	items := make([]*core.MappingNode, len(operations))
	for i, operation := range operations {
		items[i] = core.MappingNodeFromString(operation)
	}
	return core.MappingNodeFields(
		"name", core.MappingNodeFromString("bluelink-encryptFunction-dataKey"),
		"operations", &core.MappingNode{Items: items},
	)
}

func linkStateWithGrant(grant *core.MappingNode) *state.LinkState {
	return &state.LinkState{
		LinkID: "test-link",
		Data: map[string]*core.MappingNode{
			"keyGrant": grant,
		},
	}
}

func (s *FunctionKeyLinkStageChangesSuite) Test_stage_grant_changes() {
	decryptGrant := expectedGrantNode("Decrypt", "DescribeKey")
	encryptDecryptGrant := expectedGrantNode(
		"Decrypt", "DescribeKey", "Encrypt", "GenerateDataKey", "GenerateDataKeyWithoutPlaintext",
	)

	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		{
			Name: "stages grant creation when managed and absent",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: callerFunctionWithGrant(true, "decrypt")},
				ResourceBChanges: keyResourceBChanges(false),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: "keyGrant", NewValue: decryptGrant},
					},
				},
			},
		},
		{
			Name: "stages no grant change when managed and matching",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: callerFunctionWithGrant(true, "decrypt")},
				ResourceBChanges: keyResourceBChanges(false),
				CurrentLinkState: linkStateWithGrant(decryptGrant),
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					UnchangedFields: []string{"keyGrant"},
				},
			},
		},
		{
			Name: "stages grant update when the access level changes",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: callerFunctionWithGrant(true, "encryptDecrypt")},
				ResourceBChanges: keyResourceBChanges(false),
				CurrentLinkState: linkStateWithGrant(decryptGrant),
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					ModifiedFields: []*provider.FieldChange{
						{FieldPath: "keyGrant", PrevValue: decryptGrant, NewValue: encryptDecryptGrant},
					},
				},
			},
		},
		{
			Name: "stages grant revocation when management is disabled",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: callerFunctionWithGrant(false, "decrypt")},
				ResourceBChanges: keyResourceBChanges(false),
				CurrentLinkState: linkStateWithGrant(decryptGrant),
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					RemovedFields: []string{"keyGrant"},
				},
			},
		},
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		functionKeyStageLinkFactory(),
		&s.Suite,
	)
}

func TestFunctionKeyLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(FunctionKeyLinkStageChangesSuite))
}
