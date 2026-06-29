//go:build unit

package snssqs

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

type SubscriptionQueueLinkUpdateSuite struct {
	suite.Suite
}

const (
	sqQueueARN   = "arn:aws:sqs:us-west-2:123456789012:orders-consumer"
	sqQueueURL   = "https://sqs.us-west-2.amazonaws.com/123456789012/orders-consumer"
	sqTopicARN   = "arn:aws:sns:us-west-2:123456789012:orders"
	sqInstanceID = "instance-1"
	sqResourceID = "ordersSubscription__ordersQueue__sns-send-policy"
	sqSID        = "SNSordersSubscription"
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

func sqSubscriptionInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersSubscription",
		InstanceID:   sqInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"topicArn", core.MappingNodeFromString(sqTopicARN),
				"endpoint", core.MappingNodeFromString(sqQueueARN),
				"protocol", core.MappingNodeFromString("sqs"),
			),
		},
	}
}

func sqQueueInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersQueue",
		InstanceID:   sqInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(sqQueueARN),
				"queueUrl", core.MappingNodeFromString(sqQueueURL),
			),
		},
	}
}

func (s *SubscriptionQueueLinkUpdateSuite) Test_create_deploys_queue_inline_policy_intermediary() {
	mock := resourceservicemock.Create(resourceservicemock.WithDeployOutput(&provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.__ccPrimaryIdentifier": core.MappingNodeFromString(sqQueueURL),
		},
	}))

	actions := &subscriptionQueueLinkActions{}
	out, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeCreate,
		ResourceAInfo:   sqSubscriptionInfo(),
		ResourceBInfo:   sqQueueInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().NoError(err)

	s.Require().Len(mock.DeployCalls, 1)
	call := mock.DeployCalls[0]
	s.Equal("aws/sqs/queueInlinePolicy", call.ResourceType)
	s.Equal(sqResourceID, call.Input.DeployInput.ResourceID)
	s.Nil(call.Input.DeployInput.Changes.AppliedResourceInfo.CurrentResourceState)

	spec := call.Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal(sqQueueURL, core.StringValue(spec.Fields["queue"]))

	policyDocument := spec.Fields["policyDocument"]
	s.Require().NotNil(policyDocument)
	statement := policyDocument.Fields["statement"].Items[0]
	s.Equal(sqSID, core.StringValue(statement.Fields["sid"]))
	s.Equal("sqs:SendMessage", core.StringValue(statement.Fields["action"]))
	s.Equal("sns.amazonaws.com", core.StringValue(statement.Fields["principal"].Fields["service"]))
	s.Equal(sqQueueARN, core.StringValue(statement.Fields["resource"]))
	// The statement is scoped to the subscription's topic, not the queue.
	s.Equal(sqTopicARN, core.StringValue(statement.Fields["condition"].Fields["ArnEquals"].Fields["aws:SourceArn"]))

	s.Require().Len(out.IntermediaryResourceStates, 1)
	s.Equal(sqResourceID, out.IntermediaryResourceStates[0].ResourceID)
}

func (s *SubscriptionQueueLinkUpdateSuite) Test_destroy_removes_queue_inline_policy_intermediary() {
	mock := resourceservicemock.Create()

	actions := &subscriptionQueueLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeDestroy,
		ResourceAInfo:   sqSubscriptionInfo(),
		ResourceBInfo:   sqQueueInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
		CurrentLinkState: &state.LinkState{
			IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{
				{ResourceID: sqResourceID, ResourceType: "aws/sqs/queueInlinePolicy", InstanceID: sqInstanceID},
			},
		},
	})
	s.Require().NoError(err)

	s.Require().Len(mock.DestroyCalls, 1)
	s.Equal(sqResourceID, mock.DestroyCalls[0].Input.ResourceID)
	s.Empty(mock.DeployCalls)
}

func (s *SubscriptionQueueLinkUpdateSuite) Test_create_errors_when_topic_arn_missing() {
	mock := resourceservicemock.Create()

	subscriptionInfo := &provider.ResourceInfo{
		ResourceName: "ordersSubscription",
		InstanceID:   sqInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"endpoint", core.MappingNodeFromString(sqQueueARN),
			),
		},
	}

	actions := &subscriptionQueueLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeCreate,
		ResourceAInfo:   subscriptionInfo,
		ResourceBInfo:   sqQueueInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "topic ARN")
	s.Empty(mock.DeployCalls)
}

func (s *SubscriptionQueueLinkUpdateSuite) Test_deploy_error_propagates() {
	mock := resourceservicemock.Create(resourceservicemock.WithDeployError(errors.New("boom")))

	actions := &subscriptionQueueLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeCreate,
		ResourceAInfo:   sqSubscriptionInfo(),
		ResourceBInfo:   sqQueueInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().Error(err)
}

func TestSubscriptionQueueLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionQueueLinkUpdateSuite))
}
