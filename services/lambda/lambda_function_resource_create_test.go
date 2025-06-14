package lambda

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/celerity-provider-aws/internal/testutils"
	"github.com/newstack-cloud/celerity-provider-aws/utils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/blueprint/schema"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaFunctionResourceCreateSuite struct {
	suite.Suite
}

func (s *LambdaFunctionResourceCreateSuite) Test_create_lambda_function() {
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

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, Service]{
		createBasicFunctionCreateTestCase(providerCtx, loader),
		createFunctionWithTagsTestCase(providerCtx, loader),
		createFunctionWithSnapStartTestCase(providerCtx, loader),
		createFunctionFailureTestCase(providerCtx, loader),
		createFunctionWithMultipleConfigsTestCase(providerCtx, loader),
		createFunctionWithAdvancedConfigsTestCase(providerCtx, loader),
		createFunctionWithAllCodeSourceFieldsTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		FunctionResource,
		&s.Suite,
	)
}

func createBasicFunctionCreateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := createLambdaServiceMock(
		WithCreateFunctionOutput(&lambda.CreateFunctionOutput{
			FunctionArn: aws.String(resourceARN),
		}),
	)

	// Create test data for function creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"description":  core.MappingNodeFromString("Test function"),
			"memorySize":   core.MappingNodeFromInt(128),
			"timeout":      core.MappingNodeFromInt(3),
			"runtime":      core.MappingNodeFromString("nodejs18.x"),
			"handler":      core.MappingNodeFromString("index.handler"),
			"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
			"code": {
				Fields: map[string]*core.MappingNode{
					"zipFile": core.MappingNodeFromString("console.log('Hello, World!');"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, Service]{
		Name: "create basic function",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-function-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-id",
					ResourceName: "TestFunction",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.description",
					},
					{
						FieldPath: "spec.memorySize",
					},
					{
						FieldPath: "spec.timeout",
					},
					{
						FieldPath: "spec.runtime",
					},
					{
						FieldPath: "spec.handler",
					},
					{
						FieldPath: "spec.role",
					},
					{
						FieldPath: "spec.code",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(resourceARN),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateFunction": &lambda.CreateFunctionInput{
				FunctionName: aws.String("test-function"),
				Description:  aws.String("Test function"),
				MemorySize:   aws.Int32(128),
				Timeout:      aws.Int32(3),
				Runtime:      types.Runtime("nodejs18.x"),
				Handler:      aws.String("index.handler"),
				Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
				Code: &types.FunctionCode{
					ZipFile: []byte("console.log('Hello, World!');"),
				},
			},
		},
	}
}

func createFunctionWithTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := createLambdaServiceMock(
		WithCreateFunctionOutput(&lambda.CreateFunctionOutput{
			FunctionArn: aws.String(resourceARN),
		}),
	)

	// Create test data for function creation with tags
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"description":  core.MappingNodeFromString("Test function"),
			"memorySize":   core.MappingNodeFromInt(128),
			"timeout":      core.MappingNodeFromInt(3),
			"runtime":      core.MappingNodeFromString("nodejs18.x"),
			"handler":      core.MappingNodeFromString("index.handler"),
			"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
			"code": {
				Fields: map[string]*core.MappingNode{
					"zipFile": core.MappingNodeFromString("console.log('Hello, World!');"),
				},
			},
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
							"value": core.MappingNodeFromString("test-project"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, Service]{
		Name: "create function with tags",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-function-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-id",
					ResourceName: "TestFunction",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.description",
					},
					{
						FieldPath: "spec.memorySize",
					},
					{
						FieldPath: "spec.timeout",
					},
					{
						FieldPath: "spec.runtime",
					},
					{
						FieldPath: "spec.handler",
					},
					{
						FieldPath: "spec.role",
					},
					{
						FieldPath: "spec.code",
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
				"spec.arn": core.MappingNodeFromString(resourceARN),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateFunction": &lambda.CreateFunctionInput{
				FunctionName: aws.String("test-function"),
				Description:  aws.String("Test function"),
				MemorySize:   aws.Int32(128),
				Timeout:      aws.Int32(3),
				Runtime:      types.Runtime("nodejs18.x"),
				Handler:      aws.String("index.handler"),
				Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
				Code: &types.FunctionCode{
					ZipFile: []byte("console.log('Hello, World!');"),
				},
				Tags: map[string]string{
					"Environment": "test",
					"Project":     "test-project",
				},
			},
		},
	}
}

func createFunctionWithSnapStartTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := createLambdaServiceMock(
		WithCreateFunctionOutput(&lambda.CreateFunctionOutput{
			FunctionArn: aws.String(resourceARN),
			SnapStart: &types.SnapStartResponse{
				ApplyOn:            types.SnapStartApplyOnPublishedVersions,
				OptimizationStatus: types.SnapStartOptimizationStatusOn,
			},
		}),
	)

	// Create test data for function creation with SnapStart
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"description":  core.MappingNodeFromString("Test function"),
			"memorySize":   core.MappingNodeFromInt(128),
			"timeout":      core.MappingNodeFromInt(3),
			"runtime":      core.MappingNodeFromString("nodejs18.x"),
			"handler":      core.MappingNodeFromString("index.handler"),
			"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
			"code": {
				Fields: map[string]*core.MappingNode{
					"zipFile": core.MappingNodeFromString("console.log('Hello, World!');"),
				},
			},
			"snapStart": {
				Fields: map[string]*core.MappingNode{
					"applyOn": core.MappingNodeFromString("PublishedVersions"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, Service]{
		Name: "create function with SnapStart",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-function-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-id",
					ResourceName: "TestFunction",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.description",
					},
					{
						FieldPath: "spec.memorySize",
					},
					{
						FieldPath: "spec.timeout",
					},
					{
						FieldPath: "spec.runtime",
					},
					{
						FieldPath: "spec.handler",
					},
					{
						FieldPath: "spec.role",
					},
					{
						FieldPath: "spec.code",
					},
					{
						FieldPath: "spec.snapStart",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(resourceARN),
				"spec.snapStartResponseApplyOn": core.MappingNodeFromString(
					string(types.SnapStartApplyOnPublishedVersions),
				),
				"spec.snapStartResponseOptimizationStatus": core.MappingNodeFromString(
					string(types.SnapStartOptimizationStatusOn),
				),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateFunction": &lambda.CreateFunctionInput{
				FunctionName: aws.String("test-function"),
				Description:  aws.String("Test function"),
				MemorySize:   aws.Int32(128),
				Timeout:      aws.Int32(3),
				Runtime:      types.Runtime("nodejs18.x"),
				Handler:      aws.String("index.handler"),
				Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
				Code: &types.FunctionCode{
					ZipFile: []byte("console.log('Hello, World!');"),
				},
				SnapStart: &types.SnapStart{
					ApplyOn: types.SnapStartApplyOnPublishedVersions,
				},
			},
		},
	}
}

func createFunctionFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, Service] {
	service := createLambdaServiceMock(
		WithCreateFunctionError(fmt.Errorf("failed to create function")),
	)

	// Create test data for function creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"description":  core.MappingNodeFromString("Test function"),
			"memorySize":   core.MappingNodeFromInt(128),
			"timeout":      core.MappingNodeFromInt(3),
			"runtime":      core.MappingNodeFromString("nodejs18.x"),
			"handler":      core.MappingNodeFromString("index.handler"),
			"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
			"code": {
				Fields: map[string]*core.MappingNode{
					"zipFile": core.MappingNodeFromString("console.log('Hello, World!');"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, Service]{
		Name: "create function failure",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-function-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-id",
					ResourceName: "TestFunction",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.description",
					},
					{
						FieldPath: "spec.memorySize",
					},
					{
						FieldPath: "spec.timeout",
					},
					{
						FieldPath: "spec.runtime",
					},
					{
						FieldPath: "spec.handler",
					},
					{
						FieldPath: "spec.role",
					},
					{
						FieldPath: "spec.code",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"CreateFunction": &lambda.CreateFunctionInput{
				FunctionName: aws.String("test-function"),
				Description:  aws.String("Test function"),
				MemorySize:   aws.Int32(128),
				Timeout:      aws.Int32(3),
				Runtime:      types.Runtime("nodejs18.x"),
				Handler:      aws.String("index.handler"),
				Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
				Code: &types.FunctionCode{
					ZipFile: []byte("console.log('Hello, World!');"),
				},
			},
		},
	}
}

func createFunctionWithMultipleConfigsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := createLambdaServiceMock(
		WithCreateFunctionOutput(&lambda.CreateFunctionOutput{
			FunctionArn: aws.String(resourceARN),
		}),
		WithPutFunctionCodeSigningConfigOutput(&lambda.PutFunctionCodeSigningConfigOutput{}),
		WithPutFunctionConcurrencyOutput(&lambda.PutFunctionConcurrencyOutput{}),
		WithPutFunctionRecursionConfigOutput(&lambda.PutFunctionRecursionConfigOutput{}),
		WithPutRuntimeManagementConfigOutput(&lambda.PutRuntimeManagementConfigOutput{}),
	)

	// Create test data for function creation with all configs
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"description":  core.MappingNodeFromString("Test function with all configs"),
			"memorySize":   core.MappingNodeFromInt(128),
			"timeout":      core.MappingNodeFromInt(3),
			"runtime":      core.MappingNodeFromString("nodejs18.x"),
			"handler":      core.MappingNodeFromString("index.handler"),
			"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
			"code": {
				Fields: map[string]*core.MappingNode{
					"zipFile": core.MappingNodeFromString("exports.handler = async (event) => { return { statusCode: 200, body: 'Hello from Lambda!' }; };"),
				},
			},
			"codeSigningConfigArn": core.MappingNodeFromString(
				"arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-12345678901234567",
			),
			"reservedConcurrentExecutions": core.MappingNodeFromInt(100),
			"recursiveLoop":                core.MappingNodeFromString("Terminate"),
			"runtimeManagementConfig": {
				Fields: map[string]*core.MappingNode{
					"updateRuntimeOn":   core.MappingNodeFromString("Manual"),
					"runtimeVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2::runtime:nodejs18.x"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, Service]{
		Name: "create function with all configs",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-function-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-id",
					ResourceName: "TestFunction",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.description",
					},
					{
						FieldPath: "spec.memorySize",
					},
					{
						FieldPath: "spec.timeout",
					},
					{
						FieldPath: "spec.runtime",
					},
					{
						FieldPath: "spec.handler",
					},
					{
						FieldPath: "spec.role",
					},
					{
						FieldPath: "spec.code",
					},
					{
						FieldPath: "spec.codeSigningConfigArn",
					},
					{
						FieldPath: "spec.reservedConcurrentExecutions",
					},
					{
						FieldPath: "spec.recursiveLoop",
					},
					{
						FieldPath: "spec.runtimeManagementConfig.updateRuntimeOn",
					},
					{
						FieldPath: "spec.runtimeManagementConfig.runtimeVersionArn",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(resourceARN),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateFunction": &lambda.CreateFunctionInput{
				FunctionName: aws.String("test-function"),
				Description:  aws.String("Test function with all configs"),
				MemorySize:   aws.Int32(128),
				Timeout:      aws.Int32(3),
				Runtime:      types.Runtime("nodejs18.x"),
				Handler:      aws.String("index.handler"),
				Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
				Code: &types.FunctionCode{
					ZipFile: []byte("exports.handler = async (event) => { return { statusCode: 200, body: 'Hello from Lambda!' }; };"),
				},
				CodeSigningConfigArn: aws.String(
					"arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-12345678901234567",
				),
			},
			"PutFunctionConcurrency": &lambda.PutFunctionConcurrencyInput{
				FunctionName:                 aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				ReservedConcurrentExecutions: aws.Int32(100),
			},
			"PutFunctionRecursionConfig": &lambda.PutFunctionRecursionConfigInput{
				FunctionName:  aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				RecursiveLoop: types.RecursiveLoopTerminate,
			},
			"PutRuntimeManagementConfig": &lambda.PutRuntimeManagementConfigInput{
				FunctionName:      aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				UpdateRuntimeOn:   types.UpdateRuntimeOnManual,
				RuntimeVersionArn: aws.String("arn:aws:lambda:us-west-2::runtime:nodejs18.x"),
			},
		},
	}
}

func createFunctionWithAdvancedConfigsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := createLambdaServiceMock(
		WithCreateFunctionOutput(&lambda.CreateFunctionOutput{
			FunctionArn: aws.String(resourceARN),
		}),
		WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &types.FunctionConfiguration{
				FunctionArn: aws.String(resourceARN),
			},
		}),
	)

	// Create test data for the function with advanced configurations
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"description":  core.MappingNodeFromString("Test function with all configurations"),
			"memorySize":   core.MappingNodeFromInt(512),
			"timeout":      core.MappingNodeFromInt(30),
			"runtime":      core.MappingNodeFromString("nodejs18.x"),
			"handler":      core.MappingNodeFromString("index.handler"),
			"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
			"code": {
				Fields: map[string]*core.MappingNode{
					"zipFile": core.MappingNodeFromString("exports.handler = async (event) => { return { statusCode: 200, body: 'Hello World!' }; };"),
				},
			},
			"tracingConfig": {
				Fields: map[string]*core.MappingNode{
					"mode": core.MappingNodeFromString("Active"),
				},
			},
			"recursiveLoop": core.MappingNodeFromString("Terminate"),
			"vpcConfig": {
				Fields: map[string]*core.MappingNode{
					"securityGroupIds":        core.MappingNodeFromStringSlice([]string{"sg-1234567890abcdef0"}),
					"subnetIds":               core.MappingNodeFromStringSlice([]string{"subnet-1234567890abcdef0"}),
					"ipv6AllowedForDualStack": core.MappingNodeFromBool(true),
				},
			},
			"fileSystemConfig": {
				Fields: map[string]*core.MappingNode{
					"arn":            core.MappingNodeFromString("arn:aws:elasticfilesystem:us-west-2:123456789012:access-point/fsap-1234567890abcdef0"),
					"localMountPath": core.MappingNodeFromString("/mnt/efs"),
				},
			},
			"deadLetterConfig": {
				Fields: map[string]*core.MappingNode{
					"targetArn": core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:test-dlq"),
				},
			},
			"environment": {
				Fields: map[string]*core.MappingNode{
					"variables": {
						Fields: map[string]*core.MappingNode{
							"ENV_VAR_1": core.MappingNodeFromString("value1"),
							"ENV_VAR_2": core.MappingNodeFromString("value2"),
						},
					},
				},
			},
			"ephemeralStorage": {
				Fields: map[string]*core.MappingNode{
					"size": core.MappingNodeFromInt(1024),
				},
			},
			"loggingConfig": {
				Fields: map[string]*core.MappingNode{
					"applicationLogLevel": core.MappingNodeFromString("INFO"),
					"logFormat":           core.MappingNodeFromString("JSON"),
					"logGroup":            core.MappingNodeFromString("/aws/lambda/test-function"),
					"systemLogLevel":      core.MappingNodeFromString("INFO"),
				},
			},
			"imageConfig": {
				Fields: map[string]*core.MappingNode{
					"command":          core.MappingNodeFromStringSlice([]string{"app.lambda_handler"}),
					"entryPoint":       core.MappingNodeFromStringSlice([]string{"/usr/local/bin/python"}),
					"workingDirectory": core.MappingNodeFromString("/var/task"),
				},
			},
			"codeSigningConfigArn": core.MappingNodeFromString(
				"arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef0",
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, Service]{
		Name: "create function with advanced configurations",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-function-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-id",
					ResourceName: "test-function",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function",
						},
						Spec: specData,
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(resourceARN),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateFunction": &lambda.CreateFunctionInput{
				FunctionName: aws.String("test-function"),
				Description:  aws.String("Test function with all configurations"),
				MemorySize:   aws.Int32(512),
				Timeout:      aws.Int32(30),
				Runtime:      types.Runtime("nodejs18.x"),
				Handler:      aws.String("index.handler"),
				Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
				Code: &types.FunctionCode{
					ZipFile: []byte("exports.handler = async (event) => { return { statusCode: 200, body: 'Hello World!' }; };"),
				},
				TracingConfig: &types.TracingConfig{
					Mode: types.TracingMode("Active"),
				},
				VpcConfig: &types.VpcConfig{
					SecurityGroupIds:        []string{"sg-1234567890abcdef0"},
					SubnetIds:               []string{"subnet-1234567890abcdef0"},
					Ipv6AllowedForDualStack: aws.Bool(true),
				},
				FileSystemConfigs: []types.FileSystemConfig{
					{
						Arn:            aws.String("arn:aws:elasticfilesystem:us-west-2:123456789012:access-point/fsap-1234567890abcdef0"),
						LocalMountPath: aws.String("/mnt/efs"),
					},
				},
				DeadLetterConfig: &types.DeadLetterConfig{
					TargetArn: aws.String("arn:aws:sqs:us-west-2:123456789012:test-dlq"),
				},
				Environment: &types.Environment{
					Variables: map[string]string{
						"ENV_VAR_1": "value1",
						"ENV_VAR_2": "value2",
					},
				},
				EphemeralStorage: &types.EphemeralStorage{
					Size: aws.Int32(1024),
				},
				LoggingConfig: &types.LoggingConfig{
					ApplicationLogLevel: types.ApplicationLogLevel("INFO"),
					LogFormat:           types.LogFormat("JSON"),
					LogGroup:            aws.String("/aws/lambda/test-function"),
					SystemLogLevel:      types.SystemLogLevel("INFO"),
				},
				ImageConfig: &types.ImageConfig{
					Command:          []string{"app.lambda_handler"},
					EntryPoint:       []string{"/usr/local/bin/python"},
					WorkingDirectory: aws.String("/var/task"),
				},
				CodeSigningConfigArn: aws.String(
					"arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1234567890abcdef0",
				),
			},
		},
	}
}

func createFunctionWithAllCodeSourceFieldsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, Service] {
	service := createLambdaServiceMock(
		WithCreateFunctionOutput(&lambda.CreateFunctionOutput{
			FunctionArn: aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
		}),
	)

	// Create test data for function creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
			"code": {
				Fields: map[string]*core.MappingNode{
					"imageUri":        core.MappingNodeFromString("123456789012.dkr.ecr.us-east-1.amazonaws.com/test-image:latest"),
					"s3Bucket":        core.MappingNodeFromString("test-bucket"),
					"s3Key":           core.MappingNodeFromString("test-key"),
					"s3ObjectVersion": core.MappingNodeFromString("test-version"),
					"sourceKMSKeyArn": core.MappingNodeFromString("arn:aws:kms:us-east-1:123456789012:key/test-key"),
				},
			},
			"handler":    core.MappingNodeFromString("index.handler"),
			"runtime":    core.MappingNodeFromString("nodejs18.x"),
			"memorySize": core.MappingNodeFromInt(512),
			"timeout":    core.MappingNodeFromInt(30),
			"layers": core.MappingNodeFromStringSlice([]string{
				"arn:aws:lambda:us-east-1:123456789012:layer:my-layer:1",
				"arn:aws:lambda:us-east-1:123456789012:layer:my-layer:2",
			}),
			"kmsKeyArn":   core.MappingNodeFromString("arn:aws:kms:us-east-1:123456789012:key/layer-key"),
			"packageType": core.MappingNodeFromString("Image"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, Service]{
		Name: "create function with all code source fields",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-function-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-id",
					ResourceName: "TestFunction",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.role",
					},
					{
						FieldPath: "spec.code",
					},
					{
						FieldPath: "spec.handler",
					},
					{
						FieldPath: "spec.runtime",
					},
					{
						FieldPath: "spec.memorySize",
					},
					{
						FieldPath: "spec.timeout",
					},
					{
						FieldPath: "spec.layers",
					},
					{
						FieldPath: "spec.kmsKeyArn",
					},
					{
						FieldPath: "spec.packageType",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateFunction": &lambda.CreateFunctionInput{
				FunctionName: aws.String("test-function"),
				Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
				Code: &types.FunctionCode{
					ImageUri:        aws.String("123456789012.dkr.ecr.us-east-1.amazonaws.com/test-image:latest"),
					S3Bucket:        aws.String("test-bucket"),
					S3Key:           aws.String("test-key"),
					S3ObjectVersion: aws.String("test-version"),
					SourceKMSKeyArn: aws.String("arn:aws:kms:us-east-1:123456789012:key/test-key"),
				},
				Handler:    aws.String("index.handler"),
				Runtime:    types.Runtime("nodejs18.x"),
				MemorySize: aws.Int32(512),
				Timeout:    aws.Int32(30),
				Layers: []string{
					"arn:aws:lambda:us-east-1:123456789012:layer:my-layer:1",
					"arn:aws:lambda:us-east-1:123456789012:layer:my-layer:2",
				},
				KMSKeyArn:   aws.String("arn:aws:kms:us-east-1:123456789012:key/layer-key"),
				PackageType: types.PackageType("Image"),
			},
		},
	}
}

func TestLambdaFunctionResourceCreate(t *testing.T) {
	suite.Run(t, new(LambdaFunctionResourceCreateSuite))
}
