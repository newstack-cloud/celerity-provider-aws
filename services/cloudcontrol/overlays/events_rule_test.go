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

func TestEventsRuleOverlaySuite(t *testing.T) {
	suite.Run(t, new(EventsRuleOverlaySuite))
}
