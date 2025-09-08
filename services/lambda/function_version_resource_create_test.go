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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaFunctionVersionResourceCreateSuite struct {
	suite.Suite
}

func (s *LambdaFunctionVersionResourceCreateSuite) Test_create_lambda_function_version() {
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
		createBasicFunctionVersionTestCase(providerCtx, loader),
		createFunctionVersionWithDescriptionTestCase(providerCtx, loader),
		createFunctionVersionWithProvisionedConcurrencyTestCase(providerCtx, loader),
		createFunctionVersionWithRuntimePolicyTestCase(providerCtx, loader),
		createFunctionVersionFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		FunctionVersionResource,
		&s.Suite,
	)
}

func createBasicFunctionVersionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"
	version := "1"
	resourceARNWithVersion := resourceARN + ":" + version

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithPublishVersionOutput(&lambda.PublishVersionOutput{
			FunctionArn: aws.String(resourceARN),
			Version:     aws.String(version),
		}),
	)

	// Create test data for function version creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create basic function version",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function_version",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
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
			"PublishVersion": &lambda.PublishVersionInput{
				FunctionName: aws.String("test-function"),
			},
		},
	}
}

func createFunctionVersionWithDescriptionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"
	version := "1"
	resourceARNWithVersion := resourceARN + ":" + version

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithPublishVersionOutput(&lambda.PublishVersionOutput{
			FunctionArn: aws.String(resourceARN),
			Version:     aws.String(version),
		}),
	)

	// Create test data for function version creation with description
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"description":  core.MappingNodeFromString("Test function version"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create function version with description",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function_version",
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
			"PublishVersion": &lambda.PublishVersionInput{
				FunctionName: aws.String("test-function"),
				Description:  aws.String("Test function version"),
			},
		},
	}
}

func createFunctionVersionWithProvisionedConcurrencyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"
	version := "1"
	resourceARNWithVersion := resourceARN + ":" + version

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithPublishVersionOutput(&lambda.PublishVersionOutput{
			FunctionArn: aws.String(resourceARN),
			Version:     aws.String(version),
		}),
		lambdamock.WithPutProvisionedConcurrencyConfigOutput(&lambda.PutProvisionedConcurrencyConfigOutput{}),
	)

	// Create test data for function version creation with provisioned concurrency
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"provisionedConcurrencyConfig": {
				Fields: map[string]*core.MappingNode{
					"provisionedConcurrentExecutions": core.MappingNodeFromInt(100),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create function version with provisioned concurrency",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function_version",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.provisionedConcurrencyConfig",
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
			"PublishVersion": &lambda.PublishVersionInput{
				FunctionName: aws.String("test-function"),
			},
			"PutProvisionedConcurrencyConfig": &lambda.PutProvisionedConcurrencyConfigInput{
				FunctionName:                    aws.String(resourceARN),
				Qualifier:                       aws.String(version),
				ProvisionedConcurrentExecutions: aws.Int32(100),
			},
		},
	}
}

func createFunctionVersionWithRuntimePolicyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	resourceARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"
	version := "1"
	resourceARNWithVersion := resourceARN + ":" + version

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithPublishVersionOutput(&lambda.PublishVersionOutput{
			FunctionArn: aws.String(resourceARN),
			Version:     aws.String(version),
		}),
		lambdamock.WithPutRuntimeManagementConfigOutput(&lambda.PutRuntimeManagementConfigOutput{}),
	)

	// Create test data for function version creation with runtime policy
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"runtimePolicy": {
				Fields: map[string]*core.MappingNode{
					"updateRuntimeOn": core.MappingNodeFromString("Auto"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create function version with runtime policy",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function_version",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.runtimePolicy.updateRuntimeOn",
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
			"PublishVersion": &lambda.PublishVersionInput{
				FunctionName: aws.String("test-function"),
			},
			"PutRuntimeManagementConfig": &lambda.PutRuntimeManagementConfigInput{
				FunctionName:    aws.String(resourceARN),
				Qualifier:       aws.String(version),
				UpdateRuntimeOn: types.UpdateRuntimeOn("Auto"),
			},
		},
	}
}

func createFunctionVersionFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithPublishVersionError(fmt.Errorf("failed to publish version")),
	)

	// Create test data for function version creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create function version failure",
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
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/function_version",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"PublishVersion": &lambda.PublishVersionInput{
				FunctionName: aws.String("test-function"),
			},
		},
	}
}

func TestLambdaFunctionVersionResourceCreate(t *testing.T) {
	suite.Run(t, new(LambdaFunctionVersionResourceCreateSuite))
}
