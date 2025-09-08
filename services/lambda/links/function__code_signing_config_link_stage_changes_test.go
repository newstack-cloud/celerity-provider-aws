//go:build unit

package lambdalinks

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionCodeSigningConfigLinkStageChangesSuite struct {
	suite.Suite
}

func (s *FunctionCodeSigningConfigLinkStageChangesSuite) Test_stage_changes() {
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

	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		createFunctionCSCLinkChangesTestCase(providerCtx, loader),
		createFunctionCSCLinkChangesKnownOnDeployTestCase(providerCtx, loader),
		createFunctionCSCLinkChangesNoChangesTestCase(providerCtx, loader),
		createFunctionCSCLinkChangesErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		FunctionCodeSigningConfigLink,
		&s.Suite,
	)
}

func createFunctionCSCLinkChangesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	cscARN := "arn:aws:lambda:us-east-1:123456789012:code-signing-config:123456789012"

	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name: "has changes for adding code signing config changes for known value",
		Input: &provider.LinkStageChangesInput{
			// ResourceAChanges represents the changes in
			// the function resource.
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "test-function",
				},
			},
			// ResourceBChanges represents the changes in
			// the code signing config resource from which the ARN
			// is sourced to populate the code signing config ARN
			// in the function resource.
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "test-code-signing-config",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						// The resource spec containing the new code signing config ARN.
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"codeSigningConfigArn": core.MappingNodeFromString(
									cscARN,
								),
							},
						},
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.codeSigningConfigArn",
						NewValue:  core.MappingNodeFromString(cscARN),
					},
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
						FieldPath: "[\"test-function\"].codeSigningConfigArn",
						NewValue:  core.MappingNodeFromString(cscARN),
					},
				},
			},
		},
	}
}

func createFunctionCSCLinkChangesKnownOnDeployTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkChangeStagingTestCase[
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
		Name: "has changes for adding code signing config changes not known until deployment",
		Input: &provider.LinkStageChangesInput{
			// ResourceAChanges represents the changes in
			// the function resource.
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "test-function",
				},
			},
			// ResourceBChanges represents the changes in
			// the code signing config resource from which the ARN
			// is sourced to populate the code signing config ARN
			// in the function resource.
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "test-code-signing-config",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{},
						},
					},
				},
				FieldChangesKnownOnDeploy: []string{
					"spec.codeSigningConfigArn",
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
					"[\"test-function\"].codeSigningConfigArn",
				},
			},
		},
	}
}

func createFunctionCSCLinkChangesNoChangesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	cscARN := "arn:aws:lambda:us-east-1:123456789012:code-signing-config:123456789012"

	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name: "has no changes for adding code signing config changes for known value",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "test-function",
				},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "test-code-signing-config",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"codeSigningConfigArn": core.MappingNodeFromString(
									cscARN,
								),
							},
						},
					},
				},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"test-function": {
						Fields: map[string]*core.MappingNode{
							"codeSigningConfigArn": core.MappingNodeFromString(
								cscARN,
							),
						},
					},
				},
			},
		},
		ExpectedOutput: &provider.LinkStageChangesOutput{
			Changes: &provider.LinkChanges{
				UnchangedFields: []string{
					"[\"test-function\"].codeSigningConfigArn",
				},
			},
		},
	}
}

func createFunctionCSCLinkChangesErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkChangeStagingTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	serviceFactory := lambdamock.CreateLambdaServiceMockFactory()
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:            "handles error when resourceBChanges is missing resolved resource",
		ServiceFactoryA: serviceFactory,
		ConfigStoreA:    configStore,
		ServiceFactoryB: serviceFactory,
		ConfigStoreB:    configStore,
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "test-function",
				},
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "test-code-signing-config",
				},
			},
			CurrentLinkState: &state.LinkState{},
		},
		ExpectError:          true,
		ExpectedErrorMessage: "missing resolved resource",
	}
}

func TestFunctionCodeSigningConfigLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(FunctionCodeSigningConfigLinkStageChangesSuite))
}
