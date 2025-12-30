//go:build unit

package lambda

import (
	"errors"
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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type LambdaCodeSigningConfigResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *LambdaCodeSigningConfigResourceGetExternalStateSuite) Test_get_external_state() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			pluginutils.SessionIDKey: core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		createGetExternalStateBasicCodeSigningConfigTestCase(providerCtx, loader),
		createGetExternalStateCodeSigningConfigWithDescriptionTestCase(providerCtx, loader),
		createGetExternalStateCodeSigningConfigWithPolicyTestCase(providerCtx, loader),
		createGetExternalStateCodeSigningConfigWithTagsTestCase(providerCtx, loader),
		createGetExternalStateCodeSigningConfigWithAllowedPublishersTestCase(providerCtx, loader),
		createGetExternalStateCodeSigningConfigErrorTestCase(providerCtx, loader),
	}

	// Create a wrapper function that matches the expected signature
	codeSigningConfigResourceWrapper := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return CodeSigningConfigResource(serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		codeSigningConfigResourceWrapper,
		&s.Suite,
	)
}

func TestLambdaCodeSigningConfigResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(LambdaCodeSigningConfigResourceGetExternalStateSuite))
}

// Test case generator functions below.

func createGetExternalStateBasicCodeSigningConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets basic code signing config state",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetCodeSigningConfigOutput(&lambda.GetCodeSigningConfigOutput{
				CodeSigningConfig: &types.CodeSigningConfig{
					CodeSigningConfigArn: aws.String("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
					CodeSigningConfigId:  aws.String("test-config-id"),
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"codeSigningConfigArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"codeSigningConfigArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
					"codeSigningConfigId":  core.MappingNodeFromString("test-config-id"),
				},
			},
		},
		ExpectError: false,
	}
}

func createGetExternalStateCodeSigningConfigWithDescriptionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets code signing config state with description",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetCodeSigningConfigOutput(&lambda.GetCodeSigningConfigOutput{
				CodeSigningConfig: &types.CodeSigningConfig{
					CodeSigningConfigArn: aws.String("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
					CodeSigningConfigId:  aws.String("test-config-id"),
					Description:          aws.String("Test code signing config"),
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"codeSigningConfigArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"codeSigningConfigArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
					"codeSigningConfigId":  core.MappingNodeFromString("test-config-id"),
					"description":          core.MappingNodeFromString("Test code signing config"),
				},
			},
		},
		ExpectError: false,
	}
}

func createGetExternalStateCodeSigningConfigWithPolicyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets code signing config state with policy",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetCodeSigningConfigOutput(&lambda.GetCodeSigningConfigOutput{
				CodeSigningConfig: &types.CodeSigningConfig{
					CodeSigningConfigArn: aws.String("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
					CodeSigningConfigId:  aws.String("test-config-id"),
					CodeSigningPolicies: &types.CodeSigningPolicies{
						UntrustedArtifactOnDeployment: types.CodeSigningPolicyWarn,
					},
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"codeSigningConfigArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"codeSigningConfigArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
					"codeSigningConfigId":  core.MappingNodeFromString("test-config-id"),
					"codeSigningPolicies": {
						Fields: map[string]*core.MappingNode{
							"untrustedArtifactOnDeployment": core.MappingNodeFromString("Warn"),
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createGetExternalStateCodeSigningConfigWithTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	tags := map[string]string{
		"Environment": "test",
		"Project":     "bluelink",
		"Service":     "lambda",
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets code signing config state with tags",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetCodeSigningConfigOutput(&lambda.GetCodeSigningConfigOutput{
				CodeSigningConfig: &types.CodeSigningConfig{
					CodeSigningConfigArn: aws.String("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
					CodeSigningConfigId:  aws.String("test-config-id"),
				},
			}),
			lambdamock.WithListTagsOutput(&lambda.ListTagsOutput{
				Tags: tags,
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"codeSigningConfigArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"codeSigningConfigArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
					"codeSigningConfigId":  core.MappingNodeFromString("test-config-id"),
					"tags": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"key":   core.MappingNodeFromString("Environment"),
									"value": core.MappingNodeFromString("test"),
								},
							},
							{
								Fields: map[string]*core.MappingNode{
									"key":   core.MappingNodeFromString("Project"),
									"value": core.MappingNodeFromString("bluelink"),
								},
							},
							{
								Fields: map[string]*core.MappingNode{
									"key":   core.MappingNodeFromString("Service"),
									"value": core.MappingNodeFromString("lambda"),
								},
							},
						},
					},
				},
			},
		},
		CheckTags:     true,
		TagsFieldName: "tags",
		ExpectError:   false,
	}
}

func createGetExternalStateCodeSigningConfigWithAllowedPublishersTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets code signing config state with allowed publishers",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetCodeSigningConfigOutput(&lambda.GetCodeSigningConfigOutput{
				CodeSigningConfig: &types.CodeSigningConfig{
					CodeSigningConfigArn: aws.String("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
					CodeSigningConfigId:  aws.String("test-config-id"),
					AllowedPublishers: &types.AllowedPublishers{
						SigningProfileVersionArns: []string{
							"arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12",
							"arn:aws:signer:us-west-2:123456789012:/signing-profiles/BackupProfile/ghijkl34",
						},
					},
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"codeSigningConfigArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"codeSigningConfigArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
					"codeSigningConfigId":  core.MappingNodeFromString("test-config-id"),
					"allowedPublishers": {
						Fields: map[string]*core.MappingNode{
							"signingProfileVersionArns": {
								Items: []*core.MappingNode{
									core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:/signing-profiles/TestProfile/abcdef12"),
									core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:/signing-profiles/BackupProfile/ghijkl34"),
								},
							},
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createGetExternalStateCodeSigningConfigErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "handles get code signing config error",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetCodeSigningConfigError(errors.New("failed to get code signing config")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"codeSigningConfigArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
				},
			},
		},
		ExpectError: true,
	}
}
