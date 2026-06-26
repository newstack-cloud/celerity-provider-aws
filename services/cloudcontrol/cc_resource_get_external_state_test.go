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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type CCResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *CCResourceGetExternalStateSuite) Test_get_external_state() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
		map[string]*core.ScalarValue{"session_id": core.ScalarFromString("test-session-id")},
	)

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, cloudcontrolservice.Service]{
		getStateFromIdentifierCase(providerCtx, loader),
		getStateNotFoundCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(testCases, newTestResource, &s.Suite)
}

func getStateFromIdentifierCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, cloudcontrolservice.Service] {
	properties := `{
		"QueueName": "test-queue",
		"VisibilityTimeout": 30,
		"QueueUrl": "` + testQueueURL + `",
		"Tags": [
			{"Key": "Environment", "Value": "test"},
			{"Key": "bluelink:instance-id", "Value": "test-instance-id"}
		]
	}`

	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
			ResourceDescription: &cctypes.ResourceDescription{
				Identifier: aws.String(testQueueURL),
				Properties: aws.String(properties),
			},
		}),
	)

	currentSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			fieldPrimaryIdentifier: core.MappingNodeFromString(testQueueURL),
		},
	}

	expected := core.MappingNodeFields(
		"queueName", core.MappingNodeFromString("test-queue"),
		"visibilityTimeout", core.MappingNodeFromInt(30),
		"queueUrl", core.MappingNodeFromString(testQueueURL),
		"tags", core.MappingNodeItems(
			core.MappingNodeFields(
				"key", core.MappingNodeFromString("Environment"),
				"value", core.MappingNodeFromString("test"),
			),
		),
		fieldPrimaryIdentifier, core.MappingNodeFromString(testQueueURL),
	)

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:           "reads, translates and tag-filters external state by identifier",
		ServiceFactory: func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ConfigStore:    newAWSConfigStore(loader),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-resource-id",
			ResourceName:        "TestQueue",
			CurrentResourceSpec: currentSpec,
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: expected,
		},
	}
}

func getStateNotFoundCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, cloudcontrolservice.Service] {
	service := cloudcontrolmock.CreateCloudControlServiceMock()

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:           "returns empty state when no identifier and tag lookup finds nothing",
		ServiceFactory: func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ConfigStore:    newAWSConfigStore(loader),
		Input: &provider.ResourceGetExternalStateInput{
			InstanceID:          "test-instance-id",
			ResourceID:          "test-resource-id",
			ResourceName:        "TestQueue",
			CurrentResourceSpec: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
			ProviderContext:     providerCtx,
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
		},
	}
}

func TestCCResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(CCResourceGetExternalStateSuite))
}
