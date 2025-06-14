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
	"github.com/newstack-cloud/celerity/libs/blueprint/state"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaFunctionResourceUpdateSuite struct {
	suite.Suite
}

func (s *LambdaFunctionResourceUpdateSuite) Test_update_lambda_function() {
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
		createBasicFunctionUpdateTestCase(providerCtx, loader),
		createNoUpdatesTestCase(providerCtx, loader),
		createFunctionConfigAndCodeUpdateTestCase(providerCtx, loader),
		createMultipleConfigsUpdateTestCase(providerCtx, loader),
		createUpdateFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		FunctionResource,
		&s.Suite,
	)
}

func createBasicFunctionUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := createLambdaServiceMock(
		WithGetFunctionOutput(&lambda.GetFunctionOutput{
			// We only need to set the fields that are marked as "computed"
			// in the schema.
			Configuration: &types.FunctionConfiguration{
				FunctionArn: aws.String(resourceARN),
				SnapStart: &types.SnapStartResponse{
					ApplyOn:            types.SnapStartApplyOnPublishedVersions,
					OptimizationStatus: types.SnapStartOptimizationStatusOn,
				},
			},
		}),
	)

	// Create test data for function configuration updates
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":         core.MappingNodeFromString(resourceARN),
			"description": core.MappingNodeFromString("Original description"),
			"memorySize":  core.MappingNodeFromInt(128),
			"timeout":     core.MappingNodeFromInt(3),
			"runtime":     core.MappingNodeFromString("nodejs18.x"),
			"handler":     core.MappingNodeFromString("index.handler"),
			"imageConfig": {
				Fields: map[string]*core.MappingNode{
					"command": core.MappingNodeFromStringSlice(
						[]string{"node", "index.js"},
					),
					"entryPoint": core.MappingNodeFromStringSlice(
						[]string{"/bin/sh", "-c"},
					),
					"workingDirectory": core.MappingNodeFromString("/app"),
				},
			},
			"loggingConfig": {
				Fields: map[string]*core.MappingNode{
					"applicationLogLevel": core.MappingNodeFromString("INFO"),
					"logFormat":           core.MappingNodeFromString("JSON"),
					"logGroup":            core.MappingNodeFromString("test-log-group"),
					"systemLogLevel":      core.MappingNodeFromString("DEBUG"),
				},
			},
			"fileSystemConfig": {
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString(
						"arn:aws:elasticfilesystem:us-west-2:123456789012:access-point/fsap-12345678901234567",
					),
					"localMountPath": core.MappingNodeFromString("/mnt/test-efs"),
				},
			},
			// No vpc config in current state.
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"description":  core.MappingNodeFromString("Updated description"),
			"memorySize":   core.MappingNodeFromInt(256),
			"timeout":      core.MappingNodeFromInt(5),
			"runtime":      core.MappingNodeFromString("nodejs22.x"),
			"handler":      core.MappingNodeFromString("index.newHandler"),
			"imageConfig": {
				Fields: map[string]*core.MappingNode{
					"command": core.MappingNodeFromStringSlice(
						[]string{"node", "index2.js"},
					),
					"entryPoint": core.MappingNodeFromStringSlice(
						[]string{"/bin/sh", "-c"},
					),
					"workingDirectory": core.MappingNodeFromString("/app2"),
				},
			},
			"loggingConfig": {
				Fields: map[string]*core.MappingNode{
					"applicationLogLevel": core.MappingNodeFromString("DEBUG"),
					"logFormat":           core.MappingNodeFromString("Text"),
					"logGroup":            core.MappingNodeFromString("test-log-group-2"),
					"systemLogLevel":      core.MappingNodeFromString("INFO"),
				},
			},
			"fileSystemConfig": {
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString(
						"arn:aws:elasticfilesystem:us-west-2:123456789012:access-point/fsap-78901234567890123",
					),
					"localMountPath": core.MappingNodeFromString("/mnt/test-efs-2"),
				},
			},
			"vpcConfig": {
				Fields: map[string]*core.MappingNode{
					"securityGroupIds": core.MappingNodeFromStringSlice(
						[]string{"sg-12345678901234567"},
					),
					"subnetIds": core.MappingNodeFromStringSlice(
						[]string{"subnet-12345678901234567"},
					),
					"ipv6AllowedForDualStack": core.MappingNodeFromBool(true),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, Service]{
		Name: "update function configuration",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-id",
						Name:       "TestFunction",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
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
						FieldPath: "spec.imageConfig.command",
					},
					{
						FieldPath: "spec.imageConfig.workingDirectory",
					},
					{
						FieldPath: "spec.loggingConfig.applicationLogLevel",
					},
					{
						FieldPath: "spec.loggingConfig.logFormat",
					},
					{
						FieldPath: "spec.loggingConfig.logGroup",
					},
					{
						FieldPath: "spec.loggingConfig.systemLogLevel",
					},
					{
						FieldPath: "spec.fileSystemConfig.arn",
					},
					{
						FieldPath: "spec.fileSystemConfig.localMountPath",
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.vpcConfig.securityGroupIds",
					},
					{
						FieldPath: "spec.vpcConfig.subnetIds",
					},
					{
						FieldPath: "spec.vpcConfig.ipv6AllowedForDualStack",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":                                 core.MappingNodeFromString(resourceARN),
				"spec.snapStartResponseApplyOn":            core.MappingNodeFromString("PublishedVersions"),
				"spec.snapStartResponseOptimizationStatus": core.MappingNodeFromString("On"),
			},
		},
		SaveActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(resourceARN),
				Description:  aws.String("Updated description"),
				MemorySize:   aws.Int32(256),
				Timeout:      aws.Int32(5),
				Runtime:      types.RuntimeNodejs22x,
				Handler:      aws.String("index.newHandler"),
				ImageConfig: &types.ImageConfig{
					Command:          []string{"node", "index2.js"},
					WorkingDirectory: aws.String("/app2"),
				},
				LoggingConfig: &types.LoggingConfig{
					ApplicationLogLevel: types.ApplicationLogLevelDebug,
					LogFormat:           types.LogFormatText,
					LogGroup:            aws.String("test-log-group-2"),
					SystemLogLevel:      types.SystemLogLevelInfo,
				},
				FileSystemConfigs: []types.FileSystemConfig{
					{
						Arn: aws.String(
							"arn:aws:elasticfilesystem:us-west-2:123456789012:access-point/fsap-78901234567890123",
						),
						LocalMountPath: aws.String("/mnt/test-efs-2"),
					},
				},
				VpcConfig: &types.VpcConfig{
					SecurityGroupIds:        []string{"sg-12345678901234567"},
					SubnetIds:               []string{"subnet-12345678901234567"},
					Ipv6AllowedForDualStack: aws.Bool(true),
				},
			},
		},
		SaveActionsNotCalled: []string{
			"UpdateFunctionCode",
			"PutFunctionCodeSigningConfig",
			"PutFunctionConcurrency",
			"PutFunctionRecursionConfig",
			"PutRuntimeManagementConfig",
			"TagResource",
			"UntagResource",
		},
		ExpectError: false,
	}
}

func createNoUpdatesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := createLambdaServiceMock(
		WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &types.FunctionConfiguration{
				FunctionArn: aws.String(resourceARN),
				SnapStart: &types.SnapStartResponse{
					ApplyOn:            types.SnapStartApplyOnPublishedVersions,
					OptimizationStatus: types.SnapStartOptimizationStatusOn,
				},
			},
		}),
	)
	// Create test data with no changes
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                                 core.MappingNodeFromString(resourceARN),
			"snapStartResponseApplyOn":            core.MappingNodeFromString("PublishedVersions"),
			"snapStartResponseOptimizationStatus": core.MappingNodeFromString("On"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, Service]{
		Name: "no updates",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-id",
						Name:       "TestFunction",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function",
						},
						Spec: currentStateSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":                                 core.MappingNodeFromString(resourceARN),
				"spec.snapStartResponseApplyOn":            core.MappingNodeFromString("PublishedVersions"),
				"spec.snapStartResponseOptimizationStatus": core.MappingNodeFromString("On"),
			},
		},
		SaveActionsNotCalled: []string{
			"UpdateFunctionConfiguration",
			"UpdateFunctionCode",
			"PutFunctionCodeSigningConfig",
			"PutFunctionConcurrency",
			"PutFunctionRecursionConfig",
			"PutRuntimeManagementConfig",
			"TagResource",
			"UntagResource",
		},
		ExpectError: false,
	}
}

func createFunctionConfigAndCodeUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := createLambdaServiceMock(
		WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &types.FunctionConfiguration{
				FunctionArn: aws.String(resourceARN),
				SnapStart: &types.SnapStartResponse{
					ApplyOn:            types.SnapStartApplyOnPublishedVersions,
					OptimizationStatus: types.SnapStartOptimizationStatusOn,
				},
			},
		}),
	)

	// Create test data for current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":        core.MappingNodeFromString(resourceARN),
			"memorySize": core.MappingNodeFromInt(128),
			"runtime":    core.MappingNodeFromString("nodejs18.x"),
			"code": {
				Fields: map[string]*core.MappingNode{
					"zipFile": core.MappingNodeFromString("old code"),
				},
			},
		},
	}

	// Create test data for updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":        core.MappingNodeFromString(resourceARN),
			"runtime":    core.MappingNodeFromString("nodejs22.x"),
			"memorySize": core.MappingNodeFromInt(256),
			"code": {
				Fields: map[string]*core.MappingNode{
					"zipFile": core.MappingNodeFromString("new code"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, Service]{
		Name: "update function configuration and code",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-id",
						Name:       "TestFunction",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.memorySize",
					},
					{
						FieldPath: "spec.code.zipFile",
					},
					{
						FieldPath: "spec.runtime",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":                                 core.MappingNodeFromString(resourceARN),
				"spec.snapStartResponseApplyOn":            core.MappingNodeFromString("PublishedVersions"),
				"spec.snapStartResponseOptimizationStatus": core.MappingNodeFromString("On"),
			},
		},
		SaveActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(resourceARN),
				MemorySize:   aws.Int32(256),
				Runtime:      types.RuntimeNodejs22x,
			},
			"UpdateFunctionCode": &lambda.UpdateFunctionCodeInput{
				FunctionName: aws.String(resourceARN),
				ZipFile:      []byte("new code"),
				Publish:      true,
			},
			"GetFunction": &lambda.GetFunctionInput{
				FunctionName: aws.String(resourceARN),
			},
		},
		SaveActionsNotCalled: []string{
			"PutFunctionCodeSigningConfig",
			"PutFunctionConcurrency",
			"PutFunctionRecursionConfig",
			"PutRuntimeManagementConfig",
			"TagResource",
			"UntagResource",
		},
	}
}

func createMultipleConfigsUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := createLambdaServiceMock(
		WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &types.FunctionConfiguration{
				FunctionArn: aws.String(resourceARN),
				SnapStart: &types.SnapStartResponse{
					ApplyOn:            types.SnapStartApplyOnPublishedVersions,
					OptimizationStatus: types.SnapStartOptimizationStatusOn,
				},
			},
		}),
		WithPutFunctionCodeSigningConfigOutput(&lambda.PutFunctionCodeSigningConfigOutput{}),
		WithPutFunctionConcurrencyOutput(&lambda.PutFunctionConcurrencyOutput{}),
		WithPutFunctionRecursionConfigOutput(&lambda.PutFunctionRecursionConfigOutput{}),
		WithPutRuntimeManagementConfigOutput(&lambda.PutRuntimeManagementConfigOutput{}),
	)

	// Create test data for current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(resourceARN),
			"codeSigningConfig": {
				Fields: map[string]*core.MappingNode{
					"codeSigningConfigArn": core.MappingNodeFromString(
						"arn:aws:lambda:us-west-2:123456789012:code-signing-config:old-config",
					),
				},
			},
			"reservedConcurrentExecutions": core.MappingNodeFromInt(10),
			"recursiveLoop":                core.MappingNodeFromString("DISABLED"),
			"runtimeManagementConfig": {
				Fields: map[string]*core.MappingNode{
					"updateRuntimeOn": core.MappingNodeFromString("AUTO"),
				},
			},
		},
	}

	// Create test data for updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(resourceARN),
			"codeSigningConfig": {
				Fields: map[string]*core.MappingNode{
					"codeSigningConfigArn": core.MappingNodeFromString(
						"arn:aws:lambda:us-west-2:123456789012:code-signing-config:new-config",
					),
				},
			},
			"reservedConcurrentExecutions": core.MappingNodeFromInt(20),
			"recursiveLoop":                core.MappingNodeFromString("ENABLED"),
			"runtimeManagementConfig": {
				Fields: map[string]*core.MappingNode{
					"updateRuntimeOn": core.MappingNodeFromString("MANUAL"),
					"runtimeVersionArn": core.MappingNodeFromString(
						"arn:aws:lambda:us-west-2:123456789012:runtime:nodejs18.x:v1",
					),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, Service]{
		Name: "update multiple configurations",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-id",
						Name:       "TestFunction",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.codeSigningConfig.codeSigningConfigArn",
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
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.runtimeManagementConfig.runtimeVersionArn",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":                                 core.MappingNodeFromString(resourceARN),
				"spec.snapStartResponseApplyOn":            core.MappingNodeFromString("PublishedVersions"),
				"spec.snapStartResponseOptimizationStatus": core.MappingNodeFromString("On"),
			},
		},
		SaveActionsCalled: map[string]any{
			"PutFunctionCodeSigningConfig": &lambda.PutFunctionCodeSigningConfigInput{
				FunctionName:         aws.String(resourceARN),
				CodeSigningConfigArn: aws.String("arn:aws:lambda:us-west-2:123456789012:code-signing-config:new-config"),
			},
			"PutFunctionConcurrency": &lambda.PutFunctionConcurrencyInput{
				FunctionName:                 aws.String(resourceARN),
				ReservedConcurrentExecutions: aws.Int32(20),
			},
			"PutFunctionRecursionConfig": &lambda.PutFunctionRecursionConfigInput{
				FunctionName:  aws.String(resourceARN),
				RecursiveLoop: types.RecursiveLoop("ENABLED"),
			},
			"PutRuntimeManagementConfig": &lambda.PutRuntimeManagementConfigInput{
				FunctionName:      aws.String(resourceARN),
				UpdateRuntimeOn:   types.UpdateRuntimeOn("MANUAL"),
				RuntimeVersionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:runtime:nodejs18.x:v1"),
			},
			"GetFunction": &lambda.GetFunctionInput{
				FunctionName: aws.String(resourceARN),
			},
		},
		SaveActionsNotCalled: []string{
			"UpdateFunctionConfiguration",
			"UpdateFunctionCode",
			"TagResource",
			"UntagResource",
		},
	}
}

func createUpdateFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := createLambdaServiceMock(
		WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &types.FunctionConfiguration{
				FunctionArn: aws.String(resourceARN),
			},
		}),
		WithUpdateFunctionConfigurationError(fmt.Errorf("update function configuration failed")),
	)

	// Create test data for current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":        core.MappingNodeFromString(resourceARN),
			"memorySize": core.MappingNodeFromInt(128),
		},
	}

	// Create test data for updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":        core.MappingNodeFromString(resourceARN),
			"memorySize": core.MappingNodeFromInt(256),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, Service]{
		Name: "update function configuration failure",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-id",
						Name:       "TestFunction",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.memorySize",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"UpdateFunctionConfiguration": &lambda.UpdateFunctionConfigurationInput{
				FunctionName: aws.String(resourceARN),
				MemorySize:   aws.Int32(256),
			},
		},
		SaveActionsNotCalled: []string{
			"UpdateFunctionCode",
			"PutFunctionCodeSigningConfig",
			"PutFunctionConcurrency",
			"PutFunctionRecursionConfig",
			"PutRuntimeManagementConfig",
			"TagResource",
			"UntagResource",
		},
	}
}

func TestLambdaFunctionResourceUpdate(t *testing.T) {
	suite.Run(t, new(LambdaFunctionResourceUpdateSuite))
}
