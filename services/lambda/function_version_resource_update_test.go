//go:build unit

package lambda

import (
	"fmt"
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

type LambdaFunctionVersionResourceUpdateSuite struct {
	suite.Suite
}

func (s *LambdaFunctionVersionResourceUpdateSuite) Test_update_lambda_function_version() {
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
		createFunctionVersionNoUpdatesTestCase(providerCtx, loader),
		createFunctionVersionWithProvisionedConcurrencyUpdateTestCase(providerCtx, loader),
		createFunctionVersionWithRuntimePolicyUpdateTestCase(providerCtx, loader),
		createFunctionVersionUpdateFailureTestCase(providerCtx, loader),
		recreateFunctionVersionOnFunctionNameOrRuntimeVersionArnChangeTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		FunctionVersionResource,
		&s.Suite,
	)
}

func createFunctionVersionNoUpdatesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"
	version := "1"
	resourceARNWithVersion := resourceARN + ":" + version

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &types.FunctionConfiguration{
				FunctionArn: aws.String(resourceARN),
				Version:     aws.String(version),
			},
		}),
	)

	// Create test data for function version with no updates
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":           core.MappingNodeFromString("test-function"),
			"functionArn":            core.MappingNodeFromString(resourceARN),
			"version":                core.MappingNodeFromString(version),
			"functionArnWithVersion": core.MappingNodeFromString(resourceARNWithVersion),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "no updates",
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
			ResourceID: "test-function-version-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-version-id",
					ResourceName: "TestFunctionVersion",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-version-id",
						Name:       "TestFunctionVersion",
						InstanceID: "test-instance-id",
						SpecData:   specData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function_version",
						},
						Spec: specData,
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionArn":            core.MappingNodeFromString(resourceARN),
				"spec.version":                core.MappingNodeFromString(version),
				"spec.functionArnWithVersion": core.MappingNodeFromString(resourceARNWithVersion),
			},
		},
	}
}

func createFunctionVersionWithProvisionedConcurrencyUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"
	version := "1"
	resourceARNWithVersion := resourceARN + ":" + version

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &types.FunctionConfiguration{
				FunctionArn: aws.String(resourceARN),
				Version:     aws.String(version),
			},
		}),
	)

	// Create test data for function version with provisioned concurrency update
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":           core.MappingNodeFromString("test-function"),
			"functionArn":            core.MappingNodeFromString(resourceARN),
			"version":                core.MappingNodeFromString(version),
			"functionArnWithVersion": core.MappingNodeFromString(resourceARNWithVersion),
			"provisionedConcurrencyConfig": {
				Fields: map[string]*core.MappingNode{
					"provisionedConcurrentExecutions": core.MappingNodeFromInt(100),
				},
			},
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":           core.MappingNodeFromString("test-function"),
			"functionArn":            core.MappingNodeFromString(resourceARN),
			"version":                core.MappingNodeFromString(version),
			"functionArnWithVersion": core.MappingNodeFromString(resourceARNWithVersion),
			"provisionedConcurrencyConfig": {
				Fields: map[string]*core.MappingNode{
					"provisionedConcurrentExecutions": core.MappingNodeFromInt(200),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update function version with provisioned concurrency",
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
			ResourceID: "test-function-version-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-version-id",
					ResourceName: "TestFunctionVersion",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-version-id",
						Name:       "TestFunctionVersion",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function_version",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.provisionedConcurrencyConfig.provisionedConcurrentExecutions",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionArn":            core.MappingNodeFromString(resourceARN),
				"spec.version":                core.MappingNodeFromString(version),
				"spec.functionArnWithVersion": core.MappingNodeFromString(resourceARNWithVersion),
			},
		},
		SaveActionsCalled: map[string]any{
			"PutProvisionedConcurrencyConfig": &lambda.PutProvisionedConcurrencyConfigInput{
				FunctionName:                    aws.String(resourceARN),
				Qualifier:                       aws.String(version),
				ProvisionedConcurrentExecutions: aws.Int32(200),
			},
		},
	}
}

func createFunctionVersionWithRuntimePolicyUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"
	version := "1"
	resourceARNWithVersion := resourceARN + ":" + version

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &types.FunctionConfiguration{
				FunctionArn: aws.String(resourceARN),
				Version:     aws.String(version),
			},
		}),
	)

	// Create test data for function version with runtime policy update
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":           core.MappingNodeFromString("test-function"),
			"functionArn":            core.MappingNodeFromString(resourceARN),
			"version":                core.MappingNodeFromString(version),
			"functionArnWithVersion": core.MappingNodeFromString(resourceARNWithVersion),
			"runtimePolicy": {
				Fields: map[string]*core.MappingNode{
					"updateRuntimeOn": core.MappingNodeFromString("Auto"),
				},
			},
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":           core.MappingNodeFromString("test-function"),
			"functionArn":            core.MappingNodeFromString(resourceARN),
			"version":                core.MappingNodeFromString(version),
			"functionArnWithVersion": core.MappingNodeFromString(resourceARNWithVersion),
			"runtimePolicy": {
				Fields: map[string]*core.MappingNode{
					"updateRuntimeOn": core.MappingNodeFromString("Manual"),
					"runtimeVersionArn": core.MappingNodeFromString(
						"arn:aws:lambda:us-west-2::runtime:nodejs18.x",
					),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update function version with runtime policy",
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
			ResourceID: "test-function-version-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-version-id",
					ResourceName: "TestFunctionVersion",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-version-id",
						Name:       "TestFunctionVersion",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function_version",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.runtimePolicy.updateRuntimeOn",
					},
					{
						FieldPath: "spec.runtimePolicy.runtimeVersionArn",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionArn":            core.MappingNodeFromString(resourceARN),
				"spec.version":                core.MappingNodeFromString(version),
				"spec.functionArnWithVersion": core.MappingNodeFromString(resourceARNWithVersion),
			},
		},
		SaveActionsCalled: map[string]any{
			"PutRuntimeManagementConfig": &lambda.PutRuntimeManagementConfigInput{
				FunctionName:      aws.String(resourceARN),
				Qualifier:         aws.String(version),
				UpdateRuntimeOn:   types.UpdateRuntimeOn("Manual"),
				RuntimeVersionArn: aws.String("arn:aws:lambda:us-west-2::runtime:nodejs18.x"),
			},
		},
	}
}

func createFunctionVersionUpdateFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"
	version := "1"
	resourceARNWithVersion := resourceARN + ":" + version

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
			Configuration: &types.FunctionConfiguration{
				FunctionArn: aws.String(resourceARN),
				Version:     aws.String(version),
			},
		}),
		lambdamock.WithPutProvisionedConcurrencyConfigError(fmt.Errorf("failed to update provisioned concurrency")),
	)

	// Create test data for function version update failure
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":           core.MappingNodeFromString("test-function"),
			"functionArn":            core.MappingNodeFromString(resourceARN),
			"version":                core.MappingNodeFromString(version),
			"functionArnWithVersion": core.MappingNodeFromString(resourceARNWithVersion),
			"provisionedConcurrencyConfig": {
				Fields: map[string]*core.MappingNode{
					"provisionedConcurrentExecutions": core.MappingNodeFromInt(100),
				},
			},
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":           core.MappingNodeFromString("test-function"),
			"functionArn":            core.MappingNodeFromString(resourceARN),
			"version":                core.MappingNodeFromString(version),
			"functionArnWithVersion": core.MappingNodeFromString(resourceARNWithVersion),
			"provisionedConcurrencyConfig": {
				Fields: map[string]*core.MappingNode{
					"provisionedConcurrentExecutions": core.MappingNodeFromInt(200),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update function version failure",
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
			ResourceID: "test-function-version-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-version-id",
					ResourceName: "TestFunctionVersion",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-version-id",
						Name:       "TestFunctionVersion",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function_version",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.provisionedConcurrencyConfig.provisionedConcurrentExecutions",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"PutProvisionedConcurrencyConfig": &lambda.PutProvisionedConcurrencyConfigInput{
				FunctionName:                    aws.String(resourceARN),
				Qualifier:                       aws.String(version),
				ProvisionedConcurrentExecutions: aws.Int32(200),
			},
		},
	}
}

func recreateFunctionVersionOnFunctionNameOrRuntimeVersionArnChangeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	oldResourceARN := "arn:aws:lambda:us-west-2:123456789012:function:old-function"
	newResourceARN := "arn:aws:lambda:us-west-2:123456789012:function:new-function"
	version := "1"
	oldResourceARNWithVersion := oldResourceARN + ":" + version
	newResourceARNWithVersion := newResourceARN + ":" + version

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithPublishVersionOutput(&lambda.PublishVersionOutput{
			FunctionArn: aws.String(newResourceARNWithVersion),
			Version:     aws.String(version),
		}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn":            core.MappingNodeFromString(oldResourceARN),
			"functionArnWithVersion": core.MappingNodeFromString(oldResourceARNWithVersion),
			"functionName":           core.MappingNodeFromString("old-function"),
			"version":                core.MappingNodeFromString(version),
			"runtimeVersionArn":      core.MappingNodeFromString("arn:aws:lambda:us-west-2::runtime:nodejs18.x"),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":      core.MappingNodeFromString("new-function"),
			"runtimeVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2::runtime:nodejs22.x"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "recreate function version on functionName or runtimeVersionArn change",
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
			ResourceID: "test-function-version-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-version-id",
					ResourceName: "TestFunctionVersion",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-version-id",
						Name:       "TestFunctionVersion",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function_version",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
						PrevValue: core.MappingNodeFromString("old-function"),
						NewValue:  core.MappingNodeFromString("new-function"),
					},
					{
						FieldPath: "spec.runtimeVersionArn",
						PrevValue: core.MappingNodeFromString("arn:aws:lambda:us-west-2::runtime:nodejs18.x"),
						NewValue:  core.MappingNodeFromString("arn:aws:lambda:us-west-2::runtime:nodejs22.x"),
					},
				},
				MustRecreate: true,
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionArn":            core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:new-function:1"),
				"spec.functionArnWithVersion": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:new-function:1:1"),
				"spec.version":                core.MappingNodeFromString(version),
			},
		},
		SaveActionsCalled: map[string]any{
			"PublishVersion": &lambda.PublishVersionInput{
				FunctionName: aws.String("new-function"),
			},
		},
	}
}

func TestLambdaFunctionVersionResourceUpdate(t *testing.T) {
	suite.Run(t, new(LambdaFunctionVersionResourceUpdateSuite))
}
