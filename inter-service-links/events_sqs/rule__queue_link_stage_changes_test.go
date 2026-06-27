//go:build unit

package eventssqs

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	eventsservice "github.com/newstack-cloud/bluelink-provider-aws/services/events/service"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type RuleQueueLinkStageChangesSuite struct {
	suite.Suite
}

func ruleQueueLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, eventsservice.Service, *aws.Config, sqsservice.Service],
) provider.Link {
	return func(
		pluginutils.LinkServiceDeps[*aws.Config, eventsservice.Service, *aws.Config, sqsservice.Service],
	) provider.Link {
		return RuleQueueLink()
	}
}

const (
	ruleQueueRuleARN  = "arn:aws:events:us-west-2:123456789012:rule/order-created"
	ruleQueueQueueARN = "arn:aws:sqs:us-west-2:123456789012:order-queue"
	// The link-owned aws/sqs/queueInlinePolicy intermediary id (rule__queue__suffix).
	ruleQueuePolicyID = "orderCreatedRule__orderQueue__eventbridge-send-policy"
)

func ruleQueueLeaf(leaf string) string {
	return "[\"intermediaries\"][\"" + ruleQueuePolicyID + "\"][\"" + leaf + "\"]"
}

func rqResourceChangesWithARN(resourceName, arn string) *provider.Changes {
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

func rqResourceChangesKnownOnDeploy(resourceName string) *provider.Changes {
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

func (s *RuleQueueLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config,
		eventsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		{
			Name: "stages create of the queue inline policy intermediary",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: rqResourceChangesWithARN("orderCreatedRule", ruleQueueRuleARN),
				ResourceBChanges: rqResourceChangesWithARN("orderQueue", ruleQueueQueueARN),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{
							FieldPath: ruleQueueLeaf("resourceType"),
							NewValue:  core.MappingNodeFromString("aws/sqs/queueInlinePolicy"),
						},
						{
							FieldPath: ruleQueueLeaf("sourceArn"),
							NewValue:  core.MappingNodeFromString(ruleQueueRuleARN),
						},
						{
							FieldPath: ruleQueueLeaf("queueArn"),
							NewValue:  core.MappingNodeFromString(ruleQueueQueueARN),
						},
					},
				},
			},
		},
		{
			Name: "stages known-on-deploy when rule and queue are new",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: rqResourceChangesKnownOnDeploy("orderCreatedRule"),
				ResourceBChanges: rqResourceChangesKnownOnDeploy("orderQueue"),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{
							FieldPath: ruleQueueLeaf("resourceType"),
							NewValue:  core.MappingNodeFromString("aws/sqs/queueInlinePolicy"),
						},
					},
					FieldChangesKnownOnDeploy: []string{
						ruleQueueLeaf("sourceArn"),
						ruleQueueLeaf("queueArn"),
					},
				},
			},
		},
		{
			Name: "stages no changes when the intermediary is unchanged",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: rqResourceChangesWithARN("orderCreatedRule", ruleQueueRuleARN),
				ResourceBChanges: rqResourceChangesWithARN("orderQueue", ruleQueueQueueARN),
				CurrentLinkState: &state.LinkState{
					LinkID: "test-link",
					Data: map[string]*core.MappingNode{
						"intermediaries": {Fields: map[string]*core.MappingNode{
							ruleQueuePolicyID: {Fields: map[string]*core.MappingNode{
								"resourceType": core.MappingNodeFromString("aws/sqs/queueInlinePolicy"),
								"sourceArn":    core.MappingNodeFromString(ruleQueueRuleARN),
								"queueArn":     core.MappingNodeFromString(ruleQueueQueueARN),
							}},
						}},
					},
				},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					UnchangedFields: []string{
						ruleQueueLeaf("resourceType"),
						ruleQueueLeaf("sourceArn"),
						ruleQueueLeaf("queueArn"),
					},
				},
			},
		},
	}

	plugintestutils.RunLinkChangeStagingTestCases(testCases, ruleQueueLinkFactory(), &s.Suite)
}

func TestRuleQueueLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(RuleQueueLinkStageChangesSuite))
}
