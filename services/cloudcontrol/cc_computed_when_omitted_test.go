//go:build unit

package cloudcontrol

import (
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
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type CCComputedWhenOmittedSuite struct {
	suite.Suite
}

// A widget with an auto-named identifier: widgetName is user-settable but AWS
// assigns a name when it is omitted, so the engine must report the assigned
// value as computed for the deployment.
func autoNamedTestConfig() CCResourceConfig {
	return CCResourceConfig{
		BlueprintType: "aws/test/autoNamed",
		CFNType:       "AWS::Test::AutoNamedWidget",
		Label:         "Auto-Named Test Widget",
		Schema: &provider.ResourceDefinitionsSchema{
			Type:  provider.ResourceDefinitionsSchemaTypeObject,
			Label: "Auto-Named Test Widget",
			Attributes: map[string]*provider.ResourceDefinitionsSchema{
				"widgetName": {
					Type:                provider.ResourceDefinitionsSchemaTypeString,
					Label:               "Widget Name",
					MustRecreate:        true,
					Nullable:            true,
					ComputedWhenOmitted: true,
				},
				"arn": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "ARN", Computed: true},
			},
		},
		Meta: CCResourceMeta{
			PrimaryIdentifierField:  "widgetName",
			PrimaryIdentifierFields: []string{"widgetName"},
			ComputedFields:          []string{"arn"},
		},
	}
}

func (s *CCComputedWhenOmittedSuite) Test_capture_of_auto_assigned_name() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
		map[string]*core.ScalarValue{"session_id": core.ScalarFromString("test-session-id")},
	)

	const assignedName = "auto-named-widget-8f2k1"
	const widgetARN = "arn:aws:test:::widget/auto-named-widget-8f2k1"
	newServiceOpts := func() []cloudcontrolmock.CloudControlServiceMockOption {
		return []cloudcontrolmock.CloudControlServiceMockOption{
			cloudcontrolmock.WithCreateResourceOutput(&awscc.CreateResourceOutput{
				ProgressEvent: &cctypes.ProgressEvent{
					RequestToken: aws.String("tok"),
					Identifier:   aws.String(assignedName),
				},
			}),
			cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
				ProgressEvent: &cctypes.ProgressEvent{
					OperationStatus: cctypes.OperationStatusSuccess,
					Identifier:      aws.String(assignedName),
				},
			}),
			cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
				ResourceDescription: &cctypes.ResourceDescription{
					Identifier: aws.String(assignedName),
					Properties: aws.String(fmt.Sprintf(
						`{"Arn":%q,"WidgetName":%q}`, widgetARN, assignedName,
					)),
				},
			}),
		}
	}

	omittedService := cloudcontrolmock.CreateCloudControlServiceMock(newServiceOpts()...)
	explicitService := cloudcontrolmock.CreateCloudControlServiceMock(newServiceOpts()...)
	explicitName := "explicit-widget"

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		{
			Name:             "captures the AWS-assigned name when omitted",
			ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return omittedService },
			ServiceMockCalls: &omittedService.MockCalls,
			ConfigStore:      newAWSConfigStore(loader),
			Input:            deployInput(providerCtx, &core.MappingNode{Fields: map[string]*core.MappingNode{}}, nil),
			ExpectedOutput: &provider.ResourceDeployOutput{
				ComputedFieldValues: map[string]*core.MappingNode{
					"spec." + fieldRequestToken:      core.MappingNodeFromString("tok"),
					"spec." + fieldPrimaryIdentifier: core.MappingNodeFromString(assignedName),
					"spec.arn":                       core.MappingNodeFromString(widgetARN),
					"spec.widgetName":                core.MappingNodeFromString(assignedName),
				},
			},
		},
		{
			Name:             "does not report an explicitly set name as computed",
			ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return explicitService },
			ServiceMockCalls: &explicitService.MockCalls,
			ConfigStore:      newAWSConfigStore(loader),
			Input: deployInput(providerCtx, &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"widgetName": core.MappingNodeFromString(explicitName),
				},
			}, nil),
			ExpectedOutput: &provider.ResourceDeployOutput{
				ComputedFieldValues: map[string]*core.MappingNode{
					"spec." + fieldRequestToken:      core.MappingNodeFromString("tok"),
					"spec." + fieldPrimaryIdentifier: core.MappingNodeFromString(assignedName),
					"spec.arn":                       core.MappingNodeFromString(widgetARN),
				},
			},
		},
	}

	createResource := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return CCResource(autoNamedTestConfig(), serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDeployTestCases(testCases, createResource, &s.Suite)
}

func TestCCComputedWhenOmittedSuite(t *testing.T) {
	suite.Run(t, new(CCComputedWhenOmittedSuite))
}
