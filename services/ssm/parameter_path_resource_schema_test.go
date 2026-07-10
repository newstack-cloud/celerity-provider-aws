//go:build unit

package ssm

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ssmmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ssm_mock"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type SSMParameterPathResourceSchemaSuite struct {
	suite.Suite
}

func (s *SSMParameterPathResourceSchemaSuite) Test_path_validation() {
	testCases := []struct {
		name          string
		path          string
		expectedError string
	}{
		{
			name: "accepts a typical two-level path",
			path: "/my-app/config",
		},
		{
			name: "accepts a single-level path",
			path: "/my-app",
		},
		{
			name: "accepts segments with dots, dashes and underscores",
			path: "/my-app/env_1/v2.0-beta",
		},
		{
			name:          "rejects a path without a leading slash",
			path:          "my-app/config",
			expectedError: "must start with a forward slash",
		},
		{
			name:          "rejects a path with a trailing slash",
			path:          "/my-app/config/",
			expectedError: "must not end with a forward slash",
		},
		{
			name:          "rejects empty segments",
			path:          "/my-app//config",
			expectedError: "non-empty segments",
		},
		{
			name:          "rejects a bare slash",
			path:          "/",
			expectedError: "must not end with a forward slash",
		},
		{
			name:          "rejects invalid characters in segments",
			path:          "/my-app/co nfig",
			expectedError: "non-empty segments",
		},
		{
			name:          "rejects more levels than parameters can nest beneath",
			path:          "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o",
			expectedError: "at most 14 levels",
		},
		{
			name:          "rejects the reserved aws prefix",
			path:          "/aws/config",
			expectedError: "reserved by AWS",
		},
		{
			name:          "rejects the reserved ssm prefix regardless of case",
			path:          "/SSM/config",
			expectedError: "reserved by AWS",
		},
	}

	pathSchema := s.parameterPathFieldSchema()

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			diagnostics := pathSchema.ValidateFunc(
				"path",
				core.MappingNodeFromString(testCase.path),
				nil,
			)

			if testCase.expectedError == "" {
				s.Empty(diagnostics)
				return
			}

			s.Require().NotEmpty(diagnostics)
			s.True(
				strings.Contains(diagnostics[0].Message, testCase.expectedError),
				"expected diagnostic %q to contain %q",
				diagnostics[0].Message,
				testCase.expectedError,
			)
		})
	}
}

func (s *SSMParameterPathResourceSchemaSuite) parameterPathFieldSchema() *provider.ResourceDefinitionsSchema {
	service := ssmmock.CreateSSMServiceMock()
	resource := ParameterPathResource(
		func(*aws.Config, provider.Context) ssmservice.Service { return service },
		nil,
	)

	output, err := resource.GetSpecDefinition(
		context.Background(),
		&provider.ResourceGetSpecDefinitionInput{},
	)
	s.Require().NoError(err)

	pathSchema := output.SpecDefinition.Schema.Attributes["path"]
	s.Require().NotNil(pathSchema)
	s.Require().NotNil(pathSchema.ValidateFunc)
	return pathSchema
}

func TestSSMParameterPathResourceSchemaSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterPathResourceSchemaSuite))
}
