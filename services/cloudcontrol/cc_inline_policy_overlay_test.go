//go:build unit

package cloudcontrol

import (
	"testing"

	"github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/overlays"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type CCInlinePolicyOverlaySuite struct {
	suite.Suite
}

func freeFormPolicySchema(refField string) *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:  provider.ResourceDefinitionsSchemaTypeObject,
		Label: "Inline Policy",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			refField:         {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Ref"},
			"policyDocument": {Type: provider.ResourceDefinitionsSchemaTypeMap, Label: "Policy Document"},
		},
	}
}

func inlinePolicySpec(refField, refValue string) *core.MappingNode {
	return core.MappingNodeFields(
		refField, core.MappingNodeFromString(refValue),
		"policyDocument", core.MappingNodeFields(
			"version", core.MappingNodeFromString("2012-10-17"),
			"statement", core.MappingNodeItems(
				core.MappingNodeFields(
					"sid", core.MappingNodeFromString("S3Publish"),
					"effect", core.MappingNodeFromString("Allow"),
					"principal", core.MappingNodeFields(
						"service", core.MappingNodeFromString("s3.amazonaws.com"),
					),
					"action", core.MappingNodeFromString("sns:Publish"),
					"resource", core.MappingNodeFromString("arn:aws:sns:us-west-2:123456789012:uploads"),
				),
			),
		),
	)
}

// A policy document left as a free-form map goes to Cloud Control with its
// camelCase keys verbatim, which the AWS handler reads as a null policy
// ("Invalid parameter: Policy Error: null"). The overlay must re-type every
// inline-policy document so the engine translates it to the PascalCase shape.
func (s *CCInlinePolicyOverlaySuite) Test_inline_policy_documents_translate_to_pascal_case() {
	testCases := []struct {
		blueprintType string
		refField      string
	}{
		{blueprintType: "aws/sns/topicInlinePolicy", refField: "topicArn"},
		{blueprintType: "aws/sqs/queueInlinePolicy", refField: "queue"},
		{blueprintType: "aws/s3/bucketPolicy", refField: "bucket"},
	}

	for _, tc := range testCases {
		s.Run(tc.blueprintType, func() {
			schema := overlays.Apply(tc.blueprintType, freeFormPolicySchema(tc.refField))
			s.Require().Equal(
				provider.ResourceDefinitionsSchemaTypeObject,
				schema.Attributes["policyDocument"].Type,
				"the overlay must re-type the policy document to a structured object",
			)

			cfn := SpecToCFN(inlinePolicySpec(tc.refField, "test-ref"), schema, CCResourceMeta{})

			doc := cfn.Fields["PolicyDocument"]
			s.Require().NotNil(doc, "PolicyDocument should be PascalCase")
			s.Equal("2012-10-17", core.StringValue(doc.Fields["Version"]))
			stmt := doc.Fields["Statement"].Items[0].Fields
			s.Equal("Allow", core.StringValue(stmt["Effect"]))
			s.Equal("sns:Publish", core.StringValue(stmt["Action"]))
			s.NotNil(stmt["Principal"].Fields["Service"])
		})
	}
}

func TestCCInlinePolicyOverlaySuite(t *testing.T) {
	suite.Run(t, new(CCInlinePolicyOverlaySuite))
}
