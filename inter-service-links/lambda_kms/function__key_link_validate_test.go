//go:build unit

package lambdakms

import (
	"context"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type FunctionKeyLinkValidateSuite struct {
	suite.Suite
}

func delegatingKeyPolicy() *core.MappingNode {
	return core.MappingNodeFields(
		"keyPolicy", core.MappingNodeFields(
			"Version", core.MappingNodeFromString("2012-10-17"),
			"Statement", core.MappingNodeItems(
				core.MappingNodeFields(
					"Sid", core.MappingNodeFromString("Enable IAM User Permissions"),
					"Effect", core.MappingNodeFromString("Allow"),
					"Principal", core.MappingNodeFields(
						"AWS", core.MappingNodeFromString("arn:aws:iam::123456789012:root"),
					),
					"Action", core.MappingNodeFromString("kms:*"),
					"Resource", core.MappingNodeFromString("*"),
				),
			),
		),
	)
}

func lockedDownKeyPolicy() *core.MappingNode {
	return core.MappingNodeFields(
		"keyPolicy", core.MappingNodeFields(
			"Version", core.MappingNodeFromString("2012-10-17"),
			"Statement", core.MappingNodeItems(
				core.MappingNodeFields(
					"Sid", core.MappingNodeFromString("AdminOnly"),
					"Effect", core.MappingNodeFromString("Allow"),
					"Principal", core.MappingNodeFields(
						"AWS", core.MappingNodeFromString("arn:aws:iam::123456789012:role/key-admin"),
					),
					"Action", core.MappingNodeItems(
						core.MappingNodeFromString("kms:Create*"),
						core.MappingNodeFromString("kms:Describe*"),
					),
					"Resource", core.MappingNodeFromString("*"),
				),
			),
		),
	)
}

func (s *FunctionKeyLinkValidateSuite) validate(
	resourceBSpec *core.MappingNode,
	annotations map[string]*core.ScalarValue,
) *provider.LinkValidateOutput {
	actions := &functionKeyLinkActions{}
	output, err := actions.Validate(context.Background(), &provider.LinkValidateInput{
		ResourceAName: "encryptFunction",
		ResourceBName: "dataKey",
		ResourceBSpec: resourceBSpec,
		Annotations:   annotations,
	})
	s.Require().NoError(err)
	s.Require().NotNil(output)
	return output
}

func (s *FunctionKeyLinkValidateSuite) Test_warns_when_key_policy_lacks_iam_delegation() {
	output := s.validate(lockedDownKeyPolicy(), nil)
	s.Require().Len(output.Diagnostics, 1)
	s.Equal(core.DiagnosticLevelWarning, output.Diagnostics[0].Level)
	s.Contains(output.Diagnostics[0].Message, "manageKeyGrant")
}

func (s *FunctionKeyLinkValidateSuite) Test_no_warning_when_key_policy_delegates_to_iam() {
	output := s.validate(delegatingKeyPolicy(), nil)
	s.Empty(output.Diagnostics)
}

func (s *FunctionKeyLinkValidateSuite) Test_no_warning_when_no_custom_key_policy() {
	output := s.validate(&core.MappingNode{Fields: map[string]*core.MappingNode{}}, nil)
	s.Empty(output.Diagnostics)
}

func (s *FunctionKeyLinkValidateSuite) Test_no_warning_when_manage_key_grant_enabled() {
	annotations := map[string]*core.ScalarValue{
		"aws/lambda/function::aws.lambda.kms.dataKey.manageKeyGrant": core.ScalarFromBool(true),
	}
	output := s.validate(lockedDownKeyPolicy(), annotations)
	s.Empty(output.Diagnostics)
}

func TestFunctionKeyLinkValidateSuite(t *testing.T) {
	suite.Run(t, new(FunctionKeyLinkValidateSuite))
}
