//go:build unit

package lambda

import (
	"testing"
	"time"

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

type LambdaEventInvokeConfigResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *LambdaEventInvokeConfigResourceGetExternalStateSuite) Test_get_external_state_lambda_event_invoke_config() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-east-1"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		getExternalStateEventInvokeConfigBasicTestCase(providerCtx, loader),
		getExternalStateEventInvokeConfigCompleteTestCase(providerCtx, loader),
		getExternalStateEventInvokeConfigNotFoundTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		EventInvokeConfigResource,
		&s.Suite,
	)
}

func getExternalStateEventInvokeConfigBasicTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	functionArn := "arn:aws:lambda:us-east-1:123456789012:function:test-function"
	lastModified := time.Now()

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionEventInvokeConfigOutput(&lambda.GetFunctionEventInvokeConfigOutput{
			FunctionArn:          aws.String(functionArn),
			MaximumRetryAttempts: aws.Int32(1),
			LastModified:         &lastModified,
		}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "get external state event invoke config basic",
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
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName": core.MappingNodeFromString("test-function"),
					"qualifier":    core.MappingNodeFromString("$LATEST"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionArn":          core.MappingNodeFromString(functionArn),
					"maximumRetryAttempts": core.MappingNodeFromInt(1),
					"lastModified":         core.MappingNodeFromString(lastModified.String()),
				},
			},
		},
	}
}

func getExternalStateEventInvokeConfigCompleteTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	functionArn := "arn:aws:lambda:us-east-1:123456789012:function:test-function"
	lastModified := time.Now()

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionEventInvokeConfigOutput(&lambda.GetFunctionEventInvokeConfigOutput{
			FunctionArn:              aws.String(functionArn),
			MaximumRetryAttempts:     aws.Int32(2),
			MaximumEventAgeInSeconds: aws.Int32(1800),
			LastModified:             &lastModified,
			DestinationConfig: &types.DestinationConfig{
				OnSuccess: &types.OnSuccess{
					Destination: aws.String("arn:aws:sqs:us-east-1:123456789012:success-queue"),
				},
				OnFailure: &types.OnFailure{
					Destination: aws.String("arn:aws:sqs:us-east-1:123456789012:failure-queue"),
				},
			},
		}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "get external state event invoke config complete",
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
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName": core.MappingNodeFromString("test-function"),
					"qualifier":    core.MappingNodeFromString("$LATEST"),
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionArn":              core.MappingNodeFromString(functionArn),
					"maximumRetryAttempts":     core.MappingNodeFromInt(2),
					"maximumEventAgeInSeconds": core.MappingNodeFromInt(1800),
					"lastModified":             core.MappingNodeFromString(lastModified.String()),
					"destinationConfig": {
						Fields: map[string]*core.MappingNode{
							"onSuccess": {
								Fields: map[string]*core.MappingNode{
									"destination": core.MappingNodeFromString("arn:aws:sqs:us-east-1:123456789012:success-queue"),
								},
							},
							"onFailure": {
								Fields: map[string]*core.MappingNode{
									"destination": core.MappingNodeFromString("arn:aws:sqs:us-east-1:123456789012:failure-queue"),
								},
							},
						},
					},
				},
			},
		},
	}
}

func getExternalStateEventInvokeConfigNotFoundTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionEventInvokeConfigError(&types.ResourceNotFoundException{
			Message: aws.String("Function event invoke config not found"),
		}),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, lambdaservice.Service]{
		Name: "get external state event invoke config not found",
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
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"functionName": core.MappingNodeFromString("test-function"),
					"qualifier":    core.MappingNodeFromString("$LATEST"),
				},
			},
		},
		ExpectedOutput: nil,
		ExpectError:    true,
	}
}

func TestLambdaEventInvokeConfigResourceGetExternalState(t *testing.T) {
	suite.Run(t, new(LambdaEventInvokeConfigResourceGetExternalStateSuite))
}
