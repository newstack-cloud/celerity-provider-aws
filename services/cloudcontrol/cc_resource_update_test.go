//go:build unit

package cloudcontrol

import (
	"encoding/json"
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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type CCResourceUpdateSuite struct {
	suite.Suite
}

func (s *CCResourceUpdateSuite) Test_update() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
		map[string]*core.ScalarValue{"session_id": core.ScalarFromString("test-session-id")},
	)

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		updateChangedFieldCase(providerCtx, loader),
		updateNoOpCase(providerCtx, loader),
		updateIgnoresUnmanagedReadBackPropertyCase(providerCtx, loader),
		updateNoOpWithUnmanagedReadBackPropertyCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(testCases, newTestResource, &s.Suite)
}

func updateChangedFieldCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service] {
	currentProperties := `{
		"QueueName": "test-queue",
		"VisibilityTimeout": 30,
		"QueueUrl": "` + testQueueURL + `",
		"Arn": "` + testQueueARN + `"
	}`

	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
			ResourceDescription: &cctypes.ResourceDescription{
				Identifier: aws.String(testQueueURL),
				Properties: aws.String(currentProperties),
			},
		}),
		cloudcontrolmock.WithUpdateResourceOutput(&awscc.UpdateResourceOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				RequestToken: aws.String("update-token"),
				Identifier:   aws.String(testQueueURL),
			},
		}),
		// The deploy waits for the async update to complete before reading the
		// computed fields back from the live resource.
		cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				OperationStatus: cctypes.OperationStatusSuccess,
				Identifier:      aws.String(testQueueURL),
			},
		}),
	)

	desiredSpec := core.MappingNodeFields(
		"queueName", core.MappingNodeFromString("test-queue"),
		"visibilityTimeout", core.MappingNodeFromInt(60),
	)

	return plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:             "computes a JSON patch for a changed field and updates",
		ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore:      newAWSConfigStore(loader),
		Input:            deployInput(providerCtx, desiredSpec, currentStateWithIdentifier()),
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec." + fieldRequestToken:      core.MappingNodeFromString("update-token"),
				"spec." + fieldPrimaryIdentifier: core.MappingNodeFromString(testQueueURL),
				"spec.queueUrl":                  core.MappingNodeFromString(testQueueURL),
				"spec.arn":                       core.MappingNodeFromString(testQueueARN),
			},
		},
		SaveActionsCalled: map[string]any{
			"UpdateResource": matchUpdateResourceInput(
				testQueueURL,
				`[{"op": "replace", "path": "/VisibilityTimeout", "value": 60}]`,
			),
		},
	}
}

func updateNoOpCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service] {
	currentProperties := `{
		"QueueName": "test-queue",
		"VisibilityTimeout": 30,
		"QueueUrl": "` + testQueueURL + `",
		"Arn": "` + testQueueARN + `"
	}`

	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
			ResourceDescription: &cctypes.ResourceDescription{
				Identifier: aws.String(testQueueURL),
				Properties: aws.String(currentProperties),
			},
		}),
	)

	desiredSpec := core.MappingNodeFields(
		"queueName", core.MappingNodeFromString("test-queue"),
		"visibilityTimeout", core.MappingNodeFromInt(30),
	)

	return plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:             "skips the update call when nothing changed but still captures computed fields",
		ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore:      newAWSConfigStore(loader),
		Input:            deployInput(providerCtx, desiredSpec, currentStateWithIdentifier()),
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec." + fieldPrimaryIdentifier: core.MappingNodeFromString(testQueueURL),
				"spec.queueUrl":                  core.MappingNodeFromString(testQueueURL),
				"spec.arn":                       core.MappingNodeFromString(testQueueARN),
			},
		},
		SaveActionsNotCalled: []string{"UpdateResource"},
	}
}

// AWS can return read-back properties the vendored schema does not model (e.g.
// AWS::Events::Rule's read-only Id). A patch op rooted at such a property is
// rejected by AWS ("readOnlyProperties ... cannot be updated"), so the patch
// must only ever target properties in the desired schema.
func updateIgnoresUnmanagedReadBackPropertyCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service] {
	currentProperties := `{
		"QueueName": "test-queue",
		"VisibilityTimeout": 30,
		"QueueUrl": "` + testQueueURL + `",
		"Arn": "` + testQueueARN + `",
		"Id": "unmanaged-read-back-id"
	}`

	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
			ResourceDescription: &cctypes.ResourceDescription{
				Identifier: aws.String(testQueueURL),
				Properties: aws.String(currentProperties),
			},
		}),
		cloudcontrolmock.WithUpdateResourceOutput(&awscc.UpdateResourceOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				RequestToken: aws.String("update-token"),
				Identifier:   aws.String(testQueueURL),
			},
		}),
		cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				OperationStatus: cctypes.OperationStatusSuccess,
				Identifier:      aws.String(testQueueURL),
			},
		}),
	)

	desiredSpec := core.MappingNodeFields(
		"queueName", core.MappingNodeFromString("test-queue"),
		"visibilityTimeout", core.MappingNodeFromInt(60),
	)

	return plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:             "excludes properties the schema does not model from the update patch",
		ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore:      newAWSConfigStore(loader),
		Input:            deployInput(providerCtx, desiredSpec, currentStateWithIdentifier()),
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec." + fieldRequestToken:      core.MappingNodeFromString("update-token"),
				"spec." + fieldPrimaryIdentifier: core.MappingNodeFromString(testQueueURL),
				"spec.queueUrl":                  core.MappingNodeFromString(testQueueURL),
				"spec.arn":                       core.MappingNodeFromString(testQueueARN),
			},
		},
		SaveActionsCalled: map[string]any{
			"UpdateResource": matchUpdateResourceInput(
				testQueueURL,
				`[{"op": "replace", "path": "/VisibilityTimeout", "value": 60}]`,
			),
		},
	}
}

func updateNoOpWithUnmanagedReadBackPropertyCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service] {
	currentProperties := `{
		"QueueName": "test-queue",
		"VisibilityTimeout": 30,
		"QueueUrl": "` + testQueueURL + `",
		"Arn": "` + testQueueARN + `",
		"Id": "unmanaged-read-back-id"
	}`

	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
			ResourceDescription: &cctypes.ResourceDescription{
				Identifier: aws.String(testQueueURL),
				Properties: aws.String(currentProperties),
			},
		}),
	)

	desiredSpec := core.MappingNodeFields(
		"queueName", core.MappingNodeFromString("test-queue"),
		"visibilityTimeout", core.MappingNodeFromInt(30),
	)

	return plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:             "treats an unmanaged read-back property as no change",
		ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore:      newAWSConfigStore(loader),
		Input:            deployInput(providerCtx, desiredSpec, currentStateWithIdentifier()),
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec." + fieldPrimaryIdentifier: core.MappingNodeFromString(testQueueURL),
				"spec.queueUrl":                  core.MappingNodeFromString(testQueueURL),
				"spec.arn":                       core.MappingNodeFromString(testQueueARN),
			},
		},
		SaveActionsNotCalled: []string{"UpdateResource"},
	}
}

func currentStateWithIdentifier() *core.MappingNode {
	return core.MappingNodeFields(
		fieldPrimaryIdentifier, core.MappingNodeFromString(testQueueURL),
		"queueName", core.MappingNodeFromString("test-queue"),
		"visibilityTimeout", core.MappingNodeFromInt(30),
		"queueUrl", core.MappingNodeFromString(testQueueURL),
	)
}

func matchUpdateResourceInput(
	identifier string,
	expectedPatch string,
) func(arg any) (plugintestutils.EqualityCheckValues, error) {
	return func(arg any) (plugintestutils.EqualityCheckValues, error) {
		input, ok := arg.(*awscc.UpdateResourceInput)
		if !ok {
			return plugintestutils.EqualityCheckValues{}, fmt.Errorf("expected *UpdateResourceInput, got %T", arg)
		}

		var expectedOps, actualOps any
		if err := json.Unmarshal([]byte(expectedPatch), &expectedOps); err != nil {
			return plugintestutils.EqualityCheckValues{}, err
		}
		if err := json.Unmarshal([]byte(aws.ToString(input.PatchDocument)), &actualOps); err != nil {
			return plugintestutils.EqualityCheckValues{}, err
		}

		return plugintestutils.EqualityCheckValues{
			Expected: map[string]any{
				"typeName":   "AWS::SQS::Queue",
				"identifier": identifier,
				"patch":      expectedOps,
			},
			Actual: map[string]any{
				"typeName":   aws.ToString(input.TypeName),
				"identifier": aws.ToString(input.Identifier),
				"patch":      actualOps,
			},
		}, nil
	}
}

func TestCCResourceUpdateSuite(t *testing.T) {
	suite.Run(t, new(CCResourceUpdateSuite))
}
