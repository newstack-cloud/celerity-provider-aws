//go:build unit

package eventssqs

import (
	"context"
	"errors"
	"testing"

	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type RuleQueueLinkUpdateSuite struct {
	suite.Suite
}

const (
	rqlQueueARN   = "arn:aws:sqs:us-west-2:123456789012:order-queue"
	rqlQueueURL   = "https://sqs.us-west-2.amazonaws.com/123456789012/order-queue"
	rqlRuleARN    = "arn:aws:events:us-west-2:123456789012:rule/order-created-rule"
	rqlInstanceID = "instance-1"
	rqlResourceID = "orderCreatedRule__orderQueue__eventbridge-send-policy"
	rqlSID        = "EventBridgeorderCreatedRule"
)

func testLinkContext() provider.LinkContext {
	return plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {"region": core.ScalarFromString("us-west-2")},
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)
}

func rqlRuleInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "orderCreatedRule",
		InstanceID:   rqlInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(rqlRuleARN),
			),
		},
	}
}

func rqlQueueInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "orderQueue",
		InstanceID:   rqlInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(rqlQueueARN),
				"queueUrl", core.MappingNodeFromString(rqlQueueURL),
			),
		},
	}
}

func (s *RuleQueueLinkUpdateSuite) Test_create_deploys_queue_inline_policy_intermediary() {
	mock := resourceservicemock.Create(resourceservicemock.WithDeployOutput(&provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.__ccPrimaryIdentifier": core.MappingNodeFromString(rqlQueueURL),
		},
	}))

	actions := &ruleQueueLinkActions{}
	out, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeCreate,
		ResourceAInfo:   rqlRuleInfo(),
		ResourceBInfo:   rqlQueueInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().NoError(err)

	s.Require().Len(mock.DeployCalls, 1)
	call := mock.DeployCalls[0]
	s.Equal("aws/sqs/queueInlinePolicy", call.ResourceType)
	s.Equal(rqlResourceID, call.Input.DeployInput.ResourceID)
	// Create: no current resource state.
	s.Nil(call.Input.DeployInput.Changes.AppliedResourceInfo.CurrentResourceState)

	spec := call.Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	// The QueueInlinePolicy primary identifier ("queue") is the queue URL.
	s.Equal(rqlQueueURL, core.StringValue(spec.Fields["queue"]))

	// policyDocument is a structured object authored with camelCase fields (the
	// engine translates them to the PascalCase shape Cloud Control expects); the
	// condition block is a free-form map whose keys are kept verbatim.
	policyDocument := spec.Fields["policyDocument"]
	s.Require().NotNil(policyDocument)
	s.Equal("2012-10-17", core.StringValue(policyDocument.Fields["version"]))
	statement := policyDocument.Fields["statement"].Items[0]
	s.Equal(rqlSID, core.StringValue(statement.Fields["sid"]))
	s.Equal("sqs:SendMessage", core.StringValue(statement.Fields["action"]))
	s.Equal("events.amazonaws.com", core.StringValue(statement.Fields["principal"].Fields["service"]))
	s.Equal(rqlQueueARN, core.StringValue(statement.Fields["resource"]))
	s.Equal(rqlRuleARN, core.StringValue(statement.Fields["condition"].Fields["ArnEquals"].Fields["aws:SourceArn"]))

	// The intermediary state is returned so the framework persists the handle.
	s.Require().Len(out.IntermediaryResourceStates, 1)
	s.Equal(rqlResourceID, out.IntermediaryResourceStates[0].ResourceID)
	s.Equal("aws/sqs/queueInlinePolicy", out.IntermediaryResourceStates[0].ResourceType)
}

func (s *RuleQueueLinkUpdateSuite) Test_update_passes_prior_intermediary_state() {
	mock := resourceservicemock.Create(resourceservicemock.WithDeployOutput(&provider.ResourceDeployOutput{}))

	actions := &ruleQueueLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeUpdate,
		ResourceAInfo:   rqlRuleInfo(),
		ResourceBInfo:   rqlQueueInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
		CurrentLinkState: &state.LinkState{
			IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{
				{
					ResourceID:   rqlResourceID,
					ResourceType: "aws/sqs/queueInlinePolicy",
					InstanceID:   rqlInstanceID,
					ResourceSpecData: core.MappingNodeFields(
						"__ccPrimaryIdentifier", core.MappingNodeFromString(rqlQueueURL),
					),
				},
			},
		},
	})
	s.Require().NoError(err)

	s.Require().Len(mock.DeployCalls, 1)
	current := mock.DeployCalls[0].Input.DeployInput.Changes.AppliedResourceInfo.CurrentResourceState
	s.Require().NotNil(current, "update must pass current state so the engine runs Update")
	s.Equal("aws/sqs/queueInlinePolicy", current.Type)
}

func (s *RuleQueueLinkUpdateSuite) Test_destroy_removes_queue_inline_policy_intermediary() {
	mock := resourceservicemock.Create()

	actions := &ruleQueueLinkActions{}
	out, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeDestroy,
		ResourceAInfo:   rqlRuleInfo(),
		ResourceBInfo:   rqlQueueInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
		CurrentLinkState: &state.LinkState{
			IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{
				{
					ResourceID:   rqlResourceID,
					ResourceType: "aws/sqs/queueInlinePolicy",
					InstanceID:   rqlInstanceID,
				},
			},
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(out)

	s.Require().Len(mock.DestroyCalls, 1)
	s.Equal("aws/sqs/queueInlinePolicy", mock.DestroyCalls[0].ResourceType)
	s.Equal(rqlResourceID, mock.DestroyCalls[0].Input.ResourceID)
	s.Empty(mock.DeployCalls)
}

func (s *RuleQueueLinkUpdateSuite) Test_destroy_with_no_prior_state_is_noop() {
	mock := resourceservicemock.Create()

	actions := &ruleQueueLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeDestroy,
		ResourceAInfo:   rqlRuleInfo(),
		ResourceBInfo:   rqlQueueInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().NoError(err)
	s.Empty(mock.DestroyCalls)
}

func (s *RuleQueueLinkUpdateSuite) Test_create_errors_when_queue_arn_missing() {
	mock := resourceservicemock.Create()

	queueInfo := &provider.ResourceInfo{
		ResourceName: "orderQueue",
		InstanceID:   rqlInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"queueUrl", core.MappingNodeFromString(rqlQueueURL),
			),
		},
	}

	actions := &ruleQueueLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeCreate,
		ResourceAInfo:   rqlRuleInfo(),
		ResourceBInfo:   queueInfo,
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().Error(err)
	s.Empty(mock.DeployCalls)
}

func (s *RuleQueueLinkUpdateSuite) Test_create_errors_when_queue_url_missing() {
	mock := resourceservicemock.Create()

	queueInfo := &provider.ResourceInfo{
		ResourceName: "orderQueue",
		InstanceID:   rqlInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(rqlQueueARN),
			),
		},
	}

	actions := &ruleQueueLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeCreate,
		ResourceAInfo:   rqlRuleInfo(),
		ResourceBInfo:   queueInfo,
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().Error(err)
	s.Empty(mock.DeployCalls)
}

func (s *RuleQueueLinkUpdateSuite) Test_deploy_error_propagates() {
	mock := resourceservicemock.Create(resourceservicemock.WithDeployError(errors.New("boom")))

	actions := &ruleQueueLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeCreate,
		ResourceAInfo:   rqlRuleInfo(),
		ResourceBInfo:   rqlQueueInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().Error(err)
}

func TestRuleQueueLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(RuleQueueLinkUpdateSuite))
}
