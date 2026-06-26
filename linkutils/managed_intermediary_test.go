//go:build unit

package linkutils

import (
	"context"
	"testing"

	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/stretchr/testify/suite"
)

type ManagedIntermediarySuite struct {
	suite.Suite
}

const (
	miType       = "aws/lambda/permission"
	miResourceID = "rule--lambda--permission"
	miInstance   = "instance-1"
)

func miIntermediary() ManagedIntermediary {
	return ManagedIntermediary{
		ResourceType: miType,
		ResourceID:   miResourceID,
		ResourceName: "ruleInvokeLambda",
		Spec: core.MappingNodeFields(
			"functionName", core.MappingNodeFromString("processOrders"),
			"action", core.MappingNodeFromString("lambda:InvokeFunction"),
			"principal", core.MappingNodeFromString("events.amazonaws.com"),
		),
	}
}

func miDeployOutput() *provider.ResourceDeployOutput {
	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.__ccPrimaryIdentifier": core.MappingNodeFromString("processOrders|abc123"),
			"spec.__ccRequestToken":      core.MappingNodeFromString("tok"),
		},
	}
}

func (s *ManagedIntermediarySuite) Test_deploy_create_has_no_current_state_and_records_identifier() {
	mock := resourceservicemock.Create(resourceservicemock.WithDeployOutput(miDeployOutput()))

	st, err := DeployManagedIntermediary(
		context.Background(), mock, miInstance, "instance", nil, nil, miIntermediary(),
	)
	s.Require().NoError(err)

	s.Require().Len(mock.DeployCalls, 1)
	call := mock.DeployCalls[0]
	s.Equal(miType, call.ResourceType)
	s.True(call.Input.WaitUntilStable)
	s.Equal(miResourceID, call.Input.DeployInput.ResourceID)
	// Create: no current resource state.
	s.Nil(call.Input.DeployInput.Changes.AppliedResourceInfo.CurrentResourceState)
	// Spec carried through.
	spec := call.Input.DeployInput.Changes.AppliedResourceInfo.ResourceWithResolvedSubs.Spec
	s.Equal("processOrders", core.StringValue(spec.Fields["functionName"]))

	// Returned state merges the computed identifier for later updates/destroys.
	s.Equal(miResourceID, st.ResourceID)
	s.Equal(miType, st.ResourceType)
	s.Equal("processOrders|abc123", core.StringValue(st.ResourceSpecData.Fields["__ccPrimaryIdentifier"]))
}

func (s *ManagedIntermediarySuite) Test_deploy_update_passes_current_state() {
	mock := resourceservicemock.Create(resourceservicemock.WithDeployOutput(miDeployOutput()))
	prior := &state.LinkIntermediaryResourceState{
		ResourceID:   miResourceID,
		ResourceType: miType,
		InstanceID:   miInstance,
		ResourceSpecData: core.MappingNodeFields(
			"__ccPrimaryIdentifier", core.MappingNodeFromString("processOrders|abc123"),
		),
	}

	_, err := DeployManagedIntermediary(
		context.Background(), mock, miInstance, "instance", nil, prior, miIntermediary(),
	)
	s.Require().NoError(err)

	current := mock.DeployCalls[0].Input.DeployInput.Changes.AppliedResourceInfo.CurrentResourceState
	s.Require().NotNil(current, "update must pass current state so the engine runs Update")
	s.Equal(miType, current.Type)
	s.Equal("processOrders|abc123", core.StringValue(current.SpecData.Fields["__ccPrimaryIdentifier"]))
}

func (s *ManagedIntermediarySuite) Test_destroy_uses_prior_state() {
	mock := resourceservicemock.Create()
	prior := &state.LinkIntermediaryResourceState{
		ResourceID:   miResourceID,
		ResourceType: miType,
		InstanceID:   miInstance,
	}

	err := DestroyManagedIntermediary(context.Background(), mock, miInstance, nil, prior)
	s.Require().NoError(err)

	s.Require().Len(mock.DestroyCalls, 1)
	s.Equal(miType, mock.DestroyCalls[0].ResourceType)
	s.Equal(miResourceID, mock.DestroyCalls[0].Input.ResourceID)
}

func (s *ManagedIntermediarySuite) Test_destroy_nil_prior_is_noop() {
	mock := resourceservicemock.Create()
	err := DestroyManagedIntermediary(context.Background(), mock, miInstance, nil, nil)
	s.Require().NoError(err)
	s.Empty(mock.DestroyCalls)
}

func (s *ManagedIntermediarySuite) Test_find_intermediary_state() {
	linkState := &state.LinkState{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{
			{ResourceID: "other"},
			{ResourceID: miResourceID, ResourceType: miType},
		},
	}
	found := FindIntermediaryState(linkState, miResourceID)
	s.Require().NotNil(found)
	s.Equal(miType, found.ResourceType)
	s.Nil(FindIntermediaryState(linkState, "missing"))
	s.Nil(FindIntermediaryState(nil, miResourceID))
}

func TestManagedIntermediarySuite(t *testing.T) {
	suite.Run(t, new(ManagedIntermediarySuite))
}
