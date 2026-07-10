package ssm

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
)

func parameterPathResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "SSMParameterPathDefinition",
		Description: "The definition of an AWS Systems Manager (SSM) parameter path.",
		Required:    []string{"path"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"path": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The parameter hierarchy path prefix, starting with a forward slash and " +
					"without a trailing slash. For example \"/my-app/config\". Individual parameters " +
					"live beneath the path (e.g. \"/my-app/config/db-host\"). Changing the path " +
					"replaces the resource.",
				FormattedDescription: "The parameter hierarchy path prefix, starting with a forward " +
					"slash and without a trailing slash. For example `/my-app/config`. Individual " +
					"parameters live beneath the path (e.g. `/my-app/config/db-host`). Changing the " +
					"path replaces the resource.",
				MustRecreate: true,
				MinLength:    2,
				// Fully-qualified parameter names are capped at 1011 characters by SSM;
				// leaf names must still fit beneath the prefix.
				MaxLength: 1000,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("/my-app/config"),
				},
				ValidateFunc: validateParameterPath,
			},
		},
	}
}

var parameterPathPattern = regexp.MustCompile(`^/[a-zA-Z0-9_.-]+(/[a-zA-Z0-9_.-]+)*$`)

// SSM allows a maximum of 15 hierarchy levels for a parameter name; leaf parameters must
// nest at least one level beneath the prefix, so the prefix itself is capped at 14.
const maxParameterPathLevels = 14

func validateParameterPath(
	_ string,
	value *core.MappingNode,
	_ *schema.Resource,
) []*core.Diagnostic {
	path := core.StringValue(value)

	if !strings.HasPrefix(path, "/") {
		return []*core.Diagnostic{errorDiagnostic(
			"\"path\" must start with a forward slash, for example \"/my-app/config\".",
		)}
	}
	if strings.HasSuffix(path, "/") {
		return []*core.Diagnostic{errorDiagnostic(
			"\"path\" must not end with a forward slash; the trailing slash is added by consumers where needed.",
		)}
	}
	if !parameterPathPattern.MatchString(path) {
		return []*core.Diagnostic{errorDiagnostic(
			"\"path\" must be one or more non-empty segments separated by forward slashes, " +
				"where each segment contains only letters, digits, and the characters \"_.-\".",
		)}
	}

	var diagnostics []*core.Diagnostic
	levels := strings.Count(path, "/")
	if levels > maxParameterPathLevels {
		diagnostics = append(diagnostics, errorDiagnostic(fmt.Sprintf(
			"\"path\" must have at most %d levels so that parameters can nest beneath it; got %d.",
			maxParameterPathLevels, levels,
		)))
	}

	firstSegment := strings.ToLower(strings.SplitN(path[1:], "/", 2)[0])
	if strings.HasPrefix(firstSegment, "aws") || strings.HasPrefix(firstSegment, "ssm") {
		diagnostics = append(diagnostics, errorDiagnostic(
			"\"path\" must not begin with \"aws\" or \"ssm\"; these prefixes are reserved by AWS.",
		))
	}

	return diagnostics
}
