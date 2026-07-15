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

type SSMParameterTreeResourceUpdateSuite struct {
	suite.Suite
}

func (s *SSMParameterTreeResourceUpdateSuite) Test_update() {
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
		updateTreeUnchangedTestCase(providerCtx, loader),
		updateTreeChangedAndNewKeysTestCase(providerCtx, loader),
		updateTreeRemovedKeyTestCase(providerCtx, loader),
		updateTreeTypeMoveTestCase(providerCtx, loader),
		updateTreeSettingsChangePreservesValuesTestCase(providerCtx, loader),
		updateTreeTagsOnlyTestCase(providerCtx, loader),
		updateTreePutParameterErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		ParameterTreeResource,
		&s.Suite,
	)
}

// The override-protection invariant: a key whose blueprint value is unchanged (its hash
// matches the one recorded at the last apply) is never re-put, so out-of-band writes to
// the stored value survive the update untouched.
func updateTreeUnchangedTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
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

	desiredSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"path": core.MappingNodeFromString(testTreePath),
			"values": core.MappingNodeFields(
				"logLevel", core.MappingNodeFromString("info"),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		Name:             "does not re-put a key whose blueprint value is unchanged",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: treeUpdateInput(
			providerCtx,
			treeStateSpecData(map[string]string{"logLevel": "info"}, nil),
			desiredSpec,
			nil,
		),
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
		SaveActionsNotCalled: []string{"PutParameter", "GetParameter", "DeleteParameter"},
		ExpectError:          false,
	}
}

func updateTreeChangedAndNewKeysTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithPutParameterOutput(&ssm.PutParameterOutput{Version: 2}),
		ssmmock.WithListTagsForResourceOutput(&ssm.ListTagsForResourceOutput{}),
		ssmmock.WithAddTagsToResourceOutput(&ssm.AddTagsToResourceOutput{}),
		ssmmock.WithDescribeParametersOutput(&ssm.DescribeParametersOutput{}),
	)

	desiredSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"path": core.MappingNodeFromString(testTreePath),
			"values": core.MappingNodeFields(
				"logLevel", core.MappingNodeFromString("debug"),
				"featureFlag", core.MappingNodeFromString("on"),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		Name:             "re-puts changed and new keys with blueprint values",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: treeUpdateInput(
			providerCtx,
			treeStateSpecData(map[string]string{"logLevel": "info"}, nil),
			desiredSpec,
			[]provider.FieldChange{{FieldPath: "spec.values"}},
		),
		ExpectedOutput: expectedTreeHashOnlyOutput(map[string]string{
			"featureFlag": "on",
			"logLevel":    "debug",
		}),
		SaveActionsCalled: map[string]any{
			// Sorted key order: featureFlag (new, no stored hash), logLevel (hash mismatch).
			"PutParameter": []any{
				func(arg any) bool {
					in, ok := arg.(*ssm.PutParameterInput)
					return ok &&
						aws.ToString(in.Name) == testTreePath+"/featureFlag" &&
						aws.ToString(in.Value) == "on" &&
						aws.ToBool(in.Overwrite) == true &&
						len(in.Tags) == 0
				},
				func(arg any) bool {
					in, ok := arg.(*ssm.PutParameterInput)
					return ok &&
						aws.ToString(in.Name) == testTreePath+"/logLevel" &&
						aws.ToString(in.Value) == "debug" &&
						aws.ToBool(in.Overwrite) == true
				},
			},
		},
		ExpectError: false,
	}
}

func updateTreeRemovedKeyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithDeleteParameterOutput(&ssm.DeleteParameterOutput{}),
		ssmmock.WithDescribeParametersOutput(&ssm.DescribeParametersOutput{}),
	)

	desiredSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"path": core.MappingNodeFromString(testTreePath),
			"values": core.MappingNodeFields(
				"logLevel", core.MappingNodeFromString("info"),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		Name:             "deletes keys removed from the spec",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: treeUpdateInput(
			providerCtx,
			treeStateSpecData(map[string]string{"logLevel": "info", "oldKey": "bye"}, nil),
			desiredSpec,
			[]provider.FieldChange{{FieldPath: "spec.values"}},
		),
		ExpectedOutput: expectedTreeHashOnlyOutput(map[string]string{"logLevel": "info"}),
		SaveActionsCalled: map[string]any{
			"DeleteParameter": func(arg any) bool {
				in, ok := arg.(*ssm.DeleteParameterInput)
				return ok && aws.ToString(in.Name) == testTreePath+"/oldKey"
			},
		},
		SaveActionsNotCalled: []string{"PutParameter"},
		ExpectError:          false,
	}
}

func updateTreeTypeMoveTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithPutParameterOutput(&ssm.PutParameterOutput{Version: 2}),
		ssmmock.WithDescribeParametersOutput(&ssm.DescribeParametersOutput{}),
	)

	desiredSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"path": core.MappingNodeFromString(testTreePath),
			"secureValues": core.MappingNodeFields(
				"apiToken", core.MappingNodeFromString("s3cret"),
			),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		Name:             "re-puts a key that moved from values to secureValues with the new type",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: treeUpdateInput(
			providerCtx,
			treeStateSpecData(map[string]string{"apiToken": "s3cret"}, nil),
			desiredSpec,
			[]provider.FieldChange{
				{FieldPath: "spec.values"},
				{FieldPath: "spec.secureValues"},
			},
		),
		ExpectedOutput: expectedTreeHashOnlyOutput(map[string]string{"apiToken": "s3cret"}),
		SaveActionsCalled: map[string]any{
			"PutParameter": func(arg any) bool {
				in, ok := arg.(*ssm.PutParameterInput)
				return ok &&
					aws.ToString(in.Name) == testTreePath+"/apiToken" &&
					in.Type == ssmtypes.ParameterTypeSecureString &&
					aws.ToBool(in.Overwrite) == true
			},
		},
		ExpectError: false,
	}
}

// When shared settings change, every parameter must be re-put, but re-puts of keys whose
// blueprint value is unchanged carry the current cloud value (fetched transiently) so an
// out-of-band override is not clobbered by a stale blueprint value.
func updateTreeSettingsChangePreservesValuesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithPutParameterOutput(&ssm.PutParameterOutput{Version: 3}),
		ssmmock.WithGetParameterOutput(&ssm.GetParameterOutput{
			Parameter: &ssmtypes.Parameter{
				Name:  aws.String(testTreePath + "/logLevel"),
				Value: aws.String("runtime-override"),
			},
		}),
		ssmmock.WithDescribeParametersOutput(&ssm.DescribeParametersOutput{}),
	)

	desiredSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"path": core.MappingNodeFromString(testTreePath),
			"values": core.MappingNodeFields(
				"logLevel", core.MappingNodeFromString("info"),
			),
			"tier": core.MappingNodeFromString("Advanced"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service]{
		Name:             "settings change re-puts unchanged keys with the preserved cloud value",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: treeUpdateInput(
			providerCtx,
			treeStateSpecData(map[string]string{"logLevel": "info"}, nil),
			desiredSpec,
			[]provider.FieldChange{{FieldPath: "spec.tier"}},
		),
		// The recorded hash tracks the blueprint value, not the preserved cloud value, so
		// a later blueprint edit is still detected as a change.
		ExpectedOutput: expectedTreeHashOnlyOutput(map[string]string{"logLevel": "info"}),
		SaveActionsCalled: map[string]any{
			"GetParameter": func(arg any) bool {
				in, ok := arg.(*ssm.GetParameterInput)
				return ok &&
					aws.ToString(in.Name) == testTreePath+"/logLevel" &&
					aws.ToBool(in.WithDecryption) == true
			},
			"PutParameter": func(arg any) bool {
				in, ok := arg.(*ssm.PutParameterInput)
				return ok &&
					aws.ToString(in.Value) == "runtime-override" &&
					in.Tier == ssmtypes.ParameterTierAdvanced &&
					aws.ToBool(in.Overwrite) == true
			},
		},
		ExpectError: false,
	}
}

func updateTreeTagsOnlyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithListTagsForResourceOutput(&ssm.ListTagsForResourceOutput{}),
		ssmmock.WithAddTagsToResourceOutput(&ssm.AddTagsToResourceOutput{}),
		ssmmock.WithDescribeParametersOutput(&ssm.DescribeParametersOutput{}),
	)

	desiredSpec := &core.MappingNode{
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
		Name:             "tags-only change reconciles tags without touching values",
		ServiceFactory:   func(*aws.Config, provider.Context) ssmservice.Service { return service },
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: treeUpdateInput(
			providerCtx,
			treeStateSpecData(map[string]string{"logLevel": "info"}, nil),
			desiredSpec,
			[]provider.FieldChange{{FieldPath: "spec.tags"}},
		),
		ExpectedOutput: expectedTreeHashOnlyOutput(map[string]string{"logLevel": "info"}),
		SaveActionsCalled: map[string]any{
			"AddTagsToResource": func(arg any) bool {
				in, ok := arg.(*ssm.AddTagsToResourceInput)
				return ok &&
					aws.ToString(in.ResourceId) == testTreePath+"/logLevel" &&
					putTagsContain(in.Tags, "Environment", "production")
			},
		},
		SaveActionsNotCalled: []string{"PutParameter", "GetParameter"},
		ExpectError:          false,
	}
}

func updateTreePutParameterErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, ssmservice.Service] {
	service := ssmmock.CreateSSMServiceMock(
		ssmmock.WithPutParameterError(errTestPutParameter),
	)

	desiredSpec := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"path": core.MappingNodeFromString(testTreePath),
			"values": core.MappingNodeFields(
				"logLevel", core.MappingNodeFromString("debug"),
			),
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
		Input: treeUpdateInput(
			providerCtx,
			treeStateSpecData(map[string]string{"logLevel": "info"}, nil),
			desiredSpec,
			[]provider.FieldChange{{FieldPath: "spec.values"}},
		),
		ExpectedOutput:       nil,
		SaveActionsNotCalled: []string{"DescribeParameters"},
		ExpectError:          true,
	}
}

func expectedTreeHashOnlyOutput(valuesByKey map[string]string) *provider.ResourceDeployOutput {
	parameters := map[string]*core.MappingNode{}
	for key, value := range valuesByKey {
		parameters[key] = &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"valueHash": core.MappingNodeFromString(parameterTreeValueHash(value)),
			},
		}
	}
	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.parameters": {Fields: parameters},
		},
	}
}

func treeStateSpecData(
	values map[string]string,
	secureValues map[string]string,
) *core.MappingNode {
	fields := map[string]*core.MappingNode{
		"path": core.MappingNodeFromString(testTreePath),
	}
	parameters := map[string]*core.MappingNode{}

	if len(values) > 0 {
		fields["values"] = core.MappingNodeFromStringMap(values)
		for key, value := range values {
			parameters[key] = &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":       core.MappingNodeFromString(treeParameterARN(key)),
					"type":      core.MappingNodeFromString("String"),
					"valueHash": core.MappingNodeFromString(parameterTreeValueHash(value)),
				},
			}
		}
	}

	if len(secureValues) > 0 {
		fields["secureValues"] = core.MappingNodeFromStringMap(secureValues)
		for key, value := range secureValues {
			parameters[key] = &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":       core.MappingNodeFromString(treeParameterARN(key)),
					"type":      core.MappingNodeFromString("SecureString"),
					"valueHash": core.MappingNodeFromString(parameterTreeValueHash(value)),
				},
			}
		}
	}

	fields["parameters"] = &core.MappingNode{Fields: parameters}
	return &core.MappingNode{Fields: fields}
}

func treeUpdateInput(
	providerCtx provider.Context,
	currentSpec *core.MappingNode,
	updatedSpec *core.MappingNode,
	modifiedFields []provider.FieldChange,
) *provider.ResourceDeployInput {
	return &provider.ResourceDeployInput{
		InstanceID: "test-instance-id",
		ResourceID: "test-resource-id",
		Changes: &provider.Changes{
			AppliedResourceInfo: provider.ResourceInfo{
				ResourceID:   "test-resource-id",
				ResourceName: "TestParameterTree",
				InstanceID:   "test-instance-id",
				CurrentResourceState: &state.ResourceState{
					ResourceID: "test-resource-id",
					Name:       "TestParameterTree",
					InstanceID: "test-instance-id",
					SpecData:   currentSpec,
				},
				ResourceWithResolvedSubs: &provider.ResolvedResource{
					Type: &schema.ResourceTypeWrapper{
						Value: "aws/ssm/parameterTree",
					},
					Spec: updatedSpec,
				},
			},
			ModifiedFields: modifiedFields,
		},
		ProviderContext: providerCtx,
	}
}

func putTagsContain(tags []ssmtypes.Tag, key, value string) bool {
	for _, tag := range tags {
		if aws.ToString(tag.Key) == key && aws.ToString(tag.Value) == value {
			return true
		}
	}
	return false
}

func TestSSMParameterTreeResourceUpdateSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterTreeResourceUpdateSuite))
}
