package lambda

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/stretchr/testify/suite"
	"github.com/two-hundred/celerity-provider-aws/internal/testutils"
	"github.com/two-hundred/celerity-provider-aws/utils"
	"github.com/two-hundred/celerity/libs/blueprint/core"
	"github.com/two-hundred/celerity/libs/blueprint/provider"
	"github.com/two-hundred/celerity/libs/plugin-framework/sdk/pluginutils"
)

type LambdaFunctionResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *LambdaFunctionResourceGetExternalStateSuite) Test_get_external_state() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := testutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			pluginutils.SessionIDKey: core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []getExternalStateTestCase{
		createBasicFunctionStateTestCase(providerCtx, loader),
		createAllOptionalConfigsTestCase(providerCtx, loader),
		createGetFunctionErrorTestCase(providerCtx, loader),
		createGetFunctionCodeSigningErrorTestCase(providerCtx, loader),
		createEphemeralStorageTestCase(providerCtx, loader),
		createImageConfigTestCase(providerCtx, loader),
		createTracingAndRuntimeVersionTestCase(providerCtx, loader),
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			lambdaActions := lambdaFunctionResourceActions{
				lambdaServiceFactory: tc.lambdaServiceFactory,
				awsConfigStore:       tc.awsConfigStore,
			}

			output, err := lambdaActions.GetExternalState(context.Background(), tc.input)

			if tc.expectError {
				s.Error(err)
				return
			}

			s.NoError(err)

			// Special handling for tags: compare as sets (order-independent) for the all-optional-configs test case
			if tc.name == "successfully gets function state with all optional configurations" {
				expectedTags := tc.expectedOutput.ResourceSpecState.Fields["tags"].Items
				actualTags := output.ResourceSpecState.Fields["tags"].Items
				testutils.CompareTags(s.T(), expectedTags, actualTags)

				s.Equal(
					testutils.ShallowCopy(tc.expectedOutput.ResourceSpecState.Fields, "tags"),
					testutils.ShallowCopy(output.ResourceSpecState.Fields, "tags"),
				)
			} else {
				s.Equal(tc.expectedOutput, output)
			}
		})
	}
}

func TestLambdaFunctionResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionResourceGetExternalStateSuite))
}

// Test case generator functions below.

func createBasicFunctionStateTestCase(providerCtx provider.Context, loader *testutils.MockAWSConfigLoader) getExternalStateTestCase {
	return getExternalStateTestCase{
		name: "successfully gets basic function state",
		lambdaServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionOutput(createBaseTestFunctionConfig(
				"test-function",
				types.RuntimeNodejs18x,
				"index.handler",
				"arn:aws:iam::123456789012:role/test-role",
			)),
			WithGetFunctionCodeSigningOutput(&lambda.GetFunctionCodeSigningConfigOutput{}),
			WithGetFunctionRecursionOutput(&lambda.GetFunctionRecursionConfigOutput{}),
			WithGetFunctionConcurrencyOutput(&lambda.GetFunctionConcurrencyOutput{}),
		),
		awsConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		expectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"architecture": core.MappingNodeFromString("x86_64"),
					"functionName": core.MappingNodeFromString("test-function"),
					"runtime":      core.MappingNodeFromString("nodejs18.x"),
					"handler":      core.MappingNodeFromString("index.handler"),
					"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				},
			},
		},
		expectError: false,
	}
}

func createAllOptionalConfigsTestCase(providerCtx provider.Context, loader *testutils.MockAWSConfigLoader) getExternalStateTestCase {
	tags := map[string]string{
		"Environment": "test",
		"Project":     "celerity",
		"Service":     "lambda",
	}

	// Create sorted tag items for expected output
	tagItems := []*core.MappingNode{
		{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString("Environment"),
				"value": core.MappingNodeFromString("test"),
			},
		},
		{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString("Project"),
				"value": core.MappingNodeFromString("celerity"),
			},
		},
		{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString("Service"),
				"value": core.MappingNodeFromString("lambda"),
			},
		},
	}

	expectedOutput := &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"arn":          core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				"architecture": core.MappingNodeFromString("x86_64"),
				"functionName": core.MappingNodeFromString("test-function"),
				"runtime":      core.MappingNodeFromString("nodejs18.x"),
				"handler":      core.MappingNodeFromString("index.handler"),
				"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
				"description":  core.MappingNodeFromString("Test function"),
				"memorySize":   core.MappingNodeFromInt(256),
				"timeout":      core.MappingNodeFromInt(30),
				"environment": {
					Fields: map[string]*core.MappingNode{
						"TEST_VAR": core.MappingNodeFromString("test-value"),
					},
				},
				"deadLetterConfig": {
					Fields: map[string]*core.MappingNode{
						"targetArn": core.MappingNodeFromString("arn:aws:sqs:us-east-1:123456789012:test-queue"),
					},
				},
				"vpcConfig": {
					Fields: map[string]*core.MappingNode{
						"securityGroupIds": {
							Items: []*core.MappingNode{
								core.MappingNodeFromString("sg-12345678"),
							},
						},
						"subnetIds": {
							Items: []*core.MappingNode{
								core.MappingNodeFromString("subnet-12345678"),
							},
						},
					},
				},
				"fileSystemConfig": {
					Fields: map[string]*core.MappingNode{
						"arn":            core.MappingNodeFromString("arn:aws:elasticfilesystem:us-east-1:123456789012:access-point/fsap-1234567890abcdef0"),
						"localMountPath": core.MappingNodeFromString("/mnt/efs"),
					},
				},
				"loggingConfig": {
					Fields: map[string]*core.MappingNode{
						"applicationLogLevel": core.MappingNodeFromString("DEBUG"),
						"logFormat":           core.MappingNodeFromString("JSON"),
						"logGroup":            core.MappingNodeFromString("/aws/lambda/test-function"),
						"systemLogLevel":      core.MappingNodeFromString("INFO"),
					},
				},
				"snapStart": {
					Fields: map[string]*core.MappingNode{
						"applyOn": core.MappingNodeFromString("PublishedVersions"),
					},
				},
				"layers": {
					Items: []*core.MappingNode{
						core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:layer:test-layer-1:1"),
						core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:layer:test-layer-2:2"),
					},
				},
				"tags": {
					Items: tagItems,
				},
				"code": {
					Fields: map[string]*core.MappingNode{
						"s3Bucket": core.MappingNodeFromString("test-bucket"),
						"s3Key":    core.MappingNodeFromString("test-key"),
					},
				},
				"snapStartResponseApplyOn":            core.MappingNodeFromString("PublishedVersions"),
				"snapStartResponseOptimizationStatus": core.MappingNodeFromString("On"),
			},
		},
	}

	return getExternalStateTestCase{
		name: "successfully gets function state with all optional configurations",
		lambdaServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionOutput(&lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					Runtime:      types.RuntimeNodejs18x,
					Handler:      aws.String("index.handler"),
					Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
					Architectures: []types.Architecture{
						types.ArchitectureX8664,
					},
					Description: aws.String("Test function"),
					MemorySize:  aws.Int32(256),
					Timeout:     aws.Int32(30),
					Environment: &types.EnvironmentResponse{
						Variables: map[string]string{
							"TEST_VAR": "test-value",
						},
					},
					DeadLetterConfig: &types.DeadLetterConfig{
						TargetArn: aws.String("arn:aws:sqs:us-east-1:123456789012:test-queue"),
					},
					VpcConfig: &types.VpcConfigResponse{
						SecurityGroupIds: []string{"sg-12345678"},
						SubnetIds:        []string{"subnet-12345678"},
					},
					FileSystemConfigs: []types.FileSystemConfig{
						{
							Arn:            aws.String("arn:aws:elasticfilesystem:us-east-1:123456789012:access-point/fsap-1234567890abcdef0"),
							LocalMountPath: aws.String("/mnt/efs"),
						},
					},
					LoggingConfig: &types.LoggingConfig{
						ApplicationLogLevel: types.ApplicationLogLevelDebug,
						LogFormat:           types.LogFormatJson,
						LogGroup:            aws.String("/aws/lambda/test-function"),
						SystemLogLevel:      types.SystemLogLevelInfo,
					},
					SnapStart: &types.SnapStartResponse{
						ApplyOn:            types.SnapStartApplyOnPublishedVersions,
						OptimizationStatus: types.SnapStartOptimizationStatusOn,
					},
					Layers: []types.Layer{
						{
							Arn:      aws.String("arn:aws:lambda:us-east-1:123456789012:layer:test-layer-1:1"),
							CodeSize: 1024,
						},
						{
							Arn:      aws.String("arn:aws:lambda:us-east-1:123456789012:layer:test-layer-2:2"),
							CodeSize: 2048,
						},
					},
				},
				Code: &types.FunctionCodeLocation{
					Location: aws.String("https://test-bucket.s3.amazonaws.com/test-key"),
				},
				Tags: tags,
			}),
			WithGetFunctionCodeSigningOutput(&lambda.GetFunctionCodeSigningConfigOutput{}),
			WithGetFunctionRecursionOutput(&lambda.GetFunctionRecursionConfigOutput{}),
			WithGetFunctionConcurrencyOutput(&lambda.GetFunctionConcurrencyOutput{}),
		),
		awsConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		expectedOutput: expectedOutput,
		expectError:    false,
	}
}

func createGetFunctionErrorTestCase(providerCtx provider.Context, loader *testutils.MockAWSConfigLoader) getExternalStateTestCase {
	return getExternalStateTestCase{
		name: "handles get function error",
		lambdaServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionError(errors.New("failed to get function")),
		),
		awsConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				},
			},
		},
		expectError: true,
	}
}

func createGetFunctionCodeSigningErrorTestCase(providerCtx provider.Context, loader *testutils.MockAWSConfigLoader) getExternalStateTestCase {
	return getExternalStateTestCase{
		name: "handles get function code signing config error",
		lambdaServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionOutput(createBaseTestFunctionConfig(
				"test-function",
				types.RuntimeNodejs18x,
				"index.handler",
				"arn:aws:iam::123456789012:role/test-role",
			)),
			WithGetFunctionCodeSigningError(errors.New("failed to get code signing config")),
		),
		awsConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				},
			},
		},
		expectError: true,
	}
}

func createEphemeralStorageTestCase(providerCtx provider.Context, loader *testutils.MockAWSConfigLoader) getExternalStateTestCase {
	return getExternalStateTestCase{
		name: "successfully gets function state with ephemeral storage",
		lambdaServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionOutput(&lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					Runtime:      types.RuntimeNodejs18x,
					Handler:      aws.String("index.handler"),
					Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
					Architectures: []types.Architecture{
						types.ArchitectureX8664,
					},
					EphemeralStorage: &types.EphemeralStorage{
						Size: aws.Int32(1024), // 1024 MB = 1 GB
					},
				},
				Code: &types.FunctionCodeLocation{
					Location: aws.String("https://test-bucket.s3.amazonaws.com/test-key"),
				},
			}),
			WithGetFunctionCodeSigningOutput(&lambda.GetFunctionCodeSigningConfigOutput{}),
			WithGetFunctionRecursionOutput(&lambda.GetFunctionRecursionConfigOutput{}),
			WithGetFunctionConcurrencyOutput(&lambda.GetFunctionConcurrencyOutput{}),
		),
		awsConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		expectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":          core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"architecture": core.MappingNodeFromString("x86_64"),
					"functionName": core.MappingNodeFromString("test-function"),
					"runtime":      core.MappingNodeFromString("nodejs18.x"),
					"handler":      core.MappingNodeFromString("index.handler"),
					"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
					"ephemeralStorage": {
						Fields: map[string]*core.MappingNode{
							"size": core.MappingNodeFromInt(1024),
						},
					},
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		expectError: false,
	}
}

func createImageConfigTestCase(providerCtx provider.Context, loader *testutils.MockAWSConfigLoader) getExternalStateTestCase {
	return getExternalStateTestCase{
		name: "successfully gets function state with image configuration",
		lambdaServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionOutput(&lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					Runtime:      types.RuntimeNodejs18x,
					Handler:      aws.String("index.handler"),
					Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
					Architectures: []types.Architecture{
						types.ArchitectureX8664,
					},
					ImageConfigResponse: &types.ImageConfigResponse{
						ImageConfig: &types.ImageConfig{
							Command: []string{
								"app.lambda_handler",
								"--config",
								"config.json",
							},
							EntryPoint: []string{
								"/var/runtime/bootstrap",
							},
							WorkingDirectory: aws.String("/var/task"),
						},
					},
				},
				Code: &types.FunctionCodeLocation{
					Location: aws.String("https://test-bucket.s3.amazonaws.com/test-key"),
				},
			}),
			WithGetFunctionCodeSigningOutput(&lambda.GetFunctionCodeSigningConfigOutput{}),
			WithGetFunctionRecursionOutput(&lambda.GetFunctionRecursionConfigOutput{}),
			WithGetFunctionConcurrencyOutput(&lambda.GetFunctionConcurrencyOutput{}),
		),
		awsConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		expectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":          core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"architecture": core.MappingNodeFromString("x86_64"),
					"functionName": core.MappingNodeFromString("test-function"),
					"runtime":      core.MappingNodeFromString("nodejs18.x"),
					"handler":      core.MappingNodeFromString("index.handler"),
					"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
					"imageConfig": {
						Fields: map[string]*core.MappingNode{
							"command": {
								Items: []*core.MappingNode{
									core.MappingNodeFromString("app.lambda_handler"),
									core.MappingNodeFromString("--config"),
									core.MappingNodeFromString("config.json"),
								},
							},
							"entryPoint": {
								Items: []*core.MappingNode{
									core.MappingNodeFromString("/var/runtime/bootstrap"),
								},
							},
							"workingDirectory": core.MappingNodeFromString("/var/task"),
						},
					},
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		expectError: false,
	}
}

func createTracingAndRuntimeVersionTestCase(providerCtx provider.Context, loader *testutils.MockAWSConfigLoader) getExternalStateTestCase {
	return getExternalStateTestCase{
		name: "successfully gets function state with tracing and runtime version config",
		lambdaServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionOutput(&lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					Runtime:      types.RuntimeNodejs18x,
					Handler:      aws.String("index.handler"),
					Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
					Architectures: []types.Architecture{
						types.ArchitectureX8664,
					},
					TracingConfig: &types.TracingConfigResponse{
						Mode: types.TracingModeActive,
					},
					RuntimeVersionConfig: &types.RuntimeVersionConfig{
						RuntimeVersionArn: aws.String("arn:aws:lambda:us-east-1::runtime-version/test"),
					},
				},
				Code: &types.FunctionCodeLocation{
					Location: aws.String("https://test-bucket.s3.amazonaws.com/test-key"),
				},
			}),
			WithGetFunctionCodeSigningOutput(&lambda.GetFunctionCodeSigningConfigOutput{}),
			WithGetFunctionRecursionOutput(&lambda.GetFunctionRecursionConfigOutput{}),
			WithGetFunctionConcurrencyOutput(&lambda.GetFunctionConcurrencyOutput{}),
		),
		awsConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		expectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":          core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"architecture": core.MappingNodeFromString("x86_64"),
					"functionName": core.MappingNodeFromString("test-function"),
					"runtime":      core.MappingNodeFromString("nodejs18.x"),
					"handler":      core.MappingNodeFromString("index.handler"),
					"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
					"tracingConfig": {
						Fields: map[string]*core.MappingNode{
							"mode": core.MappingNodeFromString("Active"),
						},
					},
					"runtimeManagementConfig": {
						Fields: map[string]*core.MappingNode{
							"runtimeVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-east-1::runtime-version/test"),
						},
					},
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		expectError: false,
	}
}
