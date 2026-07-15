package ssm

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
)

func parameterTreeResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "SSMParameterTreeDefinition",
		Description: "The definition of an AWS Systems Manager (SSM) parameter tree.",
		Required:    []string{"path"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"path": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The parameter hierarchy path prefix the tree owns, starting with a forward " +
					"slash and without a trailing slash. For example \"/my-app/config\". One parameter is " +
					"created per entry in \"values\"/\"secureValues\" at \"<path>/<key>\". Changing the " +
					"path replaces the resource.",
				FormattedDescription: "The parameter hierarchy path prefix the tree owns, starting with a " +
					"forward slash and without a trailing slash. For example `/my-app/config`. One parameter " +
					"is created per entry in `values`/`secureValues` at `<path>/<key>`. Changing the path " +
					"replaces the resource.",
				MustRecreate: true,
				MinLength:    2,
				// Fully-qualified parameter names are capped at 1011 characters by SSM;
				// leaf names must still fit beneath the prefix.
				MaxLength: 1000,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("/my-app/config"),
				},
				ValidateFunc: validateParameterTreePath,
			},
			"region": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The AWS region to store the tree's parameters in, overriding the " +
					"provider-configured region. Changing the region replaces the tree.",
				FormattedDescription: "The AWS region to store the tree's parameters in, overriding the " +
					"provider-configured region. Changing the region replaces the tree.",
				MustRecreate: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("eu-west-1"),
				},
				ValidateFunc: validateParameterRegion,
				// When omitted, the provider-level region applies and is reported back in
				// external state, which would otherwise read as drift against an unset region.
				IgnoreDrift: true,
			},
			"values": {
				Type: provider.ResourceDefinitionsSchemaTypeMap,
				Description: "A map of key to plaintext value; each entry becomes a \"String\" parameter " +
					"at \"<path>/<key>\". Keys may contain forward slashes to namespace entries within the " +
					"tree (for example \"db/host\"). Stored values are treated like an opaque blob: they " +
					"are applied on create and on explicit blueprint change, but out-of-band changes (for " +
					"example runtime configuration writes) are never reported or reverted as drift.",
				FormattedDescription: "A map of key to plaintext value; each entry becomes a `String` " +
					"parameter at `<path>/<key>`. Keys may contain forward slashes to namespace entries " +
					"within the tree (for example `db/host`). Stored values are treated like an opaque " +
					"blob: they are applied on create and on explicit blueprint change, but out-of-band " +
					"changes (for example runtime configuration writes) are never reported or reverted " +
					"as drift.",
				MapValues: &provider.ResourceDefinitionsSchema{
					Type: provider.ResourceDefinitionsSchemaTypeString,
				},
				Examples: []*core.MappingNode{
					core.MappingNodeFields(
						"logLevel", core.MappingNodeFromString("info"),
						"db/host", core.MappingNodeFromString("db.internal.example.com"),
					),
				},
				// The tree's stored values follow the write-only blob posture: out-of-band
				// value changes must never surface as drift, so the deploy engine cannot
				// revert runtime configuration writes.
				IgnoreDrift: true,
			},
			"secureValues": {
				Type: provider.ResourceDefinitionsSchemaTypeMap,
				Description: "A map of key to secret value; each entry becomes a \"SecureString\" parameter " +
					"at \"<path>/<key>\", encrypted at rest by KMS using \"keyId\" (or the AWS-managed " +
					"alias/aws/ssm key when omitted) and redacted from diffs and logs. Keys may contain " +
					"forward slashes to namespace entries within the tree. Like \"values\", stored secrets " +
					"are never reported or reverted as drift.",
				FormattedDescription: "A map of key to secret value; each entry becomes a `SecureString` " +
					"parameter at `<path>/<key>`, encrypted at rest by KMS using `keyId` (or the " +
					"AWS-managed `alias/aws/ssm` key when omitted) and redacted from diffs and logs. Keys " +
					"may contain forward slashes to namespace entries within the tree. Like `values`, " +
					"stored secrets are never reported or reverted as drift.",
				MapValues: &provider.ResourceDefinitionsSchema{
					Type: provider.ResourceDefinitionsSchemaTypeString,
				},
				Sensitive: true,
				// See "values"; secure entries follow the same write-only blob posture.
				IgnoreDrift: true,
			},
			"keyId": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The KMS key id, ARN or alias used to encrypt the tree's \"secureValues\" " +
					"parameters. Defaults to the AWS-managed \"alias/aws/ssm\" key.",
				FormattedDescription: "The KMS key id, ARN or alias used to encrypt the tree's " +
					"`secureValues` parameters. Defaults to the AWS-managed `alias/aws/ssm` key.",
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("alias/my-key"),
				},
				// When omitted, AWS resolves the default alias/aws/ssm key and reports its
				// resolved id in external state, which would otherwise read as drift against an
				// unset keyId. Ignore drift so the default-key case does not thrash.
				IgnoreDrift: true,
			},
			"tier": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The tier applied to every parameter in the tree. \"Standard\" (default, " +
					"free, up to 4KB), \"Advanced\" (up to 8KB, chargeable), or \"Intelligent-Tiering\".",
				FormattedDescription: "The tier applied to every parameter in the tree. `Standard` " +
					"(default, free, up to 4KB), `Advanced` (up to 8KB, chargeable), or `Intelligent-Tiering`.",
				Default: core.MappingNodeFromString("Standard"),
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("Standard"),
					core.MappingNodeFromString("Advanced"),
					core.MappingNodeFromString("Intelligent-Tiering"),
				},
			},
			"description": {
				Type:                 provider.ResourceDefinitionsSchemaTypeString,
				Description:          "Information about the tree, applied to every parameter in it.",
				FormattedDescription: "Information about the tree, applied to every parameter in it.",
				MaxLength:            1024,
			},
			"tags": {
				Type:                 provider.ResourceDefinitionsSchemaTypeMap,
				Description:          "A map of tags to attach to every parameter in the tree.",
				FormattedDescription: "A map of tags to attach to every parameter in the tree.",
				MapValues: &provider.ResourceDefinitionsSchema{
					Type: provider.ResourceDefinitionsSchemaTypeString,
				},
				Examples: []*core.MappingNode{
					core.MappingNodeFields(
						"Environment", core.MappingNodeFromString("production"),
					),
				},
			},

			// Computed fields.
			"parameters": {
				Type: provider.ResourceDefinitionsSchemaTypeMap,
				Description: "Structural metadata for each parameter the tree manages, keyed by entry key. " +
					"Values are never included; \"valueHash\" is a fingerprint of the last applied value, " +
					"used to detect explicit blueprint changes without reading values back.",
				FormattedDescription: "Structural metadata for each parameter the tree manages, keyed by " +
					"entry key. Values are never included; `valueHash` is a fingerprint of the last applied " +
					"value, used to detect explicit blueprint changes without reading values back.",
				Computed: true,
				MapValues: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeObject,
					Label:       "SSMParameterTreeEntryInfo",
					Description: "Structural metadata for a single parameter managed by the tree.",
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"arn": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The ARN of the parameter.",
						},
						"type": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The parameter type, \"String\" or \"SecureString\".",
						},
						"valueHash": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "A SHA-256 fingerprint of the last value applied by a deploy.",
						},
					},
				},
			},
		},
	}
}

var parameterTreeKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]+(/[a-zA-Z0-9_.-]+)*$`)

// SSM allows a maximum of 15 hierarchy levels for a parameter name; each tree entry's
// fully-qualified name is "<path>/<key>", so path levels plus key segments must fit.
const maxParameterNameLevels = 15

// Applies the shared parameter path rules to the tree's path
// prefix and then validates the entry keys against it: key shape, combined hierarchy
// depth, no key in both value maps, and at least one entry overall.
func validateParameterTreePath(
	fieldName string,
	value *core.MappingNode,
	resource *schema.Resource,
) []*core.Diagnostic {
	diagnostics := validateParameterPath(fieldName, value, resource)

	path := core.StringValue(value)
	pathLevels := strings.Count(path, "/")

	valueKeys := parameterTreeSpecKeys("$.values", resource)
	secureValueKeys := parameterTreeSpecKeys("$.secureValues", resource)

	if len(valueKeys)+len(secureValueKeys) == 0 {
		diagnostics = append(diagnostics, errorDiagnostic(
			"at least one entry must be provided in \"values\" or \"secureValues\"; "+
				"an empty parameter tree manages nothing.",
		))
	}

	for _, key := range append(valueKeys, secureValueKeys...) {
		if !parameterTreeKeyPattern.MatchString(key) {
			diagnostics = append(diagnostics, errorDiagnostic(fmt.Sprintf(
				"entry key %q must be one or more non-empty segments separated by forward slashes, "+
					"where each segment contains only letters, digits, and the characters \"_.-\".",
				key,
			)))
			continue
		}
		keySegments := strings.Count(key, "/") + 1
		if pathLevels+keySegments > maxParameterNameLevels {
			diagnostics = append(diagnostics, errorDiagnostic(fmt.Sprintf(
				"entry key %q nests too deeply: \"path\" contributes %d levels and the key %d more, "+
					"exceeding the SSM maximum of %d levels for a parameter name.",
				key, pathLevels, keySegments, maxParameterNameLevels,
			)))
		}
	}

	secureKeySet := map[string]struct{}{}
	for _, key := range secureValueKeys {
		secureKeySet[key] = struct{}{}
	}
	for _, key := range valueKeys {
		if _, inBoth := secureKeySet[key]; inBoth {
			diagnostics = append(diagnostics, errorDiagnostic(fmt.Sprintf(
				"entry key %q cannot appear in both \"values\" and \"secureValues\".",
				key,
			)))
		}
	}

	return diagnostics
}

func parameterTreeSpecKeys(path string, resource *schema.Resource) []string {
	node, _ := core.GetPathValue(path, resource.Spec, core.MappingNodeMaxTraverseDepth)
	if core.IsNilMappingNode(node) {
		return nil
	}

	keys := make([]string, 0, len(node.Fields))
	for key := range node.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
