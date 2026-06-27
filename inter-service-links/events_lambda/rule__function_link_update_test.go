//go:build unit

package eventslambda

import (
	"context"
	"errors"
	"testing"

	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/stretchr/testify/suite"
)

type RuleFunctionLinkUpdateSuite struct {
	suite.Suite
}

const (
	rflFunctionARN = "arn:aws:lambda:us-west-2:123456789012:function:process-order"
	rflRuleARN     = "arn:aws:events:us-west-2:123456789012:rule/order-created-rule"
	rflInstanceID  = "instance-1"
	rflResourceID  = "orderCreatedRule__processOrderFunction__eventbridge-invoke-permission"
)

func rflRuleInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "orderCreatedRule",
		InstanceID:   rflInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(rflRuleARN),
			),
		},
	}
}

func rflFunctionInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "processOrderFunction",
		InstanceID:   rflInstanceID,
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(rflFunctionARN),
			),
		},
	}
}

func (s *RuleFunctionLinkUpdateSuite) Test_create_deploys_invoke_permission_intermediary() {
	mock := resourceservicemock.Create(resourceservicemock.WithDeployOutput(&provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.__ccPrimaryIdentifier": core.MappingNodeFromString(rflFunctionARN + "|abc123"),
		},
	}))

	actions := &ruleFunctionLinkActions{}
	out, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeCreate,
		ResourceAInfo:   rflRuleInfo(),
		ResourceBInfo:   rflFunctionInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().NoError(err)

	s.Require().Len(mock.DeployCalls, 1)
	call := mock.DeployCalls[0]
	s.Equal("aws/lambda/permission", call.ResourceType)
	s.Equal(rflResourceID, call.Input.DeployInput.ResourceID)
	// Create: no current resource state.
	s.Nil(call.Input.DeployInput.Changes.AppliedResourceInfo.CurrentResourceState)

	spec := call.Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal(rflFunctionARN, core.StringValue(spec.Fields["functionName"]))
	s.Equal("lambda:InvokeFunction", core.StringValue(spec.Fields["action"]))
	s.Equal("events.amazonaws.com", core.StringValue(spec.Fields["principal"]))
	s.Equal(rflRuleARN, core.StringValue(spec.Fields["sourceArn"]))

	// The intermediary state is returned so the framework persists the handle.
	s.Require().Len(out.IntermediaryResourceStates, 1)
	s.Equal(rflResourceID, out.IntermediaryResourceStates[0].ResourceID)
	s.Equal("aws/lambda/permission", out.IntermediaryResourceStates[0].ResourceType)

	// The intermediary is projected into link data (keyed by its resource id) so that
	// what is persisted matches what StageChanges diffs.
	entry := out.LinkData.Fields["intermediaries"].Fields[rflResourceID]
	s.Require().NotNil(entry)
	s.Equal("aws/lambda/permission", core.StringValue(entry.Fields["resourceType"]))
	s.Equal(rflRuleARN, core.StringValue(entry.Fields["sourceArn"]))
	s.Equal(rflFunctionARN, core.StringValue(entry.Fields["functionName"]))
}

func (s *RuleFunctionLinkUpdateSuite) Test_update_passes_prior_intermediary_state() {
	mock := resourceservicemock.Create(resourceservicemock.WithDeployOutput(&provider.ResourceDeployOutput{}))

	actions := &ruleFunctionLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeUpdate,
		ResourceAInfo:   rflRuleInfo(),
		ResourceBInfo:   rflFunctionInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
		CurrentLinkState: &state.LinkState{
			IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{
				{
					ResourceID:   rflResourceID,
					ResourceType: "aws/lambda/permission",
					InstanceID:   rflInstanceID,
					ResourceSpecData: core.MappingNodeFields(
						"__ccPrimaryIdentifier", core.MappingNodeFromString(rflFunctionARN+"|abc123"),
					),
				},
			},
		},
	})
	s.Require().NoError(err)

	s.Require().Len(mock.DeployCalls, 1)
	current := mock.DeployCalls[0].Input.DeployInput.Changes.AppliedResourceInfo.CurrentResourceState
	s.Require().NotNil(current, "update must pass current state so the engine runs Update")
	s.Equal("aws/lambda/permission", current.Type)
}

func (s *RuleFunctionLinkUpdateSuite) Test_destroy_removes_invoke_permission_intermediary() {
	mock := resourceservicemock.Create()

	actions := &ruleFunctionLinkActions{}
	out, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeDestroy,
		ResourceAInfo:   rflRuleInfo(),
		ResourceBInfo:   rflFunctionInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
		CurrentLinkState: &state.LinkState{
			IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{
				{
					ResourceID:   rflResourceID,
					ResourceType: "aws/lambda/permission",
					InstanceID:   rflInstanceID,
				},
			},
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(out)

	s.Require().Len(mock.DestroyCalls, 1)
	s.Equal("aws/lambda/permission", mock.DestroyCalls[0].ResourceType)
	s.Equal(rflResourceID, mock.DestroyCalls[0].Input.ResourceID)
	s.Empty(mock.DeployCalls)
}

func (s *RuleFunctionLinkUpdateSuite) Test_destroy_with_no_prior_state_is_noop() {
	mock := resourceservicemock.Create()

	actions := &ruleFunctionLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeDestroy,
		ResourceAInfo:   rflRuleInfo(),
		ResourceBInfo:   rflFunctionInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().NoError(err)
	s.Empty(mock.DestroyCalls)
}

func (s *RuleFunctionLinkUpdateSuite) Test_create_errors_when_function_arn_missing() {
	mock := resourceservicemock.Create()

	ruleInfo := rflRuleInfo()
	functionInfo := &provider.ResourceInfo{
		ResourceName:         "processOrderFunction",
		InstanceID:           rflInstanceID,
		CurrentResourceState: &state.ResourceState{SpecData: core.MappingNodeFields()},
	}

	actions := &ruleFunctionLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeCreate,
		ResourceAInfo:   ruleInfo,
		ResourceBInfo:   functionInfo,
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().Error(err)
	s.Empty(mock.DeployCalls)
}

func (s *RuleFunctionLinkUpdateSuite) Test_deploy_error_propagates() {
	mock := resourceservicemock.Create(resourceservicemock.WithDeployError(errors.New("boom")))

	actions := &ruleFunctionLinkActions{}
	_, err := actions.UpdateIntermediaryResources(context.Background(), &provider.LinkUpdateIntermediaryResourcesInput{
		LinkUpdateType:  provider.LinkUpdateTypeCreate,
		ResourceAInfo:   rflRuleInfo(),
		ResourceBInfo:   rflFunctionInfo(),
		ResourceService: mock,
		LinkContext:     testLinkContext(),
		InstanceName:    "instance",
	})
	s.Require().Error(err)
}

func TestRuleFunctionLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(RuleFunctionLinkUpdateSuite))
}
