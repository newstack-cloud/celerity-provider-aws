//go:build unit

package eventslambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	eventsservice "github.com/newstack-cloud/bluelink-provider-aws/services/events/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type RuleFunctionLinkStageChangesSuite struct {
	suite.Suite
}

func ruleFunctionLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, eventsservice.Service, *aws.Config, lambdaservice.Service],
) provider.Link {
	return func(
		pluginutils.LinkServiceDeps[*aws.Config, eventsservice.Service, *aws.Config, lambdaservice.Service],
	) provider.Link {
		return RuleFunctionLink()
	}
}

const (
	ruleFunctionRuleARN     = "arn:aws:events:us-west-2:123456789012:rule/order-created"
	ruleFunctionFunctionARN = "arn:aws:lambda:us-west-2:123456789012:function:process-order"
	// The link-owned aws/lambda/permission intermediary id (rule__function__suffix).
	ruleFunctionPermissionID = "orderCreatedRule__processOrderFunction__eventbridge-invoke-permission"
)

func ruleFunctionLeaf(leaf string) string {
	return "[\"intermediaries\"][\"" + ruleFunctionPermissionID + "\"][\"" + leaf + "\"]"
}

func (s *RuleFunctionLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		eventsservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		{
			Name: "stages create of the invoke permission intermediary",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: ruleChangesWithARN("orderCreatedRule", ruleFunctionRuleARN),
				ResourceBChanges: functionChangesWithARN("processOrderFunction", ruleFunctionFunctionARN),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: ruleFunctionLeaf("resourceType"), NewValue: core.MappingNodeFromString("aws/lambda/permission")},
						{FieldPath: ruleFunctionLeaf("sourceArn"), NewValue: core.MappingNodeFromString(ruleFunctionRuleARN)},
						{FieldPath: ruleFunctionLeaf("functionName"), NewValue: core.MappingNodeFromString(ruleFunctionFunctionARN)},
					},
				},
			},
		},
		{
			Name: "stages known-on-deploy when rule and function are new",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: ruleChangesKnownOnDeploy("orderCreatedRule"),
				ResourceBChanges: functionChangesKnownOnDeploy("processOrderFunction"),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: ruleFunctionLeaf("resourceType"), NewValue: core.MappingNodeFromString("aws/lambda/permission")},
					},
					FieldChangesKnownOnDeploy: []string{
						ruleFunctionLeaf("sourceArn"),
						ruleFunctionLeaf("functionName"),
					},
				},
			},
		},
		{
			Name: "stages no changes when the intermediary is unchanged",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: ruleChangesWithARN("orderCreatedRule", ruleFunctionRuleARN),
				ResourceBChanges: functionChangesWithARN("processOrderFunction", ruleFunctionFunctionARN),
				CurrentLinkState: &state.LinkState{
					LinkID: "test-link",
					Data: map[string]*core.MappingNode{
						"intermediaries": {Fields: map[string]*core.MappingNode{
							ruleFunctionPermissionID: {Fields: map[string]*core.MappingNode{
								"resourceType": core.MappingNodeFromString("aws/lambda/permission"),
								"sourceArn":    core.MappingNodeFromString(ruleFunctionRuleARN),
								"functionName": core.MappingNodeFromString(ruleFunctionFunctionARN),
							}},
						}},
					},
				},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					UnchangedFields: []string{
						ruleFunctionLeaf("resourceType"),
						ruleFunctionLeaf("sourceArn"),
						ruleFunctionLeaf("functionName"),
					},
				},
			},
		},
	}

	plugintestutils.RunLinkChangeStagingTestCases(testCases, ruleFunctionLinkFactory(), &s.Suite)
}

func ruleChangesWithARN(resourceName, arn string) *provider.Changes {
	return &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceName: resourceName,
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString(arn),
				}},
			},
		},
	}
}

func functionChangesWithARN(resourceName, arn string) *provider.Changes {
	return ruleChangesWithARN(resourceName, arn)
}

func ruleChangesKnownOnDeploy(resourceName string) *provider.Changes {
	return &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceName: resourceName,
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: &core.MappingNode{Fields: map[string]*core.MappingNode{}},
			},
		},
		FieldChangesKnownOnDeploy: []string{"spec.arn"},
	}
}

func functionChangesKnownOnDeploy(resourceName string) *provider.Changes {
	return ruleChangesKnownOnDeploy(resourceName)
}

func TestRuleFunctionLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(RuleFunctionLinkStageChangesSuite))
}
