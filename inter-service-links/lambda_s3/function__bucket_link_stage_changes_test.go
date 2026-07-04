//go:build unit

package lambdas3

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

type FunctionBucketLinkStageChangesSuite struct {
	suite.Suite
}

const fbStageEnvVarName = "UPLOADS_BUCKET_NAME"

func functionBucketStageLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
) provider.Link {
	build := FunctionBucketLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iammock.CreateIamServiceMock() },
		ec2mock.CreateEc2ServiceMockFactory(),
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
	) provider.Link {
		return build(FunctionToBucketLinkDeps(deps))
	}
}

func callerFunctionWithEnvVarAnnotation() provider.ResourceInfo {
	return provider.ResourceInfo{
		ResourceName: "processUploadsFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"aws.lambda.s3.uploadsBucket.envVarName": core.MappingNodeFromString(fbStageEnvVarName),
					},
				},
			},
		},
	}
}

func (s *FunctionBucketLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		stageFunctionBucketEnvVarChangeTestCase(),
		stageFunctionBucketNoChangesTestCase(),
		stageFunctionBucketDisableEnvVarsTestCase(),
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		functionBucketStageLinkFactory(),
		&s.Suite,
	)
}

func stageFunctionBucketEnvVarChangeTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	return plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name: "stages env var change for the linked bucket",
		Input: &provider.LinkStageChangesInput{
			ResourceAChanges: &provider.Changes{
				AppliedResourceInfo: callerFunctionWithEnvVarAnnotation(),
			},
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "uploadsBucket",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"bucketName": core.MappingNodeFromString("my-app-uploads"),
							},
						},
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.bucketName",
						NewValue:  core.MappingNodeFromString("my-app-uploads"),
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
						FieldPath: "processUploadsFunction.environmentVariables[\"" + fbStageEnvVarName + "\"]",
						NewValue:  core.MappingNodeFromString("my-app-uploads"),
					},
				},
				FieldChangesKnownOnDeploy: []string{"processUploadsFunctionExecutionRole"},
			},
		},
	}
}

func stageFunctionBucketNoChangesTestCase() plugintestutils.LinkChangeStagingTestCase[
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
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "uploadsBucket",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"bucketName": core.MappingNodeFromString("my-app-uploads"),
							},
						},
					},
				},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"processUploadsFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									fbStageEnvVarName: core.MappingNodeFromString("my-app-uploads"),
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
					"processUploadsFunction.environmentVariables[\"" + fbStageEnvVarName + "\"]",
				},
			},
		},
	}
}

func stageFunctionBucketDisableEnvVarsTestCase() plugintestutils.LinkChangeStagingTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	callerInfo := provider.ResourceInfo{
		ResourceName: "processUploadsFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"aws.lambda.s3.uploadsBucket.envVarName":      core.MappingNodeFromString(fbStageEnvVarName),
						"aws.lambda.s3.uploadsBucket.populateEnvVars": core.MappingNodeFromBool(false),
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
			ResourceBChanges: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceName: "uploadsBucket",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Spec: &core.MappingNode{
							Fields: map[string]*core.MappingNode{
								"bucketName": core.MappingNodeFromString("my-app-uploads"),
							},
						},
					},
				},
			},
			CurrentLinkState: &state.LinkState{
				LinkID: "test-link",
				Data: map[string]*core.MappingNode{
					"processUploadsFunction": {
						Fields: map[string]*core.MappingNode{
							"environmentVariables": {
								Fields: map[string]*core.MappingNode{
									fbStageEnvVarName: core.MappingNodeFromString("my-app-uploads"),
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
					"processUploadsFunction.environmentVariables[\"" + fbStageEnvVarName + "\"]",
				},
			},
		},
	}
}

func TestFunctionBucketLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(FunctionBucketLinkStageChangesSuite))
}
