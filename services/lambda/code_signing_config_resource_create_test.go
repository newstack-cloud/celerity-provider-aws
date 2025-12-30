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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type LambdaCodeSigningConfigResourceCreateSuite struct {
	suite.Suite
}

func (s *LambdaCodeSigningConfigResourceCreateSuite) Test_create_lambda_code_signing_config() {
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
		createBasicCodeSigningConfigTestCase(providerCtx, loader),
		createCodeSigningConfigWithDescriptionTestCase(providerCtx, loader),
		createCodeSigningConfigWithPolicyTestCase(providerCtx, loader),
		createCodeSigningConfigWithTagsTestCase(providerCtx, loader),
		createComplexCodeSigningConfigTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	codeSigningConfigResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return CodeSigningConfigResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		codeSigningConfigResourceWrapper,
		&s.Suite,
	)
}

func createBasicCodeSigningConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	cscArn := "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef0"
	cscId := "csc-1234567890abcdef0"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateCodeSigningConfigOutput(&lambda.CreateCodeSigningConfigOutput{
			CodeSigningConfig: &types.CodeSigningConfig{
				CodeSigningConfigArn: aws.String(cscArn),
				CodeSigningConfigId:  aws.String(cscId),
				AllowedPublishers: &types.AllowedPublishers{
					SigningProfileVersionArns: []string{
						"arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12",
					},
				},
			},
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"allowedPublishers": {
				Fields: map[string]*core.MappingNode{
					"signingProfileVersionArns": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create basic code signing config",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/codeSigningConfig",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.allowedPublishers",
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
			"CreateCodeSigningConfig": &lambda.CreateCodeSigningConfigInput{
				AllowedPublishers: &types.AllowedPublishers{
					SigningProfileVersionArns: []string{
						"arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12",
					},
				},
			},
		},
	}
}

func createCodeSigningConfigWithDescriptionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	cscArn := "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef1"
	cscId := "csc-1234567890abcdef1"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateCodeSigningConfigOutput(&lambda.CreateCodeSigningConfigOutput{
			CodeSigningConfig: &types.CodeSigningConfig{
				CodeSigningConfigArn: aws.String(cscArn),
				CodeSigningConfigId:  aws.String(cscId),
				Description:          aws.String("Test code signing configuration"),
				AllowedPublishers: &types.AllowedPublishers{
					SigningProfileVersionArns: []string{
						"arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12",
					},
				},
			},
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"allowedPublishers": {
				Fields: map[string]*core.MappingNode{
					"signingProfileVersionArns": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12"),
						},
					},
				},
			},
			"description": core.MappingNodeFromString("Test code signing configuration"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create code signing config with description",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/codeSigningConfig",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.allowedPublishers",
					},
					{
						FieldPath: "spec.description",
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
			"CreateCodeSigningConfig": &lambda.CreateCodeSigningConfigInput{
				AllowedPublishers: &types.AllowedPublishers{
					SigningProfileVersionArns: []string{
						"arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12",
					},
				},
				Description: aws.String("Test code signing configuration"),
			},
		},
	}
}

func createCodeSigningConfigWithPolicyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	cscArn := "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef2"
	cscId := "csc-1234567890abcdef2"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateCodeSigningConfigOutput(&lambda.CreateCodeSigningConfigOutput{
			CodeSigningConfig: &types.CodeSigningConfig{
				CodeSigningConfigArn: aws.String(cscArn),
				CodeSigningConfigId:  aws.String(cscId),
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
			"allowedPublishers": {
				Fields: map[string]*core.MappingNode{
					"signingProfileVersionArns": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12"),
						},
					},
				},
			},
			"codeSigningPolicies": {
				Fields: map[string]*core.MappingNode{
					"untrustedArtifactOnDeployment": core.MappingNodeFromString("Enforce"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create code signing config with policy",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/codeSigningConfig",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.allowedPublishers",
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
			"CreateCodeSigningConfig": &lambda.CreateCodeSigningConfigInput{
				AllowedPublishers: &types.AllowedPublishers{
					SigningProfileVersionArns: []string{
						"arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12",
					},
				},
				CodeSigningPolicies: &types.CodeSigningPolicies{
					UntrustedArtifactOnDeployment: types.CodeSigningPolicyEnforce,
				},
			},
		},
	}
}

func createCodeSigningConfigWithTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	cscArn := "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef3"
	cscId := "csc-1234567890abcdef3"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateCodeSigningConfigOutput(&lambda.CreateCodeSigningConfigOutput{
			CodeSigningConfig: &types.CodeSigningConfig{
				CodeSigningConfigArn: aws.String(cscArn),
				CodeSigningConfigId:  aws.String(cscId),
				AllowedPublishers: &types.AllowedPublishers{
					SigningProfileVersionArns: []string{
						"arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12",
					},
				},
			},
		}),
		lambdamock.WithTagResourceOutput(&lambda.TagResourceOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"allowedPublishers": {
				Fields: map[string]*core.MappingNode{
					"signingProfileVersionArns": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12"),
						},
					},
				},
			},
			"tags": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("Test"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Team"),
							"value": core.MappingNodeFromString("Backend"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create code signing config with tags",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/codeSigningConfig",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.allowedPublishers",
					},
					{
						FieldPath: "spec.tags",
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
			"CreateCodeSigningConfig": &lambda.CreateCodeSigningConfigInput{
				AllowedPublishers: &types.AllowedPublishers{
					SigningProfileVersionArns: []string{
						"arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12",
					},
				},
			},
			"TagResource": &lambda.TagResourceInput{
				Resource: aws.String(cscArn),
				Tags: map[string]string{
					"Environment": "Test",
					"Team":        "Backend",
				},
			},
		},
	}
}

func createComplexCodeSigningConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	cscArn := "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef4"
	cscId := "csc-1234567890abcdef4"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateCodeSigningConfigOutput(&lambda.CreateCodeSigningConfigOutput{
			CodeSigningConfig: &types.CodeSigningConfig{
				CodeSigningConfigArn: aws.String(cscArn),
				CodeSigningConfigId:  aws.String(cscId),
				Description:          aws.String("Complex production code signing configuration"),
				AllowedPublishers: &types.AllowedPublishers{
					SigningProfileVersionArns: []string{
						"arn:aws:signer:us-west-2:123456789012:/signing-profiles/ProdProfile/abcdef12",
						"arn:aws:signer:us-west-2:123456789012:/signing-profiles/BackupProfile/ghijkl34",
					},
				},
				CodeSigningPolicies: &types.CodeSigningPolicies{
					UntrustedArtifactOnDeployment: types.CodeSigningPolicyEnforce,
				},
			},
		}),
		lambdamock.WithTagResourceOutput(&lambda.TagResourceOutput{}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"allowedPublishers": {
				Fields: map[string]*core.MappingNode{
					"signingProfileVersionArns": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:/signing-profiles/ProdProfile/abcdef12"),
							core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:/signing-profiles/BackupProfile/ghijkl34"),
						},
					},
				},
			},
			"codeSigningPolicies": {
				Fields: map[string]*core.MappingNode{
					"untrustedArtifactOnDeployment": core.MappingNodeFromString("Enforce"),
				},
			},
			"description": core.MappingNodeFromString("Complex production code signing configuration"),
			"tags": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("Production"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Team"),
							"value": core.MappingNodeFromString("Security"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Project"),
							"value": core.MappingNodeFromString("MainApp"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create complex code signing config with all features",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/codeSigningConfig",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.allowedPublishers",
					},
					{
						FieldPath: "spec.codeSigningPolicies",
					},
					{
						FieldPath: "spec.description",
					},
					{
						FieldPath: "spec.tags",
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
			"CreateCodeSigningConfig": &lambda.CreateCodeSigningConfigInput{
				AllowedPublishers: &types.AllowedPublishers{
					SigningProfileVersionArns: []string{
						"arn:aws:signer:us-west-2:123456789012:/signing-profiles/ProdProfile/abcdef12",
						"arn:aws:signer:us-west-2:123456789012:/signing-profiles/BackupProfile/ghijkl34",
					},
				},
				CodeSigningPolicies: &types.CodeSigningPolicies{
					UntrustedArtifactOnDeployment: types.CodeSigningPolicyEnforce,
				},
				Description: aws.String("Complex production code signing configuration"),
			},
			"TagResource": &lambda.TagResourceInput{
				Resource: aws.String(cscArn),
				Tags: map[string]string{
					"Environment": "Production",
					"Team":        "Security",
					"Project":     "MainApp",
				},
			},
		},
	}
}

func TestLambdaCodeSigningConfigResourceCreate(t *testing.T) {
	suite.Run(t, new(LambdaCodeSigningConfigResourceCreateSuite))
}
