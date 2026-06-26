//go:build unit

package cloudcontrol

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscc "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	cctypes "github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	cloudcontrolmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/cloudcontrol_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type CCResourceDestroySuite struct {
	suite.Suite
}

func (s *CCResourceDestroySuite) Test_destroy() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
		map[string]*core.ScalarValue{"session_id": core.ScalarFromString("test-session-id")},
	)

	testCases := []plugintestutils.ResourceDestroyTestCase[*aws.Config, cloudcontrolservice.Service]{
		destroySuccessCase(providerCtx, loader),
		destroyFailureCase(providerCtx, loader),
		destroyNoIdentifierCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(testCases, newTestResource, &s.Suite)
}

func destroySuccessCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, cloudcontrolservice.Service] {
	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithDeleteResourceOutput(&awscc.DeleteResourceOutput{
			ProgressEvent: &cctypes.ProgressEvent{RequestToken: aws.String("delete-token")},
		}),
		cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
			ProgressEvent: &cctypes.ProgressEvent{OperationStatus: cctypes.OperationStatusSuccess},
		}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:             "deletes and waits for terminal success",
		ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore:      newAWSConfigStore(loader),
		Input:            destroyInput(providerCtx, testQueueURL),
		DestroyActionsCalled: map[string]any{
			"DeleteResource": func(arg any) bool {
				input, ok := arg.(*awscc.DeleteResourceInput)
				return ok &&
					aws.ToString(input.TypeName) == "AWS::SQS::Queue" &&
					aws.ToString(input.Identifier) == testQueueURL
			},
		},
	}
}

func destroyFailureCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, cloudcontrolservice.Service] {
	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithDeleteResourceOutput(&awscc.DeleteResourceOutput{
			ProgressEvent: &cctypes.ProgressEvent{RequestToken: aws.String("delete-token")},
		}),
		cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				OperationStatus: cctypes.OperationStatusFailed,
				ErrorCode:       cctypes.HandlerErrorCodeGeneralServiceException,
				StatusMessage:   aws.String("boom"),
			},
		}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:           "errors when deletion reaches a failed status",
		ServiceFactory: func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ConfigStore:    newAWSConfigStore(loader),
		Input:          destroyInput(providerCtx, testQueueURL),
		ExpectError:    true,
	}
}

func destroyNoIdentifierCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, cloudcontrolservice.Service] {
	service := cloudcontrolmock.CreateCloudControlServiceMock()

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:             "no-op when there is no recorded identifier",
		ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore:      newAWSConfigStore(loader),
		Input: &provider.ResourceDestroyInput{
			InstanceID:      "test-instance-id",
			ResourceID:      "test-resource-id",
			ResourceState:   &state.ResourceState{SpecData: &core.MappingNode{Fields: map[string]*core.MappingNode{}}},
			ProviderContext: providerCtx,
		},
		DestroyActionsNotCalled: []string{"DeleteResource"},
	}
}

func destroyInput(providerCtx provider.Context, identifier string) *provider.ResourceDestroyInput {
	return &provider.ResourceDestroyInput{
		InstanceID: "test-instance-id",
		ResourceID: "test-resource-id",
		ResourceState: &state.ResourceState{
			SpecData: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					fieldPrimaryIdentifier: core.MappingNodeFromString(identifier),
				},
			},
		},
		ProviderContext: providerCtx,
	}
}

func TestCCResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(CCResourceDestroySuite))
}
