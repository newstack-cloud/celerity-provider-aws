//go:build unit

package cloudcontrol

import (
	"testing"

	"github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/overlays"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type CCEventsConnectionOverlaySuite struct {
	suite.Suite
}

func baseConnectionSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:  provider.ResourceDefinitionsSchemaTypeObject,
		Label: "AWS Events Connection",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"name":           {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Name"},
			"authParameters": {Type: provider.ResourceDefinitionsSchemaTypeString, Nullable: true},
		},
	}
}

func (s *CCEventsConnectionOverlaySuite) Test_auth_parameters_retyped_as_open_map_and_passed_through_verbatim() {
	schema := overlays.Apply("aws/events/connection", baseConnectionSchema())

	// The overlay must have re-typed authParameters from a string to an open map so it
	// can be authored as a native object.
	authSchema := schema.Attributes["authParameters"]
	s.Require().Equal(provider.ResourceDefinitionsSchemaTypeMap, authSchema.Type)
	s.Nil(authSchema.MapValues, "open map carries no declared value schema")

	spec := &core.MappingNode{Fields: map[string]*core.MappingNode{
		"name": core.MappingNodeFromString("my-conn"),
		"authParameters": {Fields: map[string]*core.MappingNode{
			"ApiKeyAuthParameters": {Fields: map[string]*core.MappingNode{
				"ApiKeyName":  core.MappingNodeFromString("x-api-key"),
				"ApiKeyValue": core.MappingNodeFromString("secret"),
			}},
		}},
	}}

	cfn := SpecToCFN(spec, schema, CCResourceMeta{})

	// The top-level attribute key flips to PascalCase; the free-form nested keys are
	// passed through to Cloud Control verbatim (they are already CloudFormation-shaped).
	authParams := cfn.Fields["AuthParameters"]
	s.Require().NotNil(authParams)
	s.Require().NotNil(authParams.Fields, "AuthParameters must be an object, not a string")
	apiKey := authParams.Fields["ApiKeyAuthParameters"]
	s.Require().NotNil(apiKey, "nested keys preserved verbatim")
	s.Equal("x-api-key", core.StringValue(apiKey.Fields["ApiKeyName"]))
	s.Equal("secret", core.StringValue(apiKey.Fields["ApiKeyValue"]))
}

func TestCCEventsConnectionOverlaySuite(t *testing.T) {
	suite.Run(t, new(CCEventsConnectionOverlaySuite))
}
