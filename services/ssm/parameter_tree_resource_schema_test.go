//go:build unit

package ssm

import (
	"strings"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/stretchr/testify/suite"
)

type SSMParameterTreeResourceSchemaSuite struct {
	suite.Suite
}

func (s *SSMParameterTreeResourceSchemaSuite) Test_value_fields_follow_blob_drift_posture() {
	treeSchema := parameterTreeResourceSchema()

	values := treeSchema.Attributes["values"]
	s.True(values.IgnoreDrift)
	s.False(values.Sensitive)

	secureValues := treeSchema.Attributes["secureValues"]
	s.True(secureValues.IgnoreDrift)
	s.True(secureValues.Sensitive)

	s.True(treeSchema.Attributes["path"].MustRecreate)
	s.True(treeSchema.Attributes["region"].MustRecreate)
	s.True(treeSchema.Attributes["parameters"].Computed)
}

func (s *SSMParameterTreeResourceSchemaSuite) Test_validate_accepts_valid_tree() {
	diagnostics := validateTreePathForSpec(
		"/my-app/config",
		map[string]string{"logLevel": "info", "db/host": "db.internal.example.com"},
		map[string]string{"apiToken": "super-secret"},
	)
	s.Empty(diagnostics)
}

func (s *SSMParameterTreeResourceSchemaSuite) Test_validate_rejects_invalid_keys() {
	invalidKeys := []string{"/leading", "trailing/", "empty//segment", "bad*char"}
	for _, key := range invalidKeys {
		diagnostics := validateTreePathForSpec(
			"/my-app/config",
			map[string]string{key: "value"},
			nil,
		)
		s.Len(diagnostics, 1, "expected key %q to be rejected", key)
		s.Contains(diagnostics[0].Message, "segments")
	}
}

func (s *SSMParameterTreeResourceSchemaSuite) Test_validate_rejects_key_exceeding_depth_limit() {
	// The path contributes 2 levels, so a 14-segment key exceeds the 15-level maximum.
	deepKey := strings.Repeat("a/", 13) + "a"
	diagnostics := validateTreePathForSpec(
		"/my-app/config",
		map[string]string{deepKey: "value"},
		nil,
	)
	s.Len(diagnostics, 1)
	s.Contains(diagnostics[0].Message, "nests too deeply")
}

func (s *SSMParameterTreeResourceSchemaSuite) Test_validate_rejects_key_in_both_maps() {
	diagnostics := validateTreePathForSpec(
		"/my-app/config",
		map[string]string{"apiToken": "plain"},
		map[string]string{"apiToken": "secret"},
	)
	s.Len(diagnostics, 1)
	s.Contains(diagnostics[0].Message, "both")
}

func (s *SSMParameterTreeResourceSchemaSuite) Test_validate_rejects_empty_tree() {
	diagnostics := validateTreePathForSpec("/my-app/config", nil, nil)
	s.Len(diagnostics, 1)
	s.Contains(diagnostics[0].Message, "at least one entry")
}

func (s *SSMParameterTreeResourceSchemaSuite) Test_validate_rejects_invalid_path() {
	diagnostics := validateTreePathForSpec(
		"no-leading-slash",
		map[string]string{"logLevel": "info"},
		nil,
	)
	s.NotEmpty(diagnostics)
}

func validateTreePathForSpec(
	path string,
	values map[string]string,
	secureValues map[string]string,
) []*core.Diagnostic {
	specFields := map[string]*core.MappingNode{
		"path": core.MappingNodeFromString(path),
	}
	if values != nil {
		specFields["values"] = core.MappingNodeFromStringMap(values)
	}
	if secureValues != nil {
		specFields["secureValues"] = core.MappingNodeFromStringMap(secureValues)
	}

	return validateParameterTreePath(
		"path",
		core.MappingNodeFromString(path),
		&schema.Resource{Spec: &core.MappingNode{Fields: specFields}},
	)
}

func TestSSMParameterTreeResourceSchemaSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterTreeResourceSchemaSuite))
}
