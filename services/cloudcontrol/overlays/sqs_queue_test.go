//go:build unit

package overlays

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type SQSQueueOverlaySuite struct {
	suite.Suite
}

func (s *SQSQueueOverlaySuite) Test_leaves_redrive_as_json_string_and_sets_numeric_bounds() {
	schema := &provider.ResourceDefinitionsSchema{
		Type:  provider.ResourceDefinitionsSchemaTypeObject,
		Label: "Queue",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"redrivePolicy":      {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Redrive Policy"},
			"redriveAllowPolicy": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Redrive Allow Policy"},
			"delaySeconds":       {Type: provider.ResourceDefinitionsSchemaTypeInteger, Label: "Delay Seconds"},
		},
	}

	out := Apply(sqsQueueType, schema)

	// The redrive policies remain JSON-string (free-form) fields, the engine
	// parses/serialises them at the Cloud Control boundary and the DLQ link writes
	// the whole value, so no typed-object overlay is applied.
	s.Equal(provider.ResourceDefinitionsSchemaTypeString, out.Attributes["redrivePolicy"].Type)
	s.Nil(out.Attributes["redrivePolicy"].Attributes)
	s.Equal(provider.ResourceDefinitionsSchemaTypeString, out.Attributes["redriveAllowPolicy"].Type)

	// Numeric bounds (documented only in prose by CloudFormation) are restored.
	delay := out.Attributes["delaySeconds"]
	s.Require().NotNil(delay.Minimum)
	s.Require().NotNil(delay.Maximum)
	s.Equal(0, core.IntValueFromScalar(delay.Minimum))
	s.Equal(900, core.IntValueFromScalar(delay.Maximum))
}

func TestSQSQueueOverlaySuite(t *testing.T) {
	suite.Run(t, new(SQSQueueOverlaySuite))
}
