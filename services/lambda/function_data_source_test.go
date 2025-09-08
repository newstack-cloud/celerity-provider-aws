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

type LambdaFunctionDataSourceSuite struct {
	suite.Suite
}

// Custom test case structure for data source tests.
type DataSourceFetchTestCase struct {
	Name                 string
	ServiceFactory       func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service
	ConfigStore          pluginutils.ServiceConfigStore[*aws.Config]
	Input                *provider.DataSourceFetchInput
	ExpectedOutput       *provider.DataSourceFetchOutput
	ExpectError          bool
	ExpectedErrorMessage string
}

func (s *LambdaFunctionDataSourceSuite) Test_fetch() {
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

	testCases := []DataSourceFetchTestCase{
		createBasicFunctionFetchTestCase(providerCtx, loader),
		createFunctionWithAllOptionalConfigsTestCase(providerCtx, loader),
		createFunctionWithCodeSigningConfigTestCase(providerCtx, loader),
		createFunctionWithConcurrencyConfigTestCase(providerCtx, loader),
		createFunctionWithQualifierTestCase(providerCtx, loader),
		createFunctionWithRegionFilterTestCase(providerCtx, loader),
		createFunctionFetchErrorTestCase(providerCtx, loader),
		createFunctionCodeSigningConfigErrorTestCase(providerCtx, loader),
		createFunctionConcurrencyConfigErrorTestCase(providerCtx, loader),
		createFunctionMissingNameOrARNTestCase(providerCtx, loader),
	}

	for _, tc := range testCases {
		s.Run(tc.Name, func() {
			// Create the data source
			dataSource := FunctionDataSource(tc.ServiceFactory, tc.ConfigStore)

			// Execute the fetch
			output, err := dataSource.Fetch(context.Background(), tc.Input)

			// Assert results
			if tc.ExpectError {
				s.Error(err)
				if tc.ExpectedErrorMessage != "" {
					s.ErrorContains(err, tc.ExpectedErrorMessage)
				}
				s.Nil(output)
			} else {
				s.NoError(err)
				s.NotNil(output)
				s.Equal(tc.ExpectedOutput.Data, output.Data)
			}
		})
	}
}

func TestLambdaFunctionDataSourceSuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionDataSourceSuite))
}

// Test case generator functions below.

func createBasicFunctionFetchTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "successfully fetches basic function data",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionOutput(createBaseTestFunctionConfig(
				"test-function",
				types.RuntimeNodejs18x,
				"index.handler",
				"arn:aws:iam::123456789012:role/test-role",
			)),
			lambdamock.WithGetFunctionCodeSigningError(errors.New("ResourceNotFoundException")),
			lambdamock.WithGetFunctionConcurrencyError(errors.New("ResourceNotFoundException")),
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
				Filter: pluginutils.CreateStringEqualsFilter("name", "test-function"),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"architecture":   core.MappingNodeFromString("x86_64"),
				"arn":            core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				"codeSHA256":     core.MappingNodeFromString(""),
				"name":           core.MappingNodeFromString("test-function"),
				"qualifiedArn":   core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function:"),
				"sourceCodeSize": core.MappingNodeFromInt(0),
				"version":        core.MappingNodeFromString(""),
				"handler":        core.MappingNodeFromString("index.handler"),
				"role":           core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
				"runtime":        core.MappingNodeFromString("nodejs18.x"),
			},
		},
		ExpectError: false,
	}
}

func createFunctionWithAllOptionalConfigsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "successfully fetches function with all optional configurations",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionOutput(createComplexTestFunctionConfig()),
			lambdamock.WithGetFunctionCodeSigningOutput(&lambda.GetFunctionCodeSigningConfigOutput{
				CodeSigningConfigArn: aws.String("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
			}),
			lambdamock.WithGetFunctionConcurrencyOutput(&lambda.GetFunctionConcurrencyOutput{
				ReservedConcurrentExecutions: aws.Int32(10),
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
				Filter: pluginutils.CreateStringEqualsFilter("name", "test-function"),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"architecture":                    core.MappingNodeFromString("x86_64"),
				"arn":                             core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				"codeSHA256":                      core.MappingNodeFromString("test-sha256"),
				"name":                            core.MappingNodeFromString("test-function"),
				"qualifiedArn":                    core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function:$LATEST"),
				"sourceCodeSize":                  core.MappingNodeFromInt(1024),
				"version":                         core.MappingNodeFromString("$LATEST"),
				"handler":                         core.MappingNodeFromString("index.handler"),
				"role":                            core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
				"runtime":                         core.MappingNodeFromString("nodejs18.x"),
				"codeSigningConfigArn":            core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
				"reservedConcurrentExecutions":    core.MappingNodeFromInt(10),
				"memorySize":                      core.MappingNodeFromInt(256),
				"timeout":                         core.MappingNodeFromInt(30),
				"kmsKeyArn":                       core.MappingNodeFromString("arn:aws:kms:us-west-2:123456789012:key/test-key"),
				"signingJobArn":                   core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-job"),
				"deadLetterConfig.targetArn":      core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:test-queue"),
				"environment.variables":           core.MappingNodeFromString(`{"TEST_VAR":"test-value"}`),
				"ephemeralStorage.size":           core.MappingNodeFromInt(512),
				"fileSystemConfig.arn":            core.MappingNodeFromString("arn:aws:elasticfilesystem:us-west-2:123456789012:access-point/fsap-1234567890abcdef0"),
				"fileSystemConfig.localMountPath": core.MappingNodeFromString("/mnt/efs"),
				"imageUri":                        core.MappingNodeFromString("123456789012.dkr.ecr.us-west-2.amazonaws.com/test-image:latest"),
				"layers": {
					Items: []*core.MappingNode{
						core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"),
					},
				},
				"loggingConfig.applicationLogLevel": core.MappingNodeFromString("DEBUG"),
				"loggingConfig.logFormat":           core.MappingNodeFromString("JSON"),
				"loggingConfig.logGroup":            core.MappingNodeFromString("/aws/lambda/test-function"),
				"loggingConfig.systemLogLevel":      core.MappingNodeFromString("INFO"),
				"tracingConfig.mode":                core.MappingNodeFromString("Active"),
				"vpcConfig.ipv6AllowedForDualStack": core.MappingNodeFromBool(true),
				"vpcConfig.securityGroupIds": {
					Items: []*core.MappingNode{
						core.MappingNodeFromString("sg-12345678"),
					},
				},
				"vpcConfig.subnetIds": {
					Items: []*core.MappingNode{
						core.MappingNodeFromString("subnet-12345678"),
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createFunctionWithCodeSigningConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "successfully fetches function with code signing config",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionOutput(createBaseTestFunctionConfig(
				"test-function",
				types.RuntimeNodejs18x,
				"index.handler",
				"arn:aws:iam::123456789012:role/test-role",
			)),
			lambdamock.WithGetFunctionCodeSigningOutput(&lambda.GetFunctionCodeSigningConfigOutput{
				CodeSigningConfigArn: aws.String("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
			}),
			lambdamock.WithGetFunctionConcurrencyError(errors.New("ResourceNotFoundException")),
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
				Filter: pluginutils.CreateStringEqualsFilter("name", "test-function"),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"architecture":         core.MappingNodeFromString("x86_64"),
				"arn":                  core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				"codeSHA256":           core.MappingNodeFromString(""),
				"name":                 core.MappingNodeFromString("test-function"),
				"qualifiedArn":         core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function:"),
				"sourceCodeSize":       core.MappingNodeFromInt(0),
				"version":              core.MappingNodeFromString(""),
				"handler":              core.MappingNodeFromString("index.handler"),
				"role":                 core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
				"runtime":              core.MappingNodeFromString("nodejs18.x"),
				"codeSigningConfigArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-config"),
			},
		},
		ExpectError: false,
	}
}

func createFunctionWithConcurrencyConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "successfully fetches function with concurrency config",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionOutput(createBaseTestFunctionConfig(
				"test-function",
				types.RuntimeNodejs18x,
				"index.handler",
				"arn:aws:iam::123456789012:role/test-role",
			)),
			lambdamock.WithGetFunctionCodeSigningError(errors.New("ResourceNotFoundException")),
			lambdamock.WithGetFunctionConcurrencyOutput(&lambda.GetFunctionConcurrencyOutput{
				ReservedConcurrentExecutions: aws.Int32(5),
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
				Filter: pluginutils.CreateStringEqualsFilter("name", "test-function"),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"architecture":                 core.MappingNodeFromString("x86_64"),
				"arn":                          core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				"codeSHA256":                   core.MappingNodeFromString(""),
				"name":                         core.MappingNodeFromString("test-function"),
				"qualifiedArn":                 core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function:"),
				"sourceCodeSize":               core.MappingNodeFromInt(0),
				"version":                      core.MappingNodeFromString(""),
				"handler":                      core.MappingNodeFromString("index.handler"),
				"role":                         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
				"runtime":                      core.MappingNodeFromString("nodejs18.x"),
				"reservedConcurrentExecutions": core.MappingNodeFromInt(5),
			},
		},
		ExpectError: false,
	}
}

func createFunctionWithQualifierTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "successfully fetches function with qualifier",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionOutput(createBaseTestFunctionConfig(
				"test-function",
				types.RuntimeNodejs18x,
				"index.handler",
				"arn:aws:iam::123456789012:role/test-role",
			)),
			lambdamock.WithGetFunctionCodeSigningError(errors.New("ResourceNotFoundException")),
			lambdamock.WithGetFunctionConcurrencyError(errors.New("ResourceNotFoundException")),
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
				Filter: pluginutils.CreateStringEqualsFilter("name", "test-function"),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"architecture":   core.MappingNodeFromString("x86_64"),
				"arn":            core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				"codeSHA256":     core.MappingNodeFromString(""),
				"name":           core.MappingNodeFromString("test-function"),
				"qualifiedArn":   core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function:"),
				"sourceCodeSize": core.MappingNodeFromInt(0),
				"version":        core.MappingNodeFromString(""),
				"handler":        core.MappingNodeFromString("index.handler"),
				"role":           core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
				"runtime":        core.MappingNodeFromString("nodejs18.x"),
			},
		},
		ExpectError: false,
	}
}

func createFunctionWithRegionFilterTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "successfully fetches function with region filter",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionOutput(createBaseTestFunctionConfig(
				"test-function",
				types.RuntimeNodejs18x,
				"index.handler",
				"arn:aws:iam::123456789012:role/test-role",
			)),
			lambdamock.WithGetFunctionCodeSigningError(errors.New("ResourceNotFoundException")),
			lambdamock.WithGetFunctionConcurrencyError(errors.New("ResourceNotFoundException")),
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
				Filter: pluginutils.CreateStringEqualsFilter("name", "test-function"),
			},
		},
		ExpectedOutput: &provider.DataSourceFetchOutput{
			Data: map[string]*core.MappingNode{
				"architecture":   core.MappingNodeFromString("x86_64"),
				"arn":            core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				"codeSHA256":     core.MappingNodeFromString(""),
				"name":           core.MappingNodeFromString("test-function"),
				"qualifiedArn":   core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function:"),
				"sourceCodeSize": core.MappingNodeFromInt(0),
				"version":        core.MappingNodeFromString(""),
				"handler":        core.MappingNodeFromString("index.handler"),
				"role":           core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
				"runtime":        core.MappingNodeFromString("nodejs18.x"),
			},
		},
		ExpectError: false,
	}
}

func createFunctionFetchErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "handles get function error",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionError(errors.New("Function not found")),
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
				Filter: pluginutils.CreateStringEqualsFilter("name", "non-existent-function"),
			},
		},
		ExpectedOutput:       nil,
		ExpectError:          true,
		ExpectedErrorMessage: "Function not found",
	}
}

func createFunctionCodeSigningConfigErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "handles code signing config error",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionOutput(createBaseTestFunctionConfig(
				"test-function",
				types.RuntimeNodejs18x,
				"index.handler",
				"arn:aws:iam::123456789012:role/test-role",
			)),
			lambdamock.WithGetFunctionCodeSigningError(errors.New("Access denied")),
			lambdamock.WithGetFunctionConcurrencyError(errors.New("ResourceNotFoundException")),
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
				Filter: pluginutils.CreateStringEqualsFilter("name", "test-function"),
			},
		},
		ExpectedOutput:       nil,
		ExpectError:          true,
		ExpectedErrorMessage: "Access denied",
	}
}

func createFunctionConcurrencyConfigErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name: "handles concurrency config error",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionOutput(createBaseTestFunctionConfig(
				"test-function",
				types.RuntimeNodejs18x,
				"index.handler",
				"arn:aws:iam::123456789012:role/test-role",
			)),
			lambdamock.WithGetFunctionCodeSigningError(errors.New("ResourceNotFoundException")),
			lambdamock.WithGetFunctionConcurrencyError(errors.New("Access denied")),
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
				Filter: pluginutils.CreateStringEqualsFilter("name", "test-function"),
			},
		},
		ExpectedOutput:       nil,
		ExpectError:          true,
		ExpectedErrorMessage: "Access denied",
	}
}

func createFunctionMissingNameOrARNTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) DataSourceFetchTestCase {
	return DataSourceFetchTestCase{
		Name:           "handles missing function name or ARN",
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
		ExpectedOutput:       nil,
		ExpectError:          true,
		ExpectedErrorMessage: "function name or ARN is required for the lambda function data source",
	}
}

// Helper function to create a complex function configuration for testing.
func createComplexTestFunctionConfig() *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &types.FunctionConfiguration{
			FunctionName: aws.String("test-function"),
			FunctionArn:  aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			Runtime:      types.RuntimeNodejs18x,
			Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
			Handler:      aws.String("index.handler"),
			CodeSize:     1024,
			CodeSha256:   aws.String("test-sha256"),
			Version:      aws.String("$LATEST"),
			Architectures: []types.Architecture{
				types.ArchitectureX8664,
			},
			MemorySize:    aws.Int32(256),
			Timeout:       aws.Int32(30),
			KMSKeyArn:     aws.String("arn:aws:kms:us-west-2:123456789012:key/test-key"),
			SigningJobArn: aws.String("arn:aws:lambda:us-west-2:123456789012:code-signing-config:test-job"),
			DeadLetterConfig: &types.DeadLetterConfig{
				TargetArn: aws.String("arn:aws:sqs:us-west-2:123456789012:test-queue"),
			},
			Environment: &types.EnvironmentResponse{
				Variables: map[string]string{
					"TEST_VAR": "test-value",
				},
			},
			EphemeralStorage: &types.EphemeralStorage{
				Size: aws.Int32(512),
			},
			FileSystemConfigs: []types.FileSystemConfig{
				{
					Arn:            aws.String("arn:aws:elasticfilesystem:us-west-2:123456789012:access-point/fsap-1234567890abcdef0"),
					LocalMountPath: aws.String("/mnt/efs"),
				},
			},
			Layers: []types.Layer{
				{
					Arn: aws.String("arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"),
				},
			},
			LoggingConfig: &types.LoggingConfig{
				ApplicationLogLevel: types.ApplicationLogLevelDebug,
				LogFormat:           types.LogFormatJson,
				LogGroup:            aws.String("/aws/lambda/test-function"),
				SystemLogLevel:      types.SystemLogLevelInfo,
			},
			TracingConfig: &types.TracingConfigResponse{
				Mode: types.TracingModeActive,
			},
			VpcConfig: &types.VpcConfigResponse{
				SecurityGroupIds:        []string{"sg-12345678"},
				SubnetIds:               []string{"subnet-12345678"},
				Ipv6AllowedForDualStack: aws.Bool(true),
			},
		},
		Code: &types.FunctionCodeLocation{
			ImageUri: aws.String("123456789012.dkr.ecr.us-west-2.amazonaws.com/test-image:latest"),
		},
	}
}
