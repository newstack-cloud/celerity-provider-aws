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
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaEventInvokeConfigResourceUpdateSuite struct {
	suite.Suite
}

func (s *LambdaEventInvokeConfigResourceUpdateSuite) Test_update_lambda_event_invoke_config() {
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
		updateEventInvokeConfigRetryAttemptsTestCase(providerCtx, loader),
		updateEventInvokeConfigEventAgeTestCase(providerCtx, loader),
		updateEventInvokeConfigDestinationsTestCase(providerCtx, loader),
		updateEventInvokeConfigCompleteTestCase(providerCtx, loader),
		updateEventInvokeConfigFailureTestCase(providerCtx, loader),
		recreateEventInvokeConfigOnFunctionNameOrQualifierChangeTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		EventInvokeConfigResource,
		&s.Suite,
	)
}

func updateEventInvokeConfigRetryAttemptsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	functionArn := "arn:aws:lambda:us-east-1:123456789012:function:test-function"
	lastModified := time.Now()

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionEventInvokeConfigOutput(&lambda.UpdateFunctionEventInvokeConfigOutput{
			FunctionArn:          aws.String(functionArn),
			MaximumRetryAttempts: aws.Int32(2),
			LastModified:         &lastModified,
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":         core.MappingNodeFromString("test-function"),
			"qualifier":            core.MappingNodeFromString("$LATEST"),
			"maximumRetryAttempts": core.MappingNodeFromInt(2),
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":         core.MappingNodeFromString("test-function"),
			"qualifier":            core.MappingNodeFromString("$LATEST"),
			"maximumRetryAttempts": core.MappingNodeFromInt(1),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update event invoke config retry attempts",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-event-invoke-config-id",
						Name:       "TestEventInvokeConfig",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventInvokeConfig",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.maximumRetryAttempts",
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
			"UpdateFunctionEventInvokeConfig": &lambda.UpdateFunctionEventInvokeConfigInput{
				FunctionName:         aws.String("test-function"),
				Qualifier:            aws.String("$LATEST"),
				MaximumRetryAttempts: aws.Int32(2),
			},
		},
	}
}

func updateEventInvokeConfigEventAgeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	functionArn := "arn:aws:lambda:us-east-1:123456789012:function:test-function"
	lastModified := time.Now()

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionEventInvokeConfigOutput(&lambda.UpdateFunctionEventInvokeConfigOutput{
			FunctionArn:              aws.String(functionArn),
			MaximumEventAgeInSeconds: aws.Int32(1800),
			LastModified:             &lastModified,
		}),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":             core.MappingNodeFromString("test-function"),
			"qualifier":                core.MappingNodeFromString("$LATEST"),
			"maximumEventAgeInSeconds": core.MappingNodeFromInt(1800),
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":             core.MappingNodeFromString("test-function"),
			"qualifier":                core.MappingNodeFromString("$LATEST"),
			"maximumEventAgeInSeconds": core.MappingNodeFromInt(300),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update event invoke config event age",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-event-invoke-config-id",
						Name:       "TestEventInvokeConfig",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventInvokeConfig",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
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
			"UpdateFunctionEventInvokeConfig": &lambda.UpdateFunctionEventInvokeConfigInput{
				FunctionName:             aws.String("test-function"),
				Qualifier:                aws.String("$LATEST"),
				MaximumEventAgeInSeconds: aws.Int32(1800),
			},
		},
	}
}

func updateEventInvokeConfigDestinationsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	functionArn := "arn:aws:lambda:us-east-1:123456789012:function:test-function"
	lastModified := time.Now()

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionEventInvokeConfigOutput(&lambda.UpdateFunctionEventInvokeConfigOutput{
			FunctionArn:  aws.String(functionArn),
			LastModified: &lastModified,
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
			"functionName": core.MappingNodeFromString("test-function"),
			"qualifier":    core.MappingNodeFromString("$LATEST"),
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

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName": core.MappingNodeFromString("test-function"),
			"qualifier":    core.MappingNodeFromString("$LATEST"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update event invoke config destinations",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-event-invoke-config-id",
						Name:       "TestEventInvokeConfig",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventInvokeConfig",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
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
			"UpdateFunctionEventInvokeConfig": &lambda.UpdateFunctionEventInvokeConfigInput{
				FunctionName: aws.String("test-function"),
				Qualifier:    aws.String("$LATEST"),
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

func updateEventInvokeConfigCompleteTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	functionArn := "arn:aws:lambda:us-east-1:123456789012:function:test-function"
	lastModified := time.Now()

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionEventInvokeConfigOutput(&lambda.UpdateFunctionEventInvokeConfigOutput{
			FunctionArn:              aws.String(functionArn),
			MaximumRetryAttempts:     aws.Int32(3),
			MaximumEventAgeInSeconds: aws.Int32(3600),
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
			"maximumRetryAttempts":     core.MappingNodeFromInt(3),
			"maximumEventAgeInSeconds": core.MappingNodeFromInt(3600),
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

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":             core.MappingNodeFromString("test-function"),
			"qualifier":                core.MappingNodeFromString("$LATEST"),
			"maximumRetryAttempts":     core.MappingNodeFromInt(1),
			"maximumEventAgeInSeconds": core.MappingNodeFromInt(300),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update event invoke config complete",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-event-invoke-config-id",
						Name:       "TestEventInvokeConfig",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventInvokeConfig",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
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
			"UpdateFunctionEventInvokeConfig": &lambda.UpdateFunctionEventInvokeConfigInput{
				FunctionName:             aws.String("test-function"),
				Qualifier:                aws.String("$LATEST"),
				MaximumRetryAttempts:     aws.Int32(3),
				MaximumEventAgeInSeconds: aws.Int32(3600),
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

func updateEventInvokeConfigFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateFunctionEventInvokeConfigError(fmt.Errorf("failed to update event invoke config")),
	)

	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":         core.MappingNodeFromString("test-function"),
			"qualifier":            core.MappingNodeFromString("$LATEST"),
			"maximumRetryAttempts": core.MappingNodeFromInt(2),
		},
	}

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":         core.MappingNodeFromString("test-function"),
			"qualifier":            core.MappingNodeFromString("$LATEST"),
			"maximumRetryAttempts": core.MappingNodeFromInt(1),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update event invoke config failure",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-event-invoke-config-id",
						Name:       "TestEventInvokeConfig",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventInvokeConfig",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.maximumRetryAttempts",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"UpdateFunctionEventInvokeConfig": &lambda.UpdateFunctionEventInvokeConfigInput{
				FunctionName:         aws.String("test-function"),
				Qualifier:            aws.String("$LATEST"),
				MaximumRetryAttempts: aws.Int32(2),
			},
		},
	}
}

func recreateEventInvokeConfigOnFunctionNameOrQualifierChangeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	oldFunctionArn := "arn:aws:lambda:us-east-1:123456789012:function:old-function"
	newFunctionArn := "arn:aws:lambda:us-east-1:123456789012:function:new-function"
	lastModified := time.Now()

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithDeleteFunctionEventInvokeConfigOutput(&lambda.DeleteFunctionEventInvokeConfigOutput{}),
		lambdamock.WithPutFunctionEventInvokeConfigOutput(&lambda.PutFunctionEventInvokeConfigOutput{
			FunctionArn:  aws.String(newFunctionArn),
			LastModified: &lastModified,
		}),
	)

	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn":          core.MappingNodeFromString(oldFunctionArn),
			"functionName":         core.MappingNodeFromString("old-function"),
			"qualifier":            core.MappingNodeFromString("$LATEST"),
			"maximumRetryAttempts": core.MappingNodeFromInt(1),
		},
	}

	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":         core.MappingNodeFromString("new-function"),
			"qualifier":            core.MappingNodeFromString("PROD"),
			"maximumRetryAttempts": core.MappingNodeFromInt(2),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "recreate event invoke config on functionName or qualifier change",
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
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-event-invoke-config-id",
						Name:       "TestEventInvokeConfig",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventInvokeConfig",
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
						FieldPath: "spec.qualifier",
						PrevValue: core.MappingNodeFromString("$LATEST"),
						NewValue:  core.MappingNodeFromString("PROD"),
					},
				},
				MustRecreate: true,
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionArn":  core.MappingNodeFromString(newFunctionArn),
				"spec.lastModified": core.MappingNodeFromString(lastModified.String()),
			},
		},
		SaveActionsCalled: map[string]any{
			"DeleteFunctionEventInvokeConfig": &lambda.DeleteFunctionEventInvokeConfigInput{
				FunctionName: aws.String("old-function"),
				Qualifier:    aws.String("$LATEST"),
			},
			"PutFunctionEventInvokeConfig": &lambda.PutFunctionEventInvokeConfigInput{
				FunctionName:         aws.String("new-function"),
				Qualifier:            aws.String("PROD"),
				MaximumRetryAttempts: aws.Int32(2),
			},
		},
	}
}

func TestLambdaEventInvokeConfigResourceUpdate(t *testing.T) {
	suite.Run(t, new(LambdaEventInvokeConfigResourceUpdateSuite))
}
