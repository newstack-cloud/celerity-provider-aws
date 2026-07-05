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
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type SSMParameterResourceUpdateSuite struct {
	suite.Suite
}

func (s *SSMParameterResourceUpdateSuite) Test_update() {
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
		updateValueAndReconcileTagsTestCase(providerCtx, loader),
		updatePutParameterErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		ParameterResource,
		&s.Suite,
	)
}

func updateValueAndReconcileTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithPutParameterOutput(&ssm.PutParameterOutput{
			Version: 2,
			Tier:    ssmtypes.ParameterTierStandard,
		}),
		// The parameter currently has a stale Environment value and an obsolete tag.
		ssmmock.WithListTagsForResourceOutput(&ssm.ListTagsForResourceOutput{
			TagList: []ssmtypes.Tag{
				{Key: aws.String("Environment"), Value: aws.String("staging")},
				{Key: aws.String("Obsolete"), Value: aws.String("yes")},
			},
		}),
		ssmmock.WithAddTagsToResourceOutput(&ssm.AddTagsToResourceOutput{}),
		ssmmock.WithRemoveTagsFromResourceOutput(&ssm.RemoveTagsFromResourceOutput{}),
		ssmmock.WithGetParameterOutput(&ssm.GetParameterOutput{
			Parameter: &ssmtypes.Parameter{
				ARN:     aws.String(testParameterARN),
				Name:    aws.String(testParameterName),
				Version: 2,
			},
		}),
	)

	resourceSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":  core.MappingNodeFromString(testParameterName),
			"type":  core.MappingNodeFromString("String"),
			"value": core.MappingNodeFromString("db2.internal.example.com"),
			"tier":  core.MappingNodeFromString("Standard"),
			"tags": core.MappingNodeFields(
				"Environment", core.MappingNodeFromString("production"),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		Name:             "updates value and reconciles tags",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input:          updateInput(providerCtx, currentParameterState(), resourceSpecData),
		ExpectedOutput: expectedCreateOutput(testParameterARN, 2),
		SaveActionsCalled: map[string]any{
			"PutParameter": func(arg any) bool {
				in, ok := arg.(*ssm.PutParameterInput)
				return ok &&
					aws.ToString(in.Value) == "db2.internal.example.com" &&
					aws.ToBool(in.Overwrite) == true &&
					len(in.Tags) == 0
			},
			"AddTagsToResource": func(arg any) bool {
				in, ok := arg.(*ssm.AddTagsToResourceInput)
				if !ok || len(in.Tags) != 1 {
					return false
				}
				return in.ResourceType == ssmtypes.ResourceTypeForTaggingParameter &&
					aws.ToString(in.ResourceId) == testParameterName &&
					aws.ToString(in.Tags[0].Key) == "Environment" &&
					aws.ToString(in.Tags[0].Value) == "production"
			},
			"RemoveTagsFromResource": func(arg any) bool {
				in, ok := arg.(*ssm.RemoveTagsFromResourceInput)
				return ok &&
					len(in.TagKeys) == 1 &&
					in.TagKeys[0] == "Obsolete"
			},
		},
		ExpectError: false,
	}
}

func updatePutParameterErrorTestCase(
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
		Input:                updateInput(providerCtx, currentParameterState(), resourceSpecData),
		ExpectedOutput:       nil,
		SaveActionsNotCalled: []string{"ListTagsForResource", "GetParameter"},
		ExpectError:          true,
	}
}

func updateInput(
	providerCtx provider.Context,
	currentSpec *core.MappingNode,
	updatedSpec *core.MappingNode,
) *provider.ResourceDeployInput {
	return &provider.ResourceDeployInput{
		InstanceID: "test-instance-id",
		ResourceID: "test-resource-id",
		Changes: &provider.Changes{
			AppliedResourceInfo: provider.ResourceInfo{
				ResourceID:   "test-resource-id",
				ResourceName: "TestParameter",
				InstanceID:   "test-instance-id",
				CurrentResourceState: &state.ResourceState{
					ResourceID: "test-resource-id",
					Name:       "TestParameter",
					InstanceID: "test-instance-id",
					SpecData:   currentSpec,
				},
				ResourceWithResolvedSubs: &provider.ResolvedResource{
					Type: &schema.ResourceTypeWrapper{
						Value: "aws/ssm/parameter",
					},
					Spec: updatedSpec,
				},
			},
			ModifiedFields: []provider.FieldChange{
				{FieldPath: "spec.value"},
			},
		},
		ProviderContext: providerCtx,
	}
}

func currentParameterState() *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":    core.MappingNodeFromString(testParameterName),
			"type":    core.MappingNodeFromString("String"),
			"value":   core.MappingNodeFromString("db.internal.example.com"),
			"tier":    core.MappingNodeFromString("Standard"),
			"arn":     core.MappingNodeFromString(testParameterARN),
			"version": core.MappingNodeFromInt(1),
		},
	}
}

func TestSSMParameterResourceUpdateSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterResourceUpdateSuite))
}
