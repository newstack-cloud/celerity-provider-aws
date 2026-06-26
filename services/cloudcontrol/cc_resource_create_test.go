//go:build unit

package cloudcontrol

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscc "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	cctypes "github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	cloudcontrolmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/cloudcontrol_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type CCResourceCreateSuite struct {
	suite.Suite
}

const (
	testQueueURL     = "https://sqs.us-west-2.amazonaws.com/123456789012/test-queue"
	testQueueARN     = "arn:aws:sqs:us-west-2:123456789012:test-queue"
	testRequestToken = "create-request-token"
)

func (s *CCResourceCreateSuite) Test_create() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
		map[string]*core.ScalarValue{"session_id": core.ScalarFromString("test-session-id")},
	)

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		createBasicTestCase(providerCtx, loader),
		createEarlyCaptureTestCase(providerCtx, loader),
		createIncompleteModelTestCase(providerCtx, loader),
		createEmptyComputedFieldModelTestCase(providerCtx, loader),
		createFailureTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(testCases, newTestResource, &s.Suite)
}

// Covers a stabilise-required (slow) resource: Create returns the request token and
// identifier at config-complete without reading any computed fields.
func (s *CCResourceCreateSuite) Test_create_slow_type_defers_fields_to_stabilisation() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
		map[string]*core.ScalarValue{"session_id": core.ScalarFromString("test-session-id")},
	)

	const dbID = "orders-db"
	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithCreateResourceOutput(&awscc.CreateResourceOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				RequestToken: aws.String(testRequestToken),
				Identifier:   aws.String(dbID),
			},
		}),
		// Deliberately no GetResource / GetResourceRequestStatus stubs: the slow-type
		// path must not call them.
	)

	spec := core.MappingNodeFields("dbInstanceIdentifier", core.MappingNodeFromString(dbID))

	testCase := plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:             "slow type returns identifier without capturing computed fields",
		ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore:      newAWSConfigStore(loader),
		Input:            deployInput(providerCtx, spec, nil),
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				fmt.Sprintf("spec.%s", fieldRequestToken):      core.MappingNodeFromString(testRequestToken),
				fmt.Sprintf("spec.%s", fieldPrimaryIdentifier): core.MappingNodeFromString(dbID),
			},
		},
		SaveActionsNotCalled: []string{"GetResource", "GetResourceRequestStatus"},
	}

	newDatabaseResource := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return CCResource(testDatabaseConfig(), serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDeployTestCases(
		[]plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{testCase},
		newDatabaseResource,
		&s.Suite,
	)
}

func createBasicTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service] {
	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithCreateResourceOutput(&awscc.CreateResourceOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				RequestToken: aws.String(testRequestToken),
				Identifier:   aws.String(testQueueURL),
			},
		}),
		// The operation completes without a resource model on the success event, so
		// the deploy falls back to a GetResource read to capture computed fields.
		cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				OperationStatus: cctypes.OperationStatusSuccess,
				Identifier:      aws.String(testQueueURL),
			},
		}),
		cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
			ResourceDescription: &cctypes.ResourceDescription{
				Identifier: aws.String(testQueueURL),
				Properties: aws.String(fmt.Sprintf(
					`{"QueueUrl":%q,"Arn":%q,"QueueName":"test-queue"}`,
					testQueueURL, testQueueARN,
				)),
			},
		}),
	)

	spec := core.MappingNodeFields(
		"queueName", core.MappingNodeFromString("test-queue"),
		"visibilityTimeout", core.MappingNodeFromInt(30),
		"tags", core.MappingNodeItems(
			core.MappingNodeFields(
				"key", core.MappingNodeFromString("Environment"),
				"value", core.MappingNodeFromString("test"),
			),
		),
	)

	expectedDesiredState := `{
		"QueueName": "test-queue",
		"VisibilityTimeout": 30,
		"Tags": [{"Key": "Environment", "Value": "test"}]
	}`

	return plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:             "creates a resource and captures its computed fields at config-complete",
		ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore:      newAWSConfigStore(loader),
		Input:            deployInput(providerCtx, spec, nil),
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				fmt.Sprintf("spec.%s", fieldRequestToken):      core.MappingNodeFromString(testRequestToken),
				fmt.Sprintf("spec.%s", fieldPrimaryIdentifier): core.MappingNodeFromString(testQueueURL),
				"spec.queueUrl": core.MappingNodeFromString(testQueueURL),
				"spec.arn":      core.MappingNodeFromString(testQueueARN),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateResource": matchCreateResourceInput("AWS::SQS::Queue", expectedDesiredState),
		},
	}
}

// Covers a long-running async create: Cloud Control reports
// IN_PROGRESS but the request status already carries the identifier and a resource
// model with every computed field. The deploy captures computed fields at
// config-complete from that model without waiting for SUCCESS and without a separate
// GetResource and threads the request token so HasStabilised finishes the operation.
func createEarlyCaptureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service] {
	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithCreateResourceOutput(&awscc.CreateResourceOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				RequestToken: aws.String(testRequestToken),
				Identifier:   aws.String(testQueueURL),
			},
		}),
		cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				OperationStatus: cctypes.OperationStatusInProgress,
				Identifier:      aws.String(testQueueURL),
				ResourceModel: aws.String(fmt.Sprintf(
					`{"QueueUrl":%q,"Arn":%q,"QueueName":"test-queue"}`,
					testQueueURL, testQueueARN,
				)),
			},
		}),
	)

	spec := core.MappingNodeFields("queueName", core.MappingNodeFromString("test-queue"))

	return plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:             "captures computed fields from the in-progress resource model without a GetResource",
		ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore:      newAWSConfigStore(loader),
		Input:            deployInput(providerCtx, spec, nil),
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				fmt.Sprintf("spec.%s", fieldRequestToken):      core.MappingNodeFromString(testRequestToken),
				fmt.Sprintf("spec.%s", fieldPrimaryIdentifier): core.MappingNodeFromString(testQueueURL),
				"spec.queueUrl": core.MappingNodeFromString(testQueueURL),
				"spec.arn":      core.MappingNodeFromString(testQueueARN),
			},
		},
		SaveActionsNotCalled: []string{"GetResource"},
	}
}

// Covers a SUCCESS event whose resource model echoes the
// submitted desired state without the read-only computed fields (no Arn).
// The deploy must not trust the incomplete model: it falls
// back to a GetResource read so the ARN (referenced by dependants such as a Lambda's
// role) is captured at config-complete rather than resolving to null.
func createIncompleteModelTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service] {
	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithCreateResourceOutput(&awscc.CreateResourceOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				RequestToken: aws.String(testRequestToken),
				Identifier:   aws.String(testQueueURL),
			},
		}),
		// SUCCESS but the model is missing the Arn computed field, so the deploy must
		// fall back to GetResource for the authoritative, complete state.
		cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				OperationStatus: cctypes.OperationStatusSuccess,
				Identifier:      aws.String(testQueueURL),
				ResourceModel: aws.String(fmt.Sprintf(
					`{"QueueUrl":%q,"QueueName":"test-queue"}`,
					testQueueURL,
				)),
			},
		}),
		cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
			ResourceDescription: &cctypes.ResourceDescription{
				Identifier: aws.String(testQueueURL),
				Properties: aws.String(fmt.Sprintf(
					`{"QueueUrl":%q,"Arn":%q,"QueueName":"test-queue"}`,
					testQueueURL, testQueueARN,
				)),
			},
		}),
	)

	spec := core.MappingNodeFields("queueName", core.MappingNodeFromString("test-queue"))

	return plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:             "falls back to GetResource when the success model omits computed fields",
		ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore:      newAWSConfigStore(loader),
		Input:            deployInput(providerCtx, spec, nil),
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				fmt.Sprintf("spec.%s", fieldRequestToken):      core.MappingNodeFromString(testRequestToken),
				fmt.Sprintf("spec.%s", fieldPrimaryIdentifier): core.MappingNodeFromString(testQueueURL),
				"spec.queueUrl": core.MappingNodeFromString(testQueueURL),
				"spec.arn":      core.MappingNodeFromString(testQueueARN),
			},
		},
	}
}

// Covers a SUCCESS event whose model carries a
// computed field as a present-but-empty placeholder ("Arn":""). Cloud Control can surface
// a read-only property before its value is assigned. The empty value must not be trusted:
// the deploy falls back to GetResource so the real ARN is captured at config-complete
// rather than persisting a null/empty value that a dependant would resolve to null.
func createEmptyComputedFieldModelTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service] {
	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithCreateResourceOutput(&awscc.CreateResourceOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				RequestToken: aws.String(testRequestToken),
				Identifier:   aws.String(testQueueURL),
			},
		}),
		cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				OperationStatus: cctypes.OperationStatusSuccess,
				Identifier:      aws.String(testQueueURL),
				ResourceModel: aws.String(fmt.Sprintf(
					`{"QueueUrl":%q,"Arn":"","QueueName":"test-queue"}`,
					testQueueURL,
				)),
			},
		}),
		cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
			ResourceDescription: &cctypes.ResourceDescription{
				Identifier: aws.String(testQueueURL),
				Properties: aws.String(fmt.Sprintf(
					`{"QueueUrl":%q,"Arn":%q,"QueueName":"test-queue"}`,
					testQueueURL, testQueueARN,
				)),
			},
		}),
	)

	spec := core.MappingNodeFields("queueName", core.MappingNodeFromString("test-queue"))

	return plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:             "falls back to GetResource when the success model has an empty computed field",
		ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore:      newAWSConfigStore(loader),
		Input:            deployInput(providerCtx, spec, nil),
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				fmt.Sprintf("spec.%s", fieldRequestToken):      core.MappingNodeFromString(testRequestToken),
				fmt.Sprintf("spec.%s", fieldPrimaryIdentifier): core.MappingNodeFromString(testQueueURL),
				"spec.queueUrl": core.MappingNodeFromString(testQueueURL),
				"spec.arn":      core.MappingNodeFromString(testQueueARN),
			},
		},
	}
}

func createFailureTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service] {
	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithCreateResourceError(errors.New("create failed")),
	)

	spec := core.MappingNodeFields("queueName", core.MappingNodeFromString("test-queue"))

	return plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:           "returns an error when create fails",
		ServiceFactory: func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ConfigStore:    newAWSConfigStore(loader),
		Input:          deployInput(providerCtx, spec, nil),
		ExpectError:    true,
	}
}

func matchCreateResourceInput(
	typeName string,
	expectedDesiredState string,
) func(arg any) (plugintestutils.EqualityCheckValues, error) {
	return func(arg any) (plugintestutils.EqualityCheckValues, error) {
		input, ok := arg.(*awscc.CreateResourceInput)
		if !ok {
			return plugintestutils.EqualityCheckValues{}, fmt.Errorf("expected *CreateResourceInput, got %T", arg)
		}

		var expectedState, actualState any
		if err := json.Unmarshal([]byte(expectedDesiredState), &expectedState); err != nil {
			return plugintestutils.EqualityCheckValues{}, err
		}
		if err := json.Unmarshal([]byte(aws.ToString(input.DesiredState)), &actualState); err != nil {
			return plugintestutils.EqualityCheckValues{}, err
		}

		return plugintestutils.EqualityCheckValues{
			Expected: map[string]any{
				"typeName":       typeName,
				"desiredState":   expectedState,
				"hasClientToken": true,
			},
			Actual: map[string]any{
				"typeName":       aws.ToString(input.TypeName),
				"desiredState":   actualState,
				"hasClientToken": input.ClientToken != nil && *input.ClientToken != "",
			},
		}, nil
	}
}

func deployInput(
	providerCtx provider.Context,
	spec *core.MappingNode,
	currentState *core.MappingNode,
) *provider.ResourceDeployInput {
	info := provider.ResourceInfo{
		ResourceID:   "test-resource-id",
		ResourceName: "TestQueue",
		InstanceID:   "test-instance-id",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Type: &schema.ResourceTypeWrapper{Value: "aws/sqs/queue"},
			Spec: spec,
		},
	}
	if currentState != nil {
		info.CurrentResourceState = &state.ResourceState{SpecData: currentState}
	}
	return &provider.ResourceDeployInput{
		InstanceID:      "test-instance-id",
		ResourceID:      "test-resource-id",
		Changes:         &provider.Changes{AppliedResourceInfo: info},
		ProviderContext: providerCtx,
	}
}

func TestCCResourceCreateSuite(t *testing.T) {
	suite.Run(t, new(CCResourceCreateSuite))
}
