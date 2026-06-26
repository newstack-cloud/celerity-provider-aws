//go:build unit

package cloudcontrol

import (
	"testing"

	"github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/overlays"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type CCIAMRoleOverlaySuite struct {
	suite.Suite
}

func baseRoleSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:  provider.ResourceDefinitionsSchemaTypeObject,
		Label: "Role",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"assumeRolePolicyDocument": {Type: provider.ResourceDefinitionsSchemaTypeMap, Label: "Assume Role Policy Document"},
			"policies": {
				Type:  provider.ResourceDefinitionsSchemaTypeArray,
				Label: "Policies",
				Items: &provider.ResourceDefinitionsSchema{
					Type:  provider.ResourceDefinitionsSchemaTypeObject,
					Label: "Policy",
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"policyName":     {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Policy Name"},
						"policyDocument": {Type: provider.ResourceDefinitionsSchemaTypeMap, Label: "Policy Document"},
					},
				},
			},
		},
	}
}

// Builds the effective Meta after the overlay's adjustment. Policy
// documents are free-form maps (not JSON-string fields) in the generated schema; the
// overlay adds the principal.aws -> AWS name override.
func roleMeta() CCResourceMeta {
	meta := CCResourceMeta{}
	if adjustment := overlays.MetaAdjustmentFor("aws/iam/role"); adjustment != nil {
		meta.JSONStringFields = removeStrings(meta.JSONStringFields, adjustment.RemoveJSONStringFields)
		meta.FieldNameOverrides = mergeStringMaps(meta.FieldNameOverrides, adjustment.AddFieldNameOverrides)
	}
	return meta
}

func trustPolicySpec() *core.MappingNode {
	return &core.MappingNode{Fields: map[string]*core.MappingNode{
		"assumeRolePolicyDocument": {Fields: map[string]*core.MappingNode{
			"version": core.MappingNodeFromString("2012-10-17"),
			"statement": {Items: []*core.MappingNode{
				{Fields: map[string]*core.MappingNode{
					"sid":    core.MappingNodeFromString("Trust"),
					"effect": core.MappingNodeFromString("Allow"),
					"action": core.MappingNodeFromString("sts:AssumeRole"),
					"principal": {Fields: map[string]*core.MappingNode{
						"service": core.MappingNodeFromString("lambda.amazonaws.com"),
						"aws":     {Items: []*core.MappingNode{core.MappingNodeFromString("arn:aws:iam::123456789012:user/x")}},
					}},
					"condition": {Fields: map[string]*core.MappingNode{
						"StringEquals": {Fields: map[string]*core.MappingNode{
							"aws:username": core.MappingNodeFromString("bob"),
						}},
					}},
				}},
			}},
		}},
	}}
}

func (s *CCIAMRoleOverlaySuite) Test_typed_policy_document_round_trips_through_cloud_control() {
	schema := overlays.Apply("aws/iam/role", baseRoleSchema())
	meta := roleMeta()

	// The overlay must have re-typed the policy document to an object.
	doc := schema.Attributes["assumeRolePolicyDocument"]
	s.Require().Equal(provider.ResourceDefinitionsSchemaTypeObject, doc.Type)

	cfn := SpecToCFN(trustPolicySpec(), schema, meta)

	stmt := cfn.Fields["AssumeRolePolicyDocument"].Fields["Statement"].Items[0].Fields
	s.Require().NotNil(stmt["Sid"], "Sid should be PascalCase")
	s.Equal("Allow", core.StringValue(stmt["Effect"]))
	s.Equal("sts:AssumeRole", core.StringValue(stmt["Action"]))

	principal := stmt["Principal"].Fields
	s.Require().NotNil(principal["Service"], "Service key preserved")
	s.Require().NotNil(principal["AWS"], "aws must override to AWS")

	// Condition operator and key names are arbitrary and must be preserved.
	condition := stmt["Condition"].Fields["StringEquals"].Fields
	s.Require().NotNil(condition["aws:username"], "condition keys preserved verbatim")
	s.Equal("bob", core.StringValue(condition["aws:username"]))

	// Round-trip back to the camelCase spec form.
	back := CFNToSpec(cfn, schema, meta)
	backStmt := back.Fields["assumeRolePolicyDocument"].Fields["statement"].Items[0].Fields
	s.Equal("Trust", core.StringValue(backStmt["sid"]))
	s.Require().NotNil(backStmt["principal"].Fields["aws"], "AWS must round-trip back to aws")
	s.Require().NotNil(backStmt["condition"].Fields["StringEquals"].Fields["aws:username"])
}

func TestCCIAMRoleOverlaySuite(t *testing.T) {
	suite.Run(t, new(CCIAMRoleOverlaySuite))
}
