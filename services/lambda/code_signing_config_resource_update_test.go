//go:build unit

package lambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaCodeSigningConfigResourceUpdateSuite struct {
	suite.Suite
}

func (s *LambdaCodeSigningConfigResourceUpdateSuite) Test_update_lambda_code_signing_config() {
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

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		updateCodeSigningConfigTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		CodeSigningConfigResource,
		&s.Suite,
	)
}

func updateCodeSigningConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	cscArn := "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef0"
	cscId := "csc-1234567890abcdef0"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateCodeSigningConfigOutput(&lambda.UpdateCodeSigningConfigOutput{
			CodeSigningConfig: &types.CodeSigningConfig{
				CodeSigningConfigArn: aws.String(cscArn),
				CodeSigningConfigId:  aws.String(cscId),
				Description:          aws.String("Updated description"),
				AllowedPublishers: &types.AllowedPublishers{
					SigningProfileVersionArns: []string{
						"arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12",
					},
				},
				CodeSigningPolicies: &types.CodeSigningPolicies{
					UntrustedArtifactOnDeployment: types.CodeSigningPolicyEnforce,
				},
			},
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"codeSigningConfigArn": core.MappingNodeFromString(cscArn),
			"allowedPublishers": {
				Fields: map[string]*core.MappingNode{
					"signingProfileVersionArns": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12"),
						},
					},
				},
			},
			"description": core.MappingNodeFromString("Updated description"),
			"codeSigningPolicies": {
				Fields: map[string]*core.MappingNode{
					"untrustedArtifactOnDeployment": core.MappingNodeFromString("Enforce"),
				},
			},
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"codeSigningConfigArn": core.MappingNodeFromString(cscArn),
			"allowedPublishers": {
				Fields: map[string]*core.MappingNode{
					"signingProfileVersionArns": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12"),
						},
					},
				},
			},
			"description": core.MappingNodeFromString("Old description"),
			"codeSigningPolicies": {
				Fields: map[string]*core.MappingNode{
					"untrustedArtifactOnDeployment": core.MappingNodeFromString("Warn"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update code signing config",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-csc-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-csc-id",
					ResourceName: "TestCodeSigningConfig",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-csc-id",
						Name:       "TestCodeSigningConfig",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/codeSigningConfig",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.description",
					},
					{
						FieldPath: "spec.codeSigningPolicies",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.codeSigningConfigArn": core.MappingNodeFromString(cscArn),
				"spec.codeSigningConfigId":  core.MappingNodeFromString(cscId),
			},
		},
		SaveActionsCalled: map[string]any{
			"UpdateCodeSigningConfig": &lambda.UpdateCodeSigningConfigInput{
				CodeSigningConfigArn: aws.String(cscArn),
				AllowedPublishers: &types.AllowedPublishers{
					SigningProfileVersionArns: []string{
						"arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12",
					},
				},
				Description: aws.String("Updated description"),
				CodeSigningPolicies: &types.CodeSigningPolicies{
					UntrustedArtifactOnDeployment: types.CodeSigningPolicyEnforce,
				},
			},
		},
	}
}

func TestLambdaCodeSigningConfigResourceUpdate(t *testing.T) {
	suite.Run(t, new(LambdaCodeSigningConfigResourceUpdateSuite))
}
