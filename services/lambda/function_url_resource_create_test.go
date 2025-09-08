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

type LambdaFunctionUrlResourceCreateSuite struct {
	suite.Suite
}

func (s *LambdaFunctionUrlResourceCreateSuite) Test_create_lambda_function_url() {
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
		createBasicFunctionUrlTestCase(providerCtx, loader),
		createFunctionUrlWithCorsTestCase(providerCtx, loader),
		createFunctionUrlWithAuthTypeTestCase(providerCtx, loader),
		createFunctionUrlWithInvokeModeTestCase(providerCtx, loader),
		createFunctionUrlWithQualifierTestCase(providerCtx, loader),
		createComplexFunctionUrlTestCase(providerCtx, loader),
		createFunctionUrlFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		FunctionUrlResource,
		&s.Suite,
	)
}

func createBasicFunctionUrlTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	functionUrl := "https://test-function-url.lambda-url.us-west-2.on.aws/"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateFunctionUrlConfigOutput(&lambda.CreateFunctionUrlConfigOutput{
			FunctionUrl: aws.String(functionUrl),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create basic function URL",
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
			ResourceID: "test-function-url-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-url-id",
					ResourceName: "TestFunctionUrl",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionArn",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionUrl": core.MappingNodeFromString(functionUrl),
				"spec.functionArn": core.MappingNodeFromString(""),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateFunctionUrlConfig": &lambda.CreateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				AuthType:     "",
			},
		},
	}
}

func createFunctionUrlWithCorsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	functionUrl := "https://test-function-url.lambda-url.us-west-2.on.aws/"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateFunctionUrlConfigOutput(&lambda.CreateFunctionUrlConfigOutput{
			FunctionUrl: aws.String(functionUrl),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"cors": {
				Fields: map[string]*core.MappingNode{
					"allowCredentials": core.MappingNodeFromBool(true),
					"allowHeaders": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("Content-Type"),
							core.MappingNodeFromString("Authorization"),
						},
					},
					"allowMethods": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("GET"),
							core.MappingNodeFromString("POST"),
						},
					},
					"allowOrigins": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("https://example.com"),
						},
					},
					"exposeHeaders": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("X-Custom-Header"),
						},
					},
					"maxAge": core.MappingNodeFromInt(3600),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create function URL with CORS",
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
			ResourceID: "test-function-url-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-url-id",
					ResourceName: "TestFunctionUrl",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionArn",
					},
					{
						FieldPath: "spec.cors",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionUrl": core.MappingNodeFromString(functionUrl),
				"spec.functionArn": core.MappingNodeFromString(""),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateFunctionUrlConfig": &lambda.CreateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				AuthType:     "",
				Cors: &types.Cors{
					AllowCredentials: aws.Bool(true),
					AllowHeaders:     []string{"Content-Type", "Authorization"},
					AllowMethods:     []string{"GET", "POST"},
					AllowOrigins:     []string{"https://example.com"},
					ExposeHeaders:    []string{"X-Custom-Header"},
					MaxAge:           aws.Int32(3600),
				},
			},
		},
	}
}

func createFunctionUrlWithAuthTypeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	functionUrl := "https://test-function-url.lambda-url.us-west-2.on.aws/"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateFunctionUrlConfigOutput(&lambda.CreateFunctionUrlConfigOutput{
			FunctionUrl: aws.String(functionUrl),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"authType":    core.MappingNodeFromString("AWS_IAM"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create function URL with auth type",
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
			ResourceID: "test-function-url-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-url-id",
					ResourceName: "TestFunctionUrl",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionArn",
					},
					{
						FieldPath: "spec.authType",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionUrl": core.MappingNodeFromString(functionUrl),
				"spec.functionArn": core.MappingNodeFromString(""),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateFunctionUrlConfig": &lambda.CreateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				AuthType:     "AWS_IAM",
			},
		},
	}
}

func createFunctionUrlWithInvokeModeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	functionUrl := "https://test-function-url.lambda-url.us-west-2.on.aws/"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateFunctionUrlConfigOutput(&lambda.CreateFunctionUrlConfigOutput{
			FunctionUrl: aws.String(functionUrl),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"invokeMode":  core.MappingNodeFromString("BUFFERED"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create function URL with invoke mode",
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
			ResourceID: "test-function-url-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-url-id",
					ResourceName: "TestFunctionUrl",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionArn",
					},
					{
						FieldPath: "spec.invokeMode",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionUrl": core.MappingNodeFromString(functionUrl),
				"spec.functionArn": core.MappingNodeFromString(""),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateFunctionUrlConfig": &lambda.CreateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				AuthType:     "",
				InvokeMode:   types.InvokeModeBuffered,
			},
		},
	}
}

func createFunctionUrlWithQualifierTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	functionUrl := "https://test-function-url.lambda-url.us-west-2.on.aws/"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateFunctionUrlConfigOutput(&lambda.CreateFunctionUrlConfigOutput{
			FunctionUrl: aws.String(functionUrl),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"qualifier":   core.MappingNodeFromString("PROD"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create function URL with qualifier",
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
			ResourceID: "test-function-url-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-url-id",
					ResourceName: "TestFunctionUrl",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionArn",
					},
					{
						FieldPath: "spec.qualifier",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionUrl": core.MappingNodeFromString(functionUrl),
				"spec.functionArn": core.MappingNodeFromString(""),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateFunctionUrlConfig": &lambda.CreateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				AuthType:     "",
				Qualifier:    aws.String("PROD"),
			},
		},
	}
}

func createComplexFunctionUrlTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	functionUrl := "https://test-function-url.lambda-url.us-west-2.on.aws/"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateFunctionUrlConfigOutput(&lambda.CreateFunctionUrlConfigOutput{
			FunctionUrl: aws.String(functionUrl),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"authType":    core.MappingNodeFromString("AWS_IAM"),
			"invokeMode":  core.MappingNodeFromString("RESPONSE_STREAM"),
			"qualifier":   core.MappingNodeFromString("STAGING"),
			"cors": {
				Fields: map[string]*core.MappingNode{
					"allowCredentials": core.MappingNodeFromBool(false),
					"allowHeaders": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("Content-Type"),
						},
					},
					"allowMethods": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("GET"),
						},
					},
					"allowOrigins": {
						Items: []*core.MappingNode{
							core.MappingNodeFromString("*"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create complex function URL with all features",
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
			ResourceID: "test-function-url-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-url-id",
					ResourceName: "TestFunctionUrl",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionArn",
					},
					{
						FieldPath: "spec.authType",
					},
					{
						FieldPath: "spec.invokeMode",
					},
					{
						FieldPath: "spec.qualifier",
					},
					{
						FieldPath: "spec.cors",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionUrl": core.MappingNodeFromString(functionUrl),
				"spec.functionArn": core.MappingNodeFromString(""),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateFunctionUrlConfig": &lambda.CreateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				AuthType:     "AWS_IAM",
				InvokeMode:   types.InvokeModeResponseStream,
				Qualifier:    aws.String("STAGING"),
				Cors: &types.Cors{
					AllowCredentials: aws.Bool(false),
					AllowHeaders:     []string{"Content-Type"},
					AllowMethods:     []string{"GET"},
					AllowOrigins:     []string{"*"},
				},
			},
		},
	}
}

func createFunctionUrlFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateFunctionUrlConfigError(fmt.Errorf("function not found")),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:non-existent-function"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create function URL failure",
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
			ResourceID: "test-function-url-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-url-id",
					ResourceName: "TestFunctionUrl",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionArn",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"CreateFunctionUrlConfig": &lambda.CreateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:non-existent-function"),
				AuthType:     "",
			},
		},
	}
}

func TestLambdaFunctionUrlResourceCreate(t *testing.T) {
	suite.Run(t, new(LambdaFunctionUrlResourceCreateSuite))
}
