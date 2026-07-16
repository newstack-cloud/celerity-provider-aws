//go:build unit

package elasticachesecretsmanager

import (
	"context"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type ReplicationGroupSecretLinkValidateSuite struct {
	suite.Suite
}

func (s *ReplicationGroupSecretLinkValidateSuite) validate(
	resourceASpec *core.MappingNode,
) *provider.LinkValidateOutput {
	actions := &replicationGroupSecretLinkActions{}
	output, err := actions.Validate(context.Background(), &provider.LinkValidateInput{
		ResourceAName: "sessionCache",
		ResourceBName: "redisAuthSecret",
		ResourceASpec: resourceASpec,
	})
	s.Require().NoError(err)
	s.Require().NotNil(output)
	return output
}

func (s *ReplicationGroupSecretLinkValidateSuite) Test_errors_when_transit_encryption_disabled() {
	output := s.validate(core.MappingNodeFields(
		"replicationGroupId", core.MappingNodeFromString("sessions"),
		"transitEncryptionEnabled", core.MappingNodeFromBool(false),
	))
	s.Require().Len(output.Diagnostics, 1)
	s.Equal(core.DiagnosticLevelError, output.Diagnostics[0].Level)
	s.Contains(output.Diagnostics[0].Message, "transitEncryptionEnabled")
}

func (s *ReplicationGroupSecretLinkValidateSuite) Test_errors_when_transit_encryption_absent() {
	output := s.validate(core.MappingNodeFields(
		"replicationGroupId", core.MappingNodeFromString("sessions"),
	))
	s.Require().Len(output.Diagnostics, 1)
	s.Equal(core.DiagnosticLevelError, output.Diagnostics[0].Level)
}

func (s *ReplicationGroupSecretLinkValidateSuite) Test_no_diagnostic_when_transit_encryption_enabled() {
	output := s.validate(core.MappingNodeFields(
		"replicationGroupId", core.MappingNodeFromString("sessions"),
		"transitEncryptionEnabled", core.MappingNodeFromBool(true),
	))
	s.Empty(output.Diagnostics)
}

func TestReplicationGroupSecretLinkValidateSuite(t *testing.T) {
	suite.Run(t, new(ReplicationGroupSecretLinkValidateSuite))
}
