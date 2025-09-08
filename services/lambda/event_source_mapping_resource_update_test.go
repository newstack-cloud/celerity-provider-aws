//go:build unit

package lambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
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

type LambdaEventSourceMappingResourceUpdateSuite struct {
	suite.Suite
}

func (s *LambdaEventSourceMappingResourceUpdateSuite) Test_update_lambda_event_source_mapping() {
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
		updateEventSourceMappingBasicUpdateTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		EventSourceMappingResource,
		&s.Suite,
	)
}

func updateEventSourceMappingBasicUpdateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service] {
	uuid := "test-update-uuid"
	eventSourceMappingArn := "arn:aws:lambda:us-west-2:123456789012:event-source-mapping:test-update-uuid"
	functionArn := "arn:aws:lambda:us-west-2:123456789012:function:test-function"

	service := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithUpdateEventSourceMappingOutput(&lambda.UpdateEventSourceMappingOutput{
			UUID:                  aws.String(uuid),
			EventSourceMappingArn: aws.String(eventSourceMappingArn),
			FunctionArn:           aws.String(functionArn),
			State:                 aws.String("Updating"),
			EventSourceArn:        aws.String("arn:aws:sqs:us-west-2:123456789012:test-queue"),
			BatchSize:             aws.Int32(20), // Updated from 10 to 20
		}),
		lambdamock.WithTagResourceOutput(&lambda.TagResourceOutput{}),
	)

	// New spec data with updated values
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"id":             core.MappingNodeFromString(uuid),
			"functionName":   core.MappingNodeFromString("test-function"),
			"eventSourceArn": core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:test-queue"),
			"batchSize":      core.MappingNodeFromInt(20), // Updated batch size
			"tags": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("production"),
						},
					},
				},
			},
		},
	}

	// Current resource state - this is crucial for triggering Update instead of Create
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"id":                    core.MappingNodeFromString(uuid),
			"eventSourceMappingArn": core.MappingNodeFromString(eventSourceMappingArn),
			"functionArn":           core.MappingNodeFromString(functionArn),
			"functionName":          core.MappingNodeFromString("test-function"),
			"eventSourceArn":        core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:test-queue"),
			"batchSize":             core.MappingNodeFromInt(10), // Original batch size
			"state":                 core.MappingNodeFromString("Enabled"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, lambdaservice.Service]{
		Name: "update event source mapping batch size and add tags",
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
			ResourceID: "test-esm-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-esm-id",
					ResourceName: "TestEventSourceMapping",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-esm-id",
						Name:       "TestEventSourceMapping",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/lambda/eventSourceMapping",
						},
						Spec: specData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.batchSize",
					},
					{
						FieldPath: "spec.tags",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.id":                    core.MappingNodeFromString(uuid),
				"spec.eventSourceMappingArn": core.MappingNodeFromString(eventSourceMappingArn),
				"spec.functionArn":           core.MappingNodeFromString(functionArn),
				"spec.state":                 core.MappingNodeFromString("Updating"),
			},
		},
		SaveActionsCalled: map[string]any{
			"UpdateEventSourceMapping": &lambda.UpdateEventSourceMappingInput{
				UUID:         aws.String(uuid),
				FunctionName: aws.String("test-function"),
				BatchSize:    aws.Int32(20),
			},
			"TagResource": &lambda.TagResourceInput{
				Resource: aws.String(eventSourceMappingArn),
				Tags: map[string]string{
					"Environment": "production",
				},
			},
		},
	}
}

func TestLambdaEventSourceMappingResourceUpdate(t *testing.T) {
	suite.Run(t, new(LambdaEventSourceMappingResourceUpdateSuite))
}
