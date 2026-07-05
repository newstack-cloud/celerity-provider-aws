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
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type SSMParameterResourceCreateSuite struct {
	suite.Suite
}

const (
	testParameterName = "/my-app/prod/db-host"
	testParameterARN  = "arn:aws:ssm:us-west-2:123456789012:parameter/my-app/prod/db-host"
)

func (s *SSMParameterResourceCreateSuite) Test_create() {
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
		createStringParameterTestCase(providerCtx, loader),
		createSecureStringParameterTestCase(providerCtx, loader),
		createMissingNameTestCase(providerCtx, loader),
		createPutParameterErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		ParameterResource,
		&s.Suite,
	)
}

func createStringParameterTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithPutParameterOutput(&ssm.PutParameterOutput{
			Version: 1,
			Tier:    ssmtypes.ParameterTierStandard,
		}),
		ssmmock.WithGetParameterOutput(&ssm.GetParameterOutput{
			Parameter: &ssmtypes.Parameter{
				ARN:     aws.String(testParameterARN),
				Name:    aws.String(testParameterName),
				Version: 1,
			},
		}),
	)

	resourceSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":     core.MappingNodeFromString(testParameterName),
			"type":     core.MappingNodeFromString("String"),
			"value":    core.MappingNodeFromString("db.internal.example.com"),
			"tier":     core.MappingNodeFromString("Standard"),
			"dataType": core.MappingNodeFromString("text"),
			"tags": core.MappingNodeFields(
				"Environment", core.MappingNodeFromString("production"),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		Name:             "creates a String parameter with tags",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input:          deployInput(providerCtx, resourceSpecData),
		ExpectedOutput: expectedCreateOutput(testParameterARN, 1),
		SaveActionsCalled: map[string]any{
			"PutParameter": func(arg any) bool {
				in, ok := arg.(*ssm.PutParameterInput)
				if !ok {
					return false
				}
				return aws.ToString(in.Name) == testParameterName &&
					in.Type == ssmtypes.ParameterTypeString &&
					aws.ToString(in.Value) == "db.internal.example.com" &&
					aws.ToBool(in.Overwrite) == false &&
					in.KeyId == nil &&
					putParameterHasTag(in, "Environment", "production")
			},
			"GetParameter": func(arg any) bool {
				in, ok := arg.(*ssm.GetParameterInput)
				return ok && aws.ToString(in.Name) == testParameterName
			},
		},
		ExpectError: false,
	}
}

func createSecureStringParameterTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithPutParameterOutput(&ssm.PutParameterOutput{
			Version: 3,
			Tier:    ssmtypes.ParameterTierStandard,
		}),
		ssmmock.WithGetParameterOutput(&ssm.GetParameterOutput{
			Parameter: &ssmtypes.Parameter{
				ARN:     aws.String(testParameterARN),
				Name:    aws.String(testParameterName),
				Version: 3,
			},
		}),
	)

	resourceSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":        core.MappingNodeFromString(testParameterName),
			"type":        core.MappingNodeFromString("SecureString"),
			"secureValue": core.MappingNodeFromString("super-secret"),
			"keyId":       core.MappingNodeFromString("alias/my-key"),
			"tier":        core.MappingNodeFromString("Standard"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		Name:             "creates a SecureString parameter with a customer managed key",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input:          deployInput(providerCtx, resourceSpecData),
		ExpectedOutput: expectedCreateOutput(testParameterARN, 3),
		SaveActionsCalled: map[string]any{
			"PutParameter": func(arg any) bool {
				in, ok := arg.(*ssm.PutParameterInput)
				if !ok {
					return false
				}
				return in.Type == ssmtypes.ParameterTypeSecureString &&
					aws.ToString(in.Value) == "super-secret" &&
					aws.ToString(in.KeyId) == "alias/my-key"
			},
		},
		ExpectError: false,
	}
}

func createMissingNameTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock()

	resourceSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"type":  core.MappingNodeFromString("String"),
			"value": core.MappingNodeFromString("something"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		Name:             "returns error when name is missing",
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

func createPutParameterErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithPutParameterError(errTestPutParameter),
	)

	resourceSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":  core.MappingNodeFromString(testParameterName),
			"type":  core.MappingNodeFromString("String"),
			"value": core.MappingNodeFromString("something"),
		},
	}

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
		Input:                deployInput(providerCtx, resourceSpecData),
		ExpectedOutput:       nil,
		SaveActionsNotCalled: []string{"GetParameter"},
		ExpectError:          true,
	}
}

func deployInput(
	providerCtx provider.Context,
	resourceSpecData *core.MappingNode,
) *provider.ResourceDeployInput {
	return &provider.ResourceDeployInput{
		InstanceID: "test-instance-id",
		ResourceID: "test-resource-id",
		Changes: &provider.Changes{
			AppliedResourceInfo: provider.ResourceInfo{
				ResourceID:   "test-resource-id",
				ResourceName: "TestParameter",
				InstanceID:   "test-instance-id",
				ResourceWithResolvedSubs: &provider.ResolvedResource{
					Type: &schema.ResourceTypeWrapper{
						Value: "aws/ssm/parameter",
					},
					Spec: resourceSpecData,
				},
			},
		},
		ProviderContext: providerCtx,
	}
}

func expectedCreateOutput(arn string, version int) *provider.ResourceDeployOutput {
	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.arn":     core.MappingNodeFromString(arn),
			"spec.version": core.MappingNodeFromInt(version),
		},
	}
}

func putParameterHasTag(in *ssm.PutParameterInput, key, value string) bool {
	for _, tag := range in.Tags {
		if aws.ToString(tag.Key) == key && aws.ToString(tag.Value) == value {
			return true
		}
	}
	return false
}

func TestSSMParameterResourceCreateSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterResourceCreateSuite))
}
