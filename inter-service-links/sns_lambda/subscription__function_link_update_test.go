//go:build unit

package snslambda

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

type SubscriptionFunctionLinkUpdateSuite struct {
	suite.Suite
}

const (
	sfFunctionARN = "arn:aws:lambda:us-west-2:123456789012:function:process-order"
	sfTopicARN    = "arn:aws:sns:us-west-2:123456789012:orders"
	sfInstanceID  = "instance-1"
	sfResourceID  = "ordersSubscription__processOrderFunction__sns-invoke-permission"
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

func sfSubscriptionInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersSubscription",
		InstanceID:   sfInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"topicArn", core.MappingNodeFromString(sfTopicARN),
				"endpoint", core.MappingNodeFromString(sfFunctionARN),
				"protocol", core.MappingNodeFromString("lambda"),
			),
		},
	}
}

func sfFunctionInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "processOrderFunction",
		InstanceID:   sfInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(sfFunctionARN),
			),
		},
	}
}

func (s *SubscriptionFunctionLinkUpdateSuite) Test_create_deploys_lambda_permission_intermediary() {
	mock := resourceservicemock.Create(resourceservicemock.WithDeployOutput(&provider.ResourceDeployOutput{}))

	actions := &subscriptionFunctionLinkActions{}
	out, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeCreate,
		ResourceAInfo:   sfSubscriptionInfo(),
		ResourceBInfo:   sfFunctionInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().NoError(err)

	s.Require().Len(mock.DeployCalls, 1)
	call := mock.DeployCalls[0]
	s.Equal("aws/lambda/permission", call.ResourceType)
	s.Equal(sfResourceID, call.Input.DeployInput.ResourceID)

	spec := call.Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal(sfFunctionARN, core.StringValue(spec.Fields["functionName"]))
	s.Equal("lambda:InvokeFunction", core.StringValue(spec.Fields["action"]))
	s.Equal("sns.amazonaws.com", core.StringValue(spec.Fields["principal"]))
	// The permission is scoped to the subscription's topic.
	s.Equal(sfTopicARN, core.StringValue(spec.Fields["sourceArn"]))

	s.Require().Len(out.IntermediaryResourceStates, 1)
	s.Equal(sfResourceID, out.IntermediaryResourceStates[0].ResourceID)
}

func (s *SubscriptionFunctionLinkUpdateSuite) Test_destroy_removes_lambda_permission_intermediary() {
	mock := resourceservicemock.Create()

	actions := &subscriptionFunctionLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeDestroy,
		ResourceAInfo:   sfSubscriptionInfo(),
		ResourceBInfo:   sfFunctionInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
		CurrentLinkState: &state.LinkState{
			IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{
				{ResourceID: sfResourceID, ResourceType: "aws/lambda/permission", InstanceID: sfInstanceID},
			},
		},
	})
	s.Require().NoError(err)

	s.Require().Len(mock.DestroyCalls, 1)
	s.Equal(sfResourceID, mock.DestroyCalls[0].Input.ResourceID)
	s.Empty(mock.DeployCalls)
}

func (s *SubscriptionFunctionLinkUpdateSuite) Test_create_errors_when_topic_arn_missing() {
	mock := resourceservicemock.Create()

	subscriptionInfo := &provider.ResourceInfo{
		ResourceName: "ordersSubscription",
		InstanceID:   sfInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"endpoint", core.MappingNodeFromString(sfFunctionARN),
			),
		},
	}

	actions := &subscriptionFunctionLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeCreate,
		ResourceAInfo:   subscriptionInfo,
		ResourceBInfo:   sfFunctionInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "topic ARN")
	s.Empty(mock.DeployCalls)
}

func (s *SubscriptionFunctionLinkUpdateSuite) Test_deploy_error_propagates() {
	mock := resourceservicemock.Create(resourceservicemock.WithDeployError(errors.New("boom")))

	actions := &subscriptionFunctionLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeCreate,
		ResourceAInfo:   sfSubscriptionInfo(),
		ResourceBInfo:   sfFunctionInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().Error(err)
}

func TestSubscriptionFunctionLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(SubscriptionFunctionLinkUpdateSuite))
}
