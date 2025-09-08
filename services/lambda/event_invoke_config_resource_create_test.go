//go:build unit

package lambda

import (
	"fmt"
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
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaEventInvokeConfigResourceCreateSuite struct {
	suite.Suite
}

func (s *LambdaEventInvokeConfigResourceCreateSuite) Test_create_lambda_event_invoke_config() {
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

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		createBasicEventInvokeConfigTestCase(providerCtx, loader),
		createEventInvokeConfigWithDestinationsTestCase(providerCtx, loader),
		createEventInvokeConfigFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		EventInvokeConfigResource,
		&s.Suite,
	)
}

func createBasicEventInvokeConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	functionArn := "arn:aws:lambda:us-east-1:123456789012:function:test-function"
	lastModified := time.Now()

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithPutFunctionEventInvokeConfigOutput(&lambda.PutFunctionEventInvokeConfigOutput{
			FunctionArn:              aws.String(functionArn),
			MaximumRetryAttempts:     aws.Int32(1),
			MaximumEventAgeInSeconds: aws.Int32(300),
			LastModified:             &lastModified,
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":             core.MappingNodeFromString("test-function"),
			"qualifier":                core.MappingNodeFromString("$LATEST"),
			"maximumRetryAttempts":     core.MappingNodeFromInt(1),
			"maximumEventAgeInSeconds": core.MappingNodeFromInt(300),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create basic event invoke config",
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
			ResourceID: "test-event-invoke-config-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-event-invoke-config-id",
					ResourceName: "TestEventInvokeConfig",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventInvokeConfig",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.qualifier",
					},
					{
						FieldPath: "spec.maximumRetryAttempts",
					},
					{
						FieldPath: "spec.maximumEventAgeInSeconds",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionArn":  core.MappingNodeFromString(functionArn),
				"spec.lastModified": core.MappingNodeFromString(lastModified.String()),
			},
		},
		SaveActionsCalled: map[string]any{
			"PutFunctionEventInvokeConfig": &lambda.PutFunctionEventInvokeConfigInput{
				FunctionName:             aws.String("test-function"),
				Qualifier:                aws.String("$LATEST"),
				MaximumRetryAttempts:     aws.Int32(1),
				MaximumEventAgeInSeconds: aws.Int32(300),
			},
		},
	}
}

func createEventInvokeConfigWithDestinationsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	functionArn := "arn:aws:lambda:us-east-1:123456789012:function:test-function"
	lastModified := time.Now()

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithPutFunctionEventInvokeConfigOutput(&lambda.PutFunctionEventInvokeConfigOutput{
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

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":             core.MappingNodeFromString("test-function"),
			"qualifier":                core.MappingNodeFromString("$LATEST"),
			"maximumRetryAttempts":     core.MappingNodeFromInt(2),
			"maximumEventAgeInSeconds": core.MappingNodeFromInt(1800),
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
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create event invoke config with destinations",
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
			ResourceID: "test-event-invoke-config-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-event-invoke-config-id",
					ResourceName: "TestEventInvokeConfig",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventInvokeConfig",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.qualifier",
					},
					{
						FieldPath: "spec.maximumRetryAttempts",
					},
					{
						FieldPath: "spec.maximumEventAgeInSeconds",
					},
					{
						FieldPath: "spec.destinationConfig",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionArn":  core.MappingNodeFromString(functionArn),
				"spec.lastModified": core.MappingNodeFromString(lastModified.String()),
			},
		},
		SaveActionsCalled: map[string]any{
			"PutFunctionEventInvokeConfig": &lambda.PutFunctionEventInvokeConfigInput{
				FunctionName:             aws.String("test-function"),
				Qualifier:                aws.String("$LATEST"),
				MaximumRetryAttempts:     aws.Int32(2),
				MaximumEventAgeInSeconds: aws.Int32(1800),
				DestinationConfig: &types.DestinationConfig{
					OnSuccess: &types.OnSuccess{
						Destination: aws.String("arn:aws:sqs:us-east-1:123456789012:success-queue"),
					},
					OnFailure: &types.OnFailure{
						Destination: aws.String("arn:aws:sqs:us-east-1:123456789012:failure-queue"),
					},
				},
			},
		},
	}
}

func createEventInvokeConfigFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithPutFunctionEventInvokeConfigError(fmt.Errorf("failed to create event invoke config")),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"qualifier":    core.MappingNodeFromString("$LATEST"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "create event invoke config failure",
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
			ResourceID: "test-event-invoke-config-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-event-invoke-config-id",
					ResourceName: "TestEventInvokeConfig",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventInvokeConfig",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.functionName",
					},
					{
						FieldPath: "spec.qualifier",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"PutFunctionEventInvokeConfig": &lambda.PutFunctionEventInvokeConfigInput{
				FunctionName: aws.String("test-function"),
				Qualifier:    aws.String("$LATEST"),
			},
		},
	}
}

func TestLambdaEventInvokeConfigResourceCreate(t *testing.T) {
	suite.Run(t, new(LambdaEventInvokeConfigResourceCreateSuite))
}
