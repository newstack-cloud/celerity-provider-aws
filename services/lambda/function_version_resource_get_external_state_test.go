//go:build unit

package lambda

import (
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

type LambdaFunctionVersionResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *LambdaFunctionVersionResourceGetExternalStateSuite) Test_get_external_state() {
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

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		createBasicFunctionVersionStateTestCase(providerCtx, loader),
		createAllOptionalConfigsFunctionVersionTestCase(providerCtx, loader),
		createGetFunctionVersionErrorTestCase(providerCtx, loader),
		createGetProvisionedConcurrencyErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		FunctionVersionResource,
		&s.Suite,
	)
}

func TestLambdaFunctionVersionResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionVersionResourceGetExternalStateSuite))
}

// Test case generator functions below.

func createBasicFunctionVersionStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets basic function version state",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					FunctionArn: aws.String(
						"arn:aws:lambda:us-east-1:123456789012:function:test-function",
					),
					Version: aws.String("1"),
				},
			}),
			lambdamock.WithGetProvisionedConcurrencyOutput(&lambda.GetProvisionedConcurrencyConfigOutput{
				RequestedProvisionedConcurrentExecutions: aws.Int32(10),
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName": core.MappingNodeFromString("test-function"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName": core.MappingNodeFromString("test-function"),
					"functionArn": core.MappingNodeFromString(
						"arn:aws:lambda:us-east-1:123456789012:function:test-function",
					),
					"functionArnWithVersion": core.MappingNodeFromString(
						"arn:aws:lambda:us-east-1:123456789012:function:test-function:1",
					),
					"version": core.MappingNodeFromString("1"),
					"provisionedConcurrencyConfig": {
						Fields: map[string]*core.MappingNode{
							"provisionedConcurrentExecutions": core.MappingNodeFromInt(10),
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createAllOptionalConfigsFunctionVersionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "successfully gets function version state with all optional configurations",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					Version:      aws.String("1"),
					Description:  aws.String("Test function version"),
					RuntimeVersionConfig: &types.RuntimeVersionConfig{
						RuntimeVersionArn: aws.String("arn:aws:lambda:us-east-1::runtime-version/test"),
					},
				},
			}),
			lambdamock.WithGetProvisionedConcurrencyOutput(&lambda.GetProvisionedConcurrencyConfigOutput{
				RequestedProvisionedConcurrentExecutions: aws.Int32(10),
			}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName": core.MappingNodeFromString("test-function"),
					"description":  core.MappingNodeFromString("Test function version"),
					"runtimePolicy": {
						Fields: map[string]*core.MappingNode{
							"updateRuntimeOn": core.MappingNodeFromString("Auto"),
						},
					},
					"provisionedConcurrencyConfig": {
						Fields: map[string]*core.MappingNode{
							"provisionedConcurrentExecutions": core.MappingNodeFromInt(10),
						},
					},
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName": core.MappingNodeFromString("test-function"),
					"description":  core.MappingNodeFromString("Test function version"),
					"runtimePolicy": {
						Fields: map[string]*core.MappingNode{
							"updateRuntimeOn": core.MappingNodeFromString("Auto"),
							"runtimeVersionArn": core.MappingNodeFromString(
								"arn:aws:lambda:us-east-1::runtime-version/test",
							),
						},
					},
					"provisionedConcurrencyConfig": {
						Fields: map[string]*core.MappingNode{
							"provisionedConcurrentExecutions": core.MappingNodeFromInt(10),
						},
					},
					"functionArn": core.MappingNodeFromString(
						"arn:aws:lambda:us-east-1:123456789012:function:test-function",
					),
					"version": core.MappingNodeFromString("1"),
					"functionArnWithVersion": core.MappingNodeFromString(
						"arn:aws:lambda:us-east-1:123456789012:function:test-function:1",
					),
				},
			},
		},
		ExpectError: false,
	}
}

func createGetFunctionVersionErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "handles get function version error",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionError(errors.New("failed to get function")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName": core.MappingNodeFromString("test-function"),
				},
			},
		},
		ExpectError: true,
	}
}

func createGetProvisionedConcurrencyErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "handles get provisioned concurrency config error",
		ServiceFactory: lambdamock.CreateLambdaServiceMockFactory(
			lambdamock.WithGetFunctionOutput(&lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					Version:      aws.String("1"),
				},
			}),
			lambdamock.WithGetProvisionedConcurrencyError(errors.New("failed to get provisioned concurrency config")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName": core.MappingNodeFromString("test-function"),
					"provisionedConcurrencyConfig": {
						Fields: map[string]*core.MappingNode{
							"provisionedConcurrentExecutions": core.MappingNodeFromInt(10),
						},
					},
				},
			},
		},
		ExpectError: true,
	}
}
