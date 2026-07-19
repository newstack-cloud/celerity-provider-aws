//go:build unit

package overlays

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type EventsRuleOverlaySuite struct {
	suite.Suite
}

func (s *EventsRuleOverlaySuite) Test_marks_target_arn_as_a_link_wiring_slot() {
	schema := &provider.ResourceDefinitionsSchema{
		Type:  provider.ResourceDefinitionsSchemaTypeObject,
		Label: "Rule",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"targets": {
				Type:  provider.ResourceDefinitionsSchemaTypeArray,
				Label: "Targets",
				Items: &provider.ResourceDefinitionsSchema{
					Type:  provider.ResourceDefinitionsSchemaTypeObject,
					Label: "Target",
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"arn": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "ARN"},
						"id":  {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "ID"},
					},
				},
			},
		},
	}

	out := Apply(eventsRuleType, schema)

	arn := out.Attributes["targets"].Items.Attributes["arn"]
	s.True(arn.ActivatesLinkOnReference)
	// Sibling fields are untouched.
	s.False(out.Attributes["targets"].Items.Attributes["id"].ActivatesLinkOnReference)
}

func (s *EventsRuleOverlaySuite) Test_is_a_no_op_when_targets_absent() {
	schema := &provider.ResourceDefinitionsSchema{
		Type:       provider.ResourceDefinitionsSchemaTypeObject,
		Label:      "Rule",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{},
	}

	s.NotPanics(func() { Apply(eventsRuleType, schema) })
}

// The Cloud Control AWS::Events::Rule handler when Name is omitted, so
// the engine must generate a rule name client-side for name-less rules.
func (s *EventsRuleOverlaySuite) Test_generates_a_rule_name_when_omitted() {
	behaviour := BehaviourFor(eventsRuleType)
	s.Require().NotNil(behaviour)
	s.Require().NotNil(behaviour.Name)
	s.Equal("name", behaviour.Name.Field)

	name, err := behaviour.Name.Generate(&provider.ResourceDeployInput{
		Changes: &provider.Changes{
			AppliedResourceInfo: provider.ResourceInfo{
				ResourceName: "orderEventsRule",
			},
		},
	})
	s.Require().NoError(err)
	s.NotEmpty(name)
	s.LessOrEqual(len(name), 64)
	s.Contains(name, "orderEventsRule")
}

func TestEventsRuleOverlaySuite(t *testing.T) {
	suite.Run(t, new(EventsRuleOverlaySuite))
}
