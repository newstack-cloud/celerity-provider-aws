//go:build unit

package overlays

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type SNSSubscriptionOverlaySuite struct {
	suite.Suite
}

func (s *SNSSubscriptionOverlaySuite) Test_marks_endpoint_as_a_link_wiring_slot() {
	schema := &provider.ResourceDefinitionsSchema{
		Type:  provider.ResourceDefinitionsSchemaTypeObject,
		Label: "Subscription",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"endpoint": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Endpoint"},
			"topicArn": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Topic ARN"},
			"protocol": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Protocol"},
		},
	}

	out := Apply(snsSubscriptionType, schema)

	s.True(out.Attributes["endpoint"].ActivatesLinkOnReference)
	// Sibling fields are untouched.
	s.False(out.Attributes["topicArn"].ActivatesLinkOnReference)
	s.False(out.Attributes["protocol"].ActivatesLinkOnReference)
}

func (s *SNSSubscriptionOverlaySuite) Test_is_a_no_op_when_endpoint_absent() {
	schema := &provider.ResourceDefinitionsSchema{
		Type:       provider.ResourceDefinitionsSchemaTypeObject,
		Label:      "Subscription",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{},
	}

	s.NotPanics(func() { Apply(snsSubscriptionType, schema) })
}

func TestSNSSubscriptionOverlaySuite(t *testing.T) {
	suite.Run(t, new(SNSSubscriptionOverlaySuite))
}
