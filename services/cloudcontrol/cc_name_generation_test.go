//go:build unit

package cloudcontrol

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscc "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	cctypes "github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	cloudcontrolmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/cloudcontrol_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/overlays"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

const nameGenTestType = "aws/test/named"

func init() {
	overlays.RegisterBehaviour(nameGenTestType, &overlays.Behaviour{
		Name: &overlays.NameGeneration{
			Field:    "widgetName",
			Generate: utils.DefaultUniqueNameGenerator(40),
		},
	})
}

type CCNameGenerationSuite struct {
	suite.Suite
}

// Verifies a behaviour overlay's name generator
// populates the name field in the Cloud Control desired state when the user omits it.
func (s *CCNameGenerationSuite) Test_generates_name_when_omitted() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
		map[string]*core.ScalarValue{"session_id": core.ScalarFromString("test-session-id")},
	)

	const widgetARN = "arn:aws:test:::widget/x"
	service := cloudcontrolmock.CreateCloudControlServiceMock(
		cloudcontrolmock.WithCreateResourceOutput(&awscc.CreateResourceOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				RequestToken: aws.String("tok"),
				Identifier:   aws.String(widgetARN),
			},
		}),
		cloudcontrolmock.WithGetResourceRequestStatusOutput(&awscc.GetResourceRequestStatusOutput{
			ProgressEvent: &cctypes.ProgressEvent{
				OperationStatus: cctypes.OperationStatusSuccess,
				Identifier:      aws.String(widgetARN),
			},
		}),
		cloudcontrolmock.WithGetResourceOutput(&awscc.GetResourceOutput{
			ResourceDescription: &cctypes.ResourceDescription{
				Identifier: aws.String(widgetARN),
				Properties: aws.String(fmt.Sprintf(`{"Arn":%q,"WidgetName":"TestQueue-x"}`, widgetARN)),
			},
		}),
	)

	testCase := plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{
		Name:             "generates a name when omitted",
		ServiceFactory:   func(*aws.Config, provider.Context) cloudcontrolservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore:      newAWSConfigStore(loader),
		Input:            deployInput(providerCtx, &core.MappingNode{Fields: map[string]*core.MappingNode{}}, nil),
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec." + fieldRequestToken:      core.MappingNodeFromString("tok"),
				"spec." + fieldPrimaryIdentifier: core.MappingNodeFromString(widgetARN),
				"spec.arn":                       core.MappingNodeFromString(widgetARN),
			},
		},
		SaveActionsCalled: map[string]any{
			"CreateResource": matchGeneratedName,
		},
	}

	createResource := func(
		serviceFactory pluginutils.ServiceFactory[*aws.Config, cloudcontrolservice.Service],
		configStore pluginutils.ServiceConfigStore[*aws.Config],
	) provider.Resource {
		return CCResource(namedTestConfig(), serviceFactory, mockResourceGroupTaggingServiceFactory, configStore)
	}

	plugintestutils.RunResourceDeployTestCases(
		[]plugintestutils.ResourceDeployTestCase[*aws.Config, cloudcontrolservice.Service]{testCase},
		createResource,
		&s.Suite,
	)
}

func matchGeneratedName(arg any) (plugintestutils.EqualityCheckValues, error) {
	input, ok := arg.(*awscc.CreateResourceInput)
	if !ok {
		return plugintestutils.EqualityCheckValues{}, fmt.Errorf("expected *CreateResourceInput, got %T", arg)
	}
	var desired map[string]any
	if err := json.Unmarshal([]byte(aws.ToString(input.DesiredState)), &desired); err != nil {
		return plugintestutils.EqualityCheckValues{}, err
	}
	name, _ := desired["WidgetName"].(string)
	// The default generator produces "<resource>-<nanoid>" when no instance name.
	generated := name != "" && strings.Contains(name, "TestQueue")
	return plugintestutils.EqualityCheckValues{Expected: true, Actual: generated}, nil
}

func namedTestConfig() CCResourceConfig {
	return CCResourceConfig{
		BlueprintType: nameGenTestType,
		CFNType:       "AWS::Test::Widget",
		Label:         "Test Widget",
		Schema: &provider.ResourceDefinitionsSchema{
			Type:  provider.ResourceDefinitionsSchemaTypeObject,
			Label: "Test Widget",
			Attributes: map[string]*provider.ResourceDefinitionsSchema{
				"widgetName": {
					Type:         provider.ResourceDefinitionsSchemaTypeString,
					Label:        "Widget Name",
					MustRecreate: true,
					Nullable:     true,
				},
				"arn": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "ARN", Computed: true},
			},
		},
		Meta: CCResourceMeta{PrimaryIdentifierField: "arn", ComputedFields: []string{"arn"}},
	}
}

func TestCCNameGenerationSuite(t *testing.T) {
	suite.Run(t, new(CCNameGenerationSuite))
}
