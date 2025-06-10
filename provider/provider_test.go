package provider

import (
	"context"
	"testing"

	"github.com/newstack-cloud/celerity-provider-aws/services/lambda"
	"github.com/newstack-cloud/celerity-provider-aws/utils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/stretchr/testify/suite"
)

type ProviderSuite struct {
	suite.Suite
}

func (s *ProviderSuite) Test_loads_provider_and_applies_duration_validation() {
	tests := []struct {
		name        string
		duration    string
		expectError bool
	}{
		{
			name:        "valid duration - 1 hour",
			duration:    "1h",
			expectError: false,
		},
		{
			name:        "valid duration - 15 minutes",
			duration:    "15m",
			expectError: false,
		},
		{
			name:        "valid duration - 12 hours",
			duration:    "12h",
			expectError: false,
		},
		{
			name:        "invalid duration - too short",
			duration:    "14m",
			expectError: true,
		},
		{
			name:        "invalid duration - too long",
			duration:    "13h",
			expectError: true,
		},
		{
			name:        "invalid duration - invalid format",
			duration:    "invalid",
			expectError: true,
		},
	}

	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		&utils.DefaultAWSConfigLoader{},
	)
	provider := NewProvider(lambda.NewService, configStore)
	configDef, err := provider.ConfigDefinition(context.Background())
	s.Require().NoError(err, "should get config definition without error")

	durationField := configDef.Fields["assumeRole.duration"]
	s.Require().NotNil(durationField, "assumeRole.duration field should exist in provider config")
	s.Require().NotNil(durationField.ValidateFunc, "assumeRole.duration field should have a validation function")

	for _, tt := range tests {
		s.Run(tt.name, func() {
			diagnostics := durationField.ValidateFunc(
				"assumeRole.duration",
				core.ScalarFromString(tt.duration),
				nil,
			)

			if tt.expectError {
				s.NotEmpty(diagnostics, "expected validation error for duration %s", tt.duration)
			} else {
				s.Empty(diagnostics, "unexpected validation error for duration %s", tt.duration)
			}
		})
	}
}

func (s *ProviderSuite) Test_loads_provider_and_applies_role_arn_validation() {
	tests := []struct {
		name        string
		roleArn     string
		expectError bool
	}{
		{
			name:        "valid role ARN",
			roleArn:     "arn:aws:iam::123456789012:role/test-role",
			expectError: false,
		},
		{
			name:        "valid role ARN with path",
			roleArn:     "arn:aws:iam::123456789012:role/path/to/test-role",
			expectError: false,
		},
		{
			name:        "valid role ARN with special characters",
			roleArn:     "arn:aws:iam::123456789012:role/test-role@=,.+-_",
			expectError: false,
		},
		{
			name:        "invalid ARN - missing partition",
			roleArn:     "arn::iam::123456789012:role/test-role",
			expectError: true,
		},
		{
			name:        "invalid ARN - invalid partition",
			roleArn:     "arn:invalid:iam::123456789012:role/test-role",
			expectError: true,
		},
		{
			name:        "invalid ARN - invalid region",
			roleArn:     "arn:aws:iam:invalid-region:123456789012:role/test-role",
			expectError: true,
		},
		{
			name:        "invalid ARN - invalid account ID",
			roleArn:     "arn:aws:iam::invalid:role/test-role",
			expectError: true,
		},
		{
			name:        "invalid ARN - missing resource",
			roleArn:     "arn:aws:iam::123456789012:",
			expectError: true,
		},
	}

	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		&utils.DefaultAWSConfigLoader{},
	)
	provider := NewProvider(lambda.NewService, configStore)
	configDef, err := provider.ConfigDefinition(context.Background())
	s.Require().NoError(err, "should get config definition without error")

	roleArnField := configDef.Fields["assumeRole.roleArn"]
	s.Require().NotNil(roleArnField, "assumeRole.roleArn field should exist in provider config")
	s.Require().NotNil(roleArnField.ValidateFunc, "assumeRole.roleArn field should have a validation function")

	for _, tt := range tests {
		s.Run(tt.name, func() {
			diagnostics := roleArnField.ValidateFunc(
				"assumeRole.roleArn",
				core.ScalarFromString(tt.roleArn),
				nil,
			)

			if tt.expectError {
				s.NotEmpty(diagnostics, "expected validation error for role ARN %s", tt.roleArn)
			} else {
				s.Empty(diagnostics, "unexpected validation error for role ARN %s", tt.roleArn)
			}
		})
	}
}

func TestProviderSuite(t *testing.T) {
	suite.Run(t, new(ProviderSuite))
}
