//go:build unit

package lambda

import (
	"context"
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

type LambdaCodeSigningConfigDataSourceSuite struct {
	suite.Suite
}

type CodeSigningConfigDataSourceFetchTestCase struct {
	Name                 string
	ServiceFactory       func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service
	ConfigStore          pluginutils.ServiceConfigStore[*aws.Config]
	Input                *provider.DataSourceFetchInput
	ExpectedOutput       *provider.DataSourceFetchOutput
	ExpectError          bool
	ExpectedErrorMessage string
}

func (s *LambdaCodeSigningConfigDataSourceSuite) Test_fetch() {
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			pluginutils.SessionIDKey: core.ScalarFromString("test-session-id"),
		},
	)
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []CodeSigningConfigDataSourceFetchTestCase{
		createBasicCodeSigningConfigFetchTestCase(providerCtx, loader),
		createDataSourceCodeSigningConfigWithDescriptionTestCase(providerCtx, loader),
		createDataSourceCodeSigningConfigWithPolicyTestCase(providerCtx, loader),
		createCodeSigningConfigWithAllConfigsTestCase(providerCtx, loader),
		createCodeSigningConfigFetchErrorTestCase(providerCtx, loader),
		createCodeSigningConfigMissingARNFilterTestCase(providerCtx, loader),
	}

	for _, tc := range testCases {
		s.Run(tc.Name, func() {
			dataSource := CodeSigningConfigDataSource(tc.ServiceFactory, tc.ConfigStore)
			output, err := dataSource.Fetch(context.Background(), tc.Input)

			if tc.ExpectError {
				s.Error(err)
				if tc.ExpectedErrorMessage != "" {
					s.Contains(err.Error(), tc.ExpectedErrorMessage)
				}
			} else {
				s.NoError(err)
				s.Equal(tc.ExpectedOutput, output)
			}
		})
	}
}

func TestLambdaCodeSigningConfigDataSourceSuite(t *testing.T) {
	suite.Run(t, new(LambdaCodeSigningConfigDataSourceSuite))
}

func createBasicCodeSigningConfigFetchTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) CodeSigningConfigDataSourceFetchTestCase {
	return CodeSigningConfigDataSourceFetchTestCase{
		Name: "successfully fetches basic code signing config data",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetCodeSigningConfigOutput(&lambda.GetCodeSigningConfigOutput{
				CodeSigningConfig: &types.CodeSigningConfig{
					CodeSigningConfigArn: aws.String("arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef"),
					CodeSigningConfigId:  aws.String("csc-1234567890abcdef"),
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("arn", "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef"),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"arn":                 core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef"),
				"codeSigningConfigId": core.MappingNodeFromString("csc-1234567890abcdef"),
			},
		},
		ExpectError: false,
	}
}

func createDataSourceCodeSigningConfigWithDescriptionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) CodeSigningConfigDataSourceFetchTestCase {
	return CodeSigningConfigDataSourceFetchTestCase{
		Name: "successfully fetches code signing config with description",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetCodeSigningConfigOutput(&lambda.GetCodeSigningConfigOutput{
				CodeSigningConfig: &types.CodeSigningConfig{
					CodeSigningConfigArn: aws.String("arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef"),
					CodeSigningConfigId:  aws.String("csc-1234567890abcdef"),
					Description:          aws.String("Production code signing configuration"),
				},
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("arn", "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef"),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"arn":                 core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef"),
				"codeSigningConfigId": core.MappingNodeFromString("csc-1234567890abcdef"),
				"description":         core.MappingNodeFromString("Production code signing configuration"),
			},
		},
		ExpectError: false,
	}
}

func createDataSourceCodeSigningConfigWithPolicyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) CodeSigningConfigDataSourceFetchTestCase {
	return CodeSigningConfigDataSourceFetchTestCase{
		Name: "successfully fetches code signing config with policy",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetCodeSigningConfigOutput(&lambda.GetCodeSigningConfigOutput{
				CodeSigningConfig: &types.CodeSigningConfig{
					CodeSigningConfigArn: aws.String("arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef"),
					CodeSigningConfigId:  aws.String("csc-1234567890abcdef"),
					CodeSigningPolicies: &types.CodeSigningPolicies{
						UntrustedArtifactOnDeployment: types.CodeSigningPolicyEnforce,
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
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("arn", "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef"),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"arn":                 core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef"),
				"codeSigningConfigId": core.MappingNodeFromString("csc-1234567890abcdef"),
				"codeSigningPolicies.untrustedArtifactOnDeployment": core.MappingNodeFromString("Enforce"),
			},
		},
		ExpectError: false,
	}
}

func createCodeSigningConfigWithAllConfigsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) CodeSigningConfigDataSourceFetchTestCase {
	return CodeSigningConfigDataSourceFetchTestCase{
		Name: "successfully fetches code signing config with all configurations",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetCodeSigningConfigOutput(&lambda.GetCodeSigningConfigOutput{
				CodeSigningConfig: &types.CodeSigningConfig{
					CodeSigningConfigArn: aws.String("arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef"),
					CodeSigningConfigId:  aws.String("csc-1234567890abcdef"),
					Description:          aws.String("Production code signing configuration"),
					AllowedPublishers: &types.AllowedPublishers{
						SigningProfileVersionArns: []string{
							"arn:aws:signer:us-west-2:123456789012:/signing-profiles/test-profile/1234567890abcdef",
						},
					},
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
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("arn", "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef"),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"arn":                 core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef"),
				"codeSigningConfigId": core.MappingNodeFromString("csc-1234567890abcdef"),
				"description":         core.MappingNodeFromString("Production code signing configuration"),
				"allowedPublishers.signingProfileVersionArns":       {Items: []*core.MappingNode{core.MappingNodeFromString("arn:aws:signer:us-west-2:123456789012:/signing-profiles/test-profile/1234567890abcdef")}},
				"codeSigningPolicies.untrustedArtifactOnDeployment": core.MappingNodeFromString("Warn"),
			},
		},
		ExpectError: false,
	}
}

func createCodeSigningConfigFetchErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) CodeSigningConfigDataSourceFetchTestCase {
	return CodeSigningConfigDataSourceFetchTestCase{
		Name: "handles get code signing config error",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetCodeSigningConfigError(errors.New("ResourceNotFoundException")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: pluginutils.CreateStringEqualsFilter("arn", "arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef"),
			},
		},
		ExpectError:          true,
		ExpectedErrorMessage: "failed to get Lambda code signing config",
	}
}

func createCodeSigningConfigMissingARNFilterTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) CodeSigningConfigDataSourceFetchTestCase {
	return CodeSigningConfigDataSourceFetchTestCase{
		Name:           "handles missing ARN filter",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.DataSourceFetchInput{
			ProviderContext: providerCtx,
			DataSourceWithResolvedSubs: &provider.ResolvedDataSource{
				Filter: &provider.ResolvedDataSourceFilters{
					Filters: []*provider.ResolvedDataSourceFilter{},
				},
			},
		},
		ExpectError:          true,
		ExpectedErrorMessage: "arn filter is required",
	}
}
