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

type LambdaFunctionUrlResourceUpdateSuite struct {
	suite.Suite
}

func (s *LambdaFunctionUrlResourceUpdateSuite) Test_update_lambda_function_url() {
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
		updateFunctionUrlCorsTestCase(providerCtx, loader),
		updateFunctionUrlAuthTypeTestCase(providerCtx, loader),
		updateFunctionUrlInvokeModeTestCase(providerCtx, loader),
		updateFunctionUrlQualifierTestCase(providerCtx, loader),
		updateFunctionUrlComplexTestCase(providerCtx, loader),
		updateFunctionUrlNoChangesTestCase(providerCtx, loader),
		updateFunctionUrlFailureTestCase(providerCtx, loader),
		recreateFunctionUrlOnTargetFunctionArnOrAuthTypeOrQualifierChangeTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		FunctionUrlResource,
		&s.Suite,
	)
}

func updateFunctionUrlCorsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionUrlConfigOutput(&lambda.UpdateFunctionUrlConfigOutput{
			FunctionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			FunctionUrl: aws.String("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
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
				},
			},
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update function URL CORS",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-url-id",
						Name:       "TestFunctionUrl",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.cors",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"UpdateFunctionUrlConfig": &lambda.UpdateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				Cors: &types.Cors{
					AllowCredentials: aws.Bool(true),
					AllowHeaders:     []string{"Content-Type", "Authorization"},
					AllowMethods:     []string{"GET", "POST"},
					AllowOrigins:     []string{"https://example.com"},
				},
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				"spec.functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			},
		},
	}
}

func updateFunctionUrlAuthTypeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionUrlConfigOutput(&lambda.UpdateFunctionUrlConfigOutput{
			FunctionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			FunctionUrl: aws.String("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			"authType":    core.MappingNodeFromString("AWS_IAM"),
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			"authType":    core.MappingNodeFromString("NONE"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update function URL auth type",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-url-id",
						Name:       "TestFunctionUrl",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.authType",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"UpdateFunctionUrlConfig": &lambda.UpdateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				AuthType:     types.FunctionUrlAuthTypeAwsIam,
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				"spec.functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			},
		},
	}
}

func updateFunctionUrlInvokeModeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionUrlConfigOutput(&lambda.UpdateFunctionUrlConfigOutput{
			FunctionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			FunctionUrl: aws.String("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			"invokeMode":  core.MappingNodeFromString("RESPONSE_STREAM"),
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			"invokeMode":  core.MappingNodeFromString("BUFFERED"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update function URL invoke mode",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-url-id",
						Name:       "TestFunctionUrl",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.invokeMode",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"UpdateFunctionUrlConfig": &lambda.UpdateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				InvokeMode:   types.InvokeModeResponseStream,
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
				"spec.functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			},
		},
	}
}

func updateFunctionUrlQualifierTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionUrlConfigOutput(&lambda.UpdateFunctionUrlConfigOutput{
			FunctionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			FunctionUrl: aws.String("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			"qualifier":   core.MappingNodeFromString("STAGING"),
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			"qualifier":   core.MappingNodeFromString("PROD"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update function URL qualifier",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-url-id",
						Name:       "TestFunctionUrl",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.qualifier",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"UpdateFunctionUrlConfig": &lambda.UpdateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				Qualifier:    aws.String("STAGING"),
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
				"spec.functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			},
		},
	}
}

func updateFunctionUrlComplexTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionUrlConfigOutput(&lambda.UpdateFunctionUrlConfigOutput{
			FunctionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			FunctionUrl: aws.String("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			"authType":    core.MappingNodeFromString("AWS_IAM"),
			"invokeMode":  core.MappingNodeFromString("RESPONSE_STREAM"),
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

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			"authType":    core.MappingNodeFromString("NONE"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update function URL with complex changes",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-url-id",
						Name:       "TestFunctionUrl",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.authType",
					},
					{
						FieldPath: "spec.invokeMode",
					},
					{
						FieldPath: "spec.cors",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"UpdateFunctionUrlConfig": &lambda.UpdateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				AuthType:     types.FunctionUrlAuthTypeAwsIam,
				InvokeMode:   types.InvokeModeResponseStream,
				Cors: &types.Cors{
					AllowCredentials: aws.Bool(false),
					AllowHeaders:     []string{"Content-Type"},
					AllowMethods:     []string{"GET"},
					AllowOrigins:     []string{"*"},
				},
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
				"spec.functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			},
		},
	}
}

func updateFunctionUrlNoChangesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionUrlConfigOutput(&lambda.UpdateFunctionUrlConfigOutput{}),
	)

	functionARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"
	functionURL := "https://test-function-url.lambda-url.us-west-2.on.aws/"

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString(functionARN),
			"functionUrl": core.MappingNodeFromString(functionURL),
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
			ResourceID: "test-function-url-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-function-url-id",
					ResourceName: "TestFunctionUrl",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-url-id",
						Name:       "TestFunctionUrl",
						InstanceID: "test-instance-id",
						SpecData:   specData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionArn": core.MappingNodeFromString(functionARN),
				"spec.functionUrl": core.MappingNodeFromString(functionURL),
			},
		},
	}
}

func updateFunctionUrlFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionUrlConfigError(fmt.Errorf("function URL not found")),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			"authType":    core.MappingNodeFromString("AWS_IAM"),
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update function URL failure",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-url-id",
						Name:       "TestFunctionUrl",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.authType",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"UpdateFunctionUrlConfig": &lambda.UpdateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				AuthType:     types.FunctionUrlAuthTypeAwsIam,
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
				"spec.functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
			},
		},
	}
}

func recreateFunctionUrlOnTargetFunctionArnOrAuthTypeOrQualifierChangeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	oldFunctionUrl := "https://old-function-url.lambda-url.us-west-2.on.aws/"
	newFunctionUrl := "https://new-function-url.lambda-url.us-west-2.on.aws/"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteFunctionUrlConfigOutput(&lambda.DeleteFunctionUrlConfigOutput{}),
		lambdamock.WithCreateFunctionUrlConfigOutput(&lambda.CreateFunctionUrlConfigOutput{
			FunctionArn: aws.String("arn:aws:lambda:us-west-2:123456789012:function:new-function"),
			FunctionUrl: aws.String(newFunctionUrl),
		}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn":       core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:old-function"),
			"functionUrl":       core.MappingNodeFromString(oldFunctionUrl),
			"targetFunctionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:old-function"),
			"authType":          core.MappingNodeFromString("NONE"),
			"qualifier":         core.MappingNodeFromString(""),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn":       core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:new-function"),
			"targetFunctionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:new-function"),
			"authType":          core.MappingNodeFromString("AWS_IAM"),
			"qualifier":         core.MappingNodeFromString("PROD"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "recreate function URL on targetFunctionArn, authType, or qualifier change",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-function-url-id",
						Name:       "TestFunctionUrl",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/functionUrl",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.targetFunctionArn",
						PrevValue: core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:old-function"),
						NewValue:  core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:new-function"),
					},
					{
						FieldPath: "spec.authType",
						PrevValue: core.MappingNodeFromString("NONE"),
						NewValue:  core.MappingNodeFromString("AWS_IAM"),
					},
					{
						FieldPath: "spec.qualifier",
						PrevValue: core.MappingNodeFromString(""),
						NewValue:  core.MappingNodeFromString("PROD"),
					},
				},
				MustRecreate: true,
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:new-function"),
				"spec.functionUrl": core.MappingNodeFromString(newFunctionUrl),
			},
		},
		SaveActionsCalled: map[string]any{
			"DeleteFunctionUrlConfig": &lambda.DeleteFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:old-function"),
				Qualifier:    aws.String(""),
			},
			"CreateFunctionUrlConfig": &lambda.CreateFunctionUrlConfigInput{
				FunctionName: aws.String("arn:aws:lambda:us-west-2:123456789012:function:new-function"),
				AuthType:     types.FunctionUrlAuthTypeAwsIam,
				Qualifier:    aws.String("PROD"),
			},
		},
	}
}

func TestLambdaFunctionUrlResourceUpdate(t *testing.T) {
	suite.Run(t, new(LambdaFunctionUrlResourceUpdateSuite))
}
