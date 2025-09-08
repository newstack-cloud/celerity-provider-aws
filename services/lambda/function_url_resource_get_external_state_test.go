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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaFunctionUrlResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *LambdaFunctionUrlResourceGetExternalStateSuite) Test_get_external_state() {
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

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		getBasicFunctionUrlExternalStateTestCase(providerCtx, loader),
		getFunctionUrlWithCorsExternalStateTestCase(providerCtx, loader),
		getFunctionUrlWithAuthTypeExternalStateTestCase(providerCtx, loader),
		getFunctionUrlWithInvokeModeExternalStateTestCase(providerCtx, loader),
		getFunctionUrlWithQualifierExternalStateTestCase(providerCtx, loader),
		getComplexFunctionUrlExternalStateTestCase(providerCtx, loader),
		getFunctionUrlExternalStateFailureTestCase(providerCtx, loader),
		getFunctionUrlExternalStateNoIDTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		FunctionUrlResource,
		&s.Suite,
	)
}

func getBasicFunctionUrlExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	functionUrl := "https://test-function-url.lambda-url.us-west-2.on.aws/"
	functionARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{
			FunctionUrl: aws.String(functionUrl),
			FunctionArn: aws.String(functionARN),
			AuthType:    types.FunctionUrlAuthTypeNone,
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString(functionARN),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets basic function URL state",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext:     providerCtx,
			CurrentResourceSpec: specData,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionArn": core.MappingNodeFromString(functionARN),
					"functionUrl": core.MappingNodeFromString(functionUrl),
					"authType":    core.MappingNodeFromString("NONE"),
				},
			},
		},
	}
}

func getFunctionUrlWithCorsExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	functionUrl := "https://test-function-url.lambda-url.us-west-2.on.aws/"
	functionARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{
			FunctionUrl: aws.String(functionUrl),
			FunctionArn: aws.String(functionARN),
			AuthType:    types.FunctionUrlAuthTypeNone,
			Cors: &types.Cors{
				AllowCredentials: aws.Bool(true),
				AllowHeaders:     []string{"Content-Type", "Authorization"},
				AllowMethods:     []string{"GET", "POST"},
				AllowOrigins:     []string{"https://example.com"},
				ExposeHeaders:    []string{"X-Custom-Header"},
				MaxAge:           aws.Int32(3600),
			},
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString(functionARN),
			"functionUrl": core.MappingNodeFromString(functionUrl),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets function URL state with CORS",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext:     providerCtx,
			CurrentResourceSpec: specData,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionArn": core.MappingNodeFromString(functionARN),
					"functionUrl": core.MappingNodeFromString(functionUrl),
					"authType":    core.MappingNodeFromString("NONE"),
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
			},
		},
	}
}

func getFunctionUrlWithAuthTypeExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	functionUrl := "https://test-function-url.lambda-url.us-west-2.on.aws/"
	functionARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{
			FunctionUrl: aws.String(functionUrl),
			FunctionArn: aws.String(functionARN),
			AuthType:    types.FunctionUrlAuthTypeAwsIam,
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString(functionARN),
			"functionUrl": core.MappingNodeFromString(functionUrl),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets function URL state with auth type",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext:     providerCtx,
			CurrentResourceSpec: specData,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionArn": core.MappingNodeFromString(functionARN),
					"functionUrl": core.MappingNodeFromString(functionUrl),
					"authType":    core.MappingNodeFromString("AWS_IAM"),
				},
			},
		},
	}
}

func getFunctionUrlWithInvokeModeExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	functionUrl := "https://test-function-url.lambda-url.us-west-2.on.aws/"
	functionARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{
			FunctionUrl: aws.String(functionUrl),
			FunctionArn: aws.String(functionARN),
			AuthType:    types.FunctionUrlAuthTypeNone,
			InvokeMode:  types.InvokeModeBuffered,
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString(functionARN),
			"functionUrl": core.MappingNodeFromString(functionUrl),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets function URL state with invoke mode",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext:     providerCtx,
			CurrentResourceSpec: specData,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionArn": core.MappingNodeFromString(functionARN),
					"functionUrl": core.MappingNodeFromString(functionUrl),
					"authType":    core.MappingNodeFromString("NONE"),
					"invokeMode":  core.MappingNodeFromString("BUFFERED"),
				},
			},
		},
	}
}

func getFunctionUrlWithQualifierExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	functionUrl := "https://test-function-url.lambda-url.us-west-2.on.aws/"
	functionARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{
			FunctionUrl: aws.String(functionUrl),
			FunctionArn: aws.String(functionARN),
			AuthType:    types.FunctionUrlAuthTypeNone,
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString(functionARN),
			"functionUrl": core.MappingNodeFromString(functionUrl),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets function URL state with qualifier",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext:     providerCtx,
			CurrentResourceSpec: specData,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionArn": core.MappingNodeFromString(functionARN),
					"functionUrl": core.MappingNodeFromString(functionUrl),
					"authType":    core.MappingNodeFromString("NONE"),
				},
			},
		},
	}
}

func getComplexFunctionUrlExternalStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	functionUrl := "https://test-function-url.lambda-url.us-west-2.on.aws/"
	functionARN := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionUrlConfigOutput(&lambda.GetFunctionUrlConfigOutput{
			FunctionUrl: aws.String(functionUrl),
			FunctionArn: aws.String(functionARN),
			AuthType:    types.FunctionUrlAuthTypeAwsIam,
			InvokeMode:  types.InvokeModeResponseStream,
			Cors: &types.Cors{
				AllowCredentials: aws.Bool(false),
				AllowHeaders:     []string{"Content-Type"},
				AllowMethods:     []string{"GET"},
				AllowOrigins:     []string{"*"},
			},
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString(functionARN),
			"functionUrl": core.MappingNodeFromString(functionUrl),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets complex function URL state with all features",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext:     providerCtx,
			CurrentResourceSpec: specData,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionArn": core.MappingNodeFromString(functionARN),
					"functionUrl": core.MappingNodeFromString(functionUrl),
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
			},
		},
	}
}

func getFunctionUrlExternalStateFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionUrlConfigError(fmt.Errorf("function URL not found")),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function"),
			"functionUrl": core.MappingNodeFromString("https://test-function-url.lambda-url.us-west-2.on.aws/"),
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "handles get function URL error",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext:     providerCtx,
			CurrentResourceSpec: specData,
		},
		ExpectError: true,
	}
}

func getFunctionUrlExternalStateNoIDTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock()

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "returns error when no function arn is present",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext:     providerCtx,
			CurrentResourceSpec: specData,
		},
		ExpectError: true,
	}
}

func TestLambdaFunctionUrlResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionUrlResourceGetExternalStateSuite))
}
