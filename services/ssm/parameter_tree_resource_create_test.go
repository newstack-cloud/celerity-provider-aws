//go:build unit

package ssm

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ssmmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ssm_mock"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type SSMParameterTreeResourceCreateSuite struct {
	suite.Suite
}

const testTreePath = "/my-app/config"

func (s *SSMParameterTreeResourceCreateSuite) Test_create() {
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

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		createTreeTestCase(providerCtx, loader),
		createTreeExistingParameterTestCase(providerCtx, loader),
		createTreeMissingPathTestCase(providerCtx, loader),
		createTreePutParameterErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		ParameterTreeResource,
		&s.Suite,
	)
}

func createTreeTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithPutParameterOutput(&ssm.PutParameterOutput{Version: 1}),
		ssmmock.WithDescribeParametersOutput(describeTreeParametersOutput()),
	)

	resourceSpecData := treeSpecData()

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		Name:             "creates one parameter per entry in sorted key order",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input:          deployInput(providerCtx, resourceSpecData),
		ExpectedOutput: expectedTreeDeployOutput(),
		SaveActionsCalled: map[string]any{
			// Sorted key order: apiToken, db/host, logLevel.
			"PutParameter": []any{
				func(arg any) bool {
					in, ok := arg.(*ssm.PutParameterInput)
					return ok &&
						aws.ToString(in.Name) == testTreePath+"/apiToken" &&
						in.Type == ssmtypes.ParameterTypeSecureString &&
						aws.ToString(in.Value) == "super-secret" &&
						aws.ToString(in.KeyId) == "alias/my-key" &&
						aws.ToBool(in.Overwrite) == false &&
						putParameterHasTag(in, "Environment", "production")
				},
				func(arg any) bool {
					in, ok := arg.(*ssm.PutParameterInput)
					return ok &&
						aws.ToString(in.Name) == testTreePath+"/db/host" &&
						in.Type == ssmtypes.ParameterTypeString &&
						aws.ToString(in.Value) == "db.internal.example.com" &&
						in.KeyId == nil &&
						aws.ToBool(in.Overwrite) == false
				},
				func(arg any) bool {
					in, ok := arg.(*ssm.PutParameterInput)
					return ok &&
						aws.ToString(in.Name) == testTreePath+"/logLevel" &&
						in.Type == ssmtypes.ParameterTypeString &&
						aws.ToString(in.Value) == "info" &&
						aws.ToBool(in.Overwrite) == false
				},
			},
		},
		ExpectError: false,
	}
}

func createTreeExistingParameterTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithPutParameterError(&ssmtypes.ParameterAlreadyExists{}),
		ssmmock.WithListTagsForResourceOutput(&ssm.ListTagsForResourceOutput{}),
		ssmmock.WithAddTagsToResourceOutput(&ssm.AddTagsToResourceOutput{}),
		ssmmock.WithDescribeParametersOutput(&ssm.DescribeParametersOutput{
			Parameters: []ssmtypes.ParameterMetadata{
				{
					Name: aws.String(testTreePath + "/logLevel"),
					ARN:  aws.String(treeParameterARN("logLevel")),
					Type: ssmtypes.ParameterTypeString,
				},
			},
		}),
	)

	resourceSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"path": core.MappingNodeFromString(testTreePath),
			"values": core.MappingNodeFields(
				"logLevel", core.MappingNodeFromString("info"),
			),
			"tags": core.MappingNodeFields(
				"Environment", core.MappingNodeFromString("production"),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		Name:             "preserves a pre-existing parameter value and syncs only tags",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: deployInput(providerCtx, resourceSpecData),
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.parameters": {
					Fields: map[string]*core.MappingNode{
						"logLevel": {
							Fields: map[string]*core.MappingNode{
								"arn":       core.MappingNodeFromString(treeParameterARN("logLevel")),
								"type":      core.MappingNodeFromString("String"),
								"valueHash": core.MappingNodeFromString(parameterTreeValueHash("info")),
							},
						},
					},
				},
			},
		},
		SaveActionsCalled: map[string]any{
			"AddTagsToResource": func(arg any) bool {
				in, ok := arg.(*ssm.AddTagsToResourceInput)
				return ok && aws.ToString(in.ResourceId) == testTreePath+"/logLevel"
			},
		},
		ExpectError: false,
	}
}

func createTreeMissingPathTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock()

	resourceSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"values": core.MappingNodeFields(
				"logLevel", core.MappingNodeFromString("info"),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		Name:             "returns error when path is missing",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input:                deployInput(providerCtx, resourceSpecData),
		ExpectedOutput:       nil,
		SaveActionsNotCalled: []string{"PutParameter"},
		ExpectError:          true,
	}
}

func createTreePutParameterErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithPutParameterError(errTestPutParameter),
	)

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		Name:             "returns error when PutParameter fails",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input:                deployInput(providerCtx, treeSpecData()),
		ExpectedOutput:       nil,
		SaveActionsNotCalled: []string{"DescribeParameters"},
		ExpectError:          true,
	}
}

func treeSpecData() *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"path": core.MappingNodeFromString(testTreePath),
			"values": core.MappingNodeFields(
				"logLevel", core.MappingNodeFromString("info"),
				"db/host", core.MappingNodeFromString("db.internal.example.com"),
			),
			"secureValues": core.MappingNodeFields(
				"apiToken", core.MappingNodeFromString("super-secret"),
			),
			"keyId": core.MappingNodeFromString("alias/my-key"),
			"tags": core.MappingNodeFields(
				"Environment", core.MappingNodeFromString("production"),
			),
		},
	}
}

func treeParameterARN(key string) string {
	return "arn:aws:ssm:us-west-2:123456789012:parameter" + testTreePath + "/" + key
}

func describeTreeParametersOutput() *ssm.DescribeParametersOutput {
	return &ssm.DescribeParametersOutput{
		Parameters: []ssmtypes.ParameterMetadata{
			{
				Name: aws.String(testTreePath + "/apiToken"),
				ARN:  aws.String(treeParameterARN("apiToken")),
				Type: ssmtypes.ParameterTypeSecureString,
			},
			{
				Name: aws.String(testTreePath + "/db/host"),
				ARN:  aws.String(treeParameterARN("db/host")),
				Type: ssmtypes.ParameterTypeString,
			},
			{
				Name: aws.String(testTreePath + "/logLevel"),
				ARN:  aws.String(treeParameterARN("logLevel")),
				Type: ssmtypes.ParameterTypeString,
			},
			{
				Name: aws.String(testTreePath + "/foreign"),
				ARN:  aws.String(treeParameterARN("foreign")),
				Type: ssmtypes.ParameterTypeString,
			},
		},
	}
}

func expectedTreeDeployOutput() *provider.ResourceDeployOutput {
	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.parameters": {
				Fields: map[string]*core.MappingNode{
					"apiToken": {
						Fields: map[string]*core.MappingNode{
							"arn":       core.MappingNodeFromString(treeParameterARN("apiToken")),
							"type":      core.MappingNodeFromString("SecureString"),
							"valueHash": core.MappingNodeFromString(parameterTreeValueHash("super-secret")),
						},
					},
					"db/host": {
						Fields: map[string]*core.MappingNode{
							"arn":       core.MappingNodeFromString(treeParameterARN("db/host")),
							"type":      core.MappingNodeFromString("String"),
							"valueHash": core.MappingNodeFromString(parameterTreeValueHash("db.internal.example.com")),
						},
					},
					"logLevel": {
						Fields: map[string]*core.MappingNode{
							"arn":       core.MappingNodeFromString(treeParameterARN("logLevel")),
							"type":      core.MappingNodeFromString("String"),
							"valueHash": core.MappingNodeFromString(parameterTreeValueHash("info")),
						},
					},
				},
			},
		},
	}
}

func TestSSMParameterTreeResourceCreateSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterTreeResourceCreateSuite))
}
