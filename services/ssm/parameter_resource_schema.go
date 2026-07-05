package ssm

import (
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
)

func parameterResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "SSMParameterDefinition",
		Description: "The definition of an AWS Systems Manager (SSM) parameter.",
		Required:    []string{"name", "type"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"name": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The fully-qualified name of the parameter, including any hierarchy. " +
					"For example \"/my-app/prod/db-host\". Changing the name replaces the parameter.",
				FormattedDescription: "The fully-qualified name of the parameter, including any hierarchy. " +
					"For example `/my-app/prod/db-host`. Changing the name replaces the parameter.",
				MustRecreate: true,
				MinLength:    1,
				MaxLength:    2048,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("/my-app/prod/db-host"),
				},
			},
			"type": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The type of the parameter. \"String\" and \"StringList\" hold plaintext values " +
					"(set via \"value\"); \"SecureString\" holds a KMS-encrypted value (set via \"secureValue\").",
				FormattedDescription: "The type of the parameter. `String` and `StringList` hold plaintext " +
					"values (set via `value`); `SecureString` holds a KMS-encrypted value (set via `secureValue`).",
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("String"),
					core.MappingNodeFromString("StringList"),
					core.MappingNodeFromString("SecureString"),
				},
				// Cross-field rule: exactly one of value/secureValue must be set, and which one
				// depends on the type. SecureString -> secureValue (sensitive); String/StringList
				// -> value (plaintext, visible in diffs).
				ValidateFunc: validateParameterTypeAndValue,
			},
			"value": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The plaintext value of the parameter, used when \"type\" is \"String\" or " +
					"\"StringList\". For \"StringList\", provide a comma-separated string. This value is NOT " +
					"redacted from diffs and logs; use \"secureValue\" with type \"SecureString\" for secrets.",
				FormattedDescription: "The plaintext value of the parameter, used when `type` is `String` or " +
					"`StringList`. For `StringList`, provide a comma-separated string. This value is **not** " +
					"redacted from diffs and logs; use `secureValue` with type `SecureString` for secrets.",
			},
			"secureValue": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The value of a \"SecureString\" parameter. It is encrypted at rest by KMS using " +
					"\"keyId\" (or the AWS-managed alias/aws/ssm key when omitted), and is redacted from diffs and logs.",
				FormattedDescription: "The value of a `SecureString` parameter. It is encrypted at rest by KMS " +
					"using `keyId` (or the AWS-managed `alias/aws/ssm` key when omitted), and is redacted from " +
					"diffs and logs.",
				Sensitive: true,
			},
			"keyId": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The KMS key id, ARN or alias used to encrypt a \"SecureString\" parameter. " +
					"Defaults to the AWS-managed \"alias/aws/ssm\" key. Only valid when \"type\" is \"SecureString\".",
				FormattedDescription: "The KMS key id, ARN or alias used to encrypt a `SecureString` parameter. " +
					"Defaults to the AWS-managed `alias/aws/ssm` key. Only valid when `type` is `SecureString`.",
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("alias/my-key"),
				},
				// When omitted, AWS resolves the default alias/aws/ssm key and reports its
				// resolved id in external state, which would otherwise read as drift against an
				// unset keyId. Ignore drift so the default-key case does not thrash.
				IgnoreDrift: true,
			},
			"description": {
				Type:                 provider.ResourceDefinitionsSchemaTypeString,
				Description:          "Information about the parameter.",
				FormattedDescription: "Information about the parameter.",
				MaxLength:            1024,
			},
			"tier": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The parameter tier. \"Standard\" (default, free, up to 4KB), \"Advanced\" " +
					"(up to 8KB, chargeable, supports policies), or \"Intelligent-Tiering\".",
				FormattedDescription: "The parameter tier. `Standard` (default, free, up to 4KB), `Advanced` " +
					"(up to 8KB, chargeable, supports policies), or `Intelligent-Tiering`.",
				Default: core.MappingNodeFromString("Standard"),
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("Standard"),
					core.MappingNodeFromString("Advanced"),
					core.MappingNodeFromString("Intelligent-Tiering"),
				},
			},
			"allowedPattern": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "A regular expression used to validate the parameter value. For example, " +
					"\"^\\d+$\" restricts the value to digits.",
				FormattedDescription: "A regular expression used to validate the parameter value. For example, " +
					"`^\\d+$` restricts the value to digits.",
				MaxLength: 1024,
			},
			"dataType": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The data type of the parameter. \"text\" (default) for arbitrary values, or " +
					"\"aws:ec2:image\"/\"aws:ssm:integration\" for validated values.",
				FormattedDescription: "The data type of the parameter. `text` (default) for arbitrary values, or " +
					"`aws:ec2:image`/`aws:ssm:integration` for validated values.",
				Default: core.MappingNodeFromString("text"),
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("text"),
					core.MappingNodeFromString("aws:ec2:image"),
					core.MappingNodeFromString("aws:ssm:integration"),
				},
			},
			"policies": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "A JSON string describing the parameter policies (expiration, expiration " +
					"notification, no-change notification). Only supported on the \"Advanced\" tier.",
				FormattedDescription: "A JSON string describing the parameter policies (expiration, expiration " +
					"notification, no-change notification). Only supported on the `Advanced` tier. See " +
					"[Assigning parameter policies](https://docs.aws.amazon.com/systems-manager/latest/userguide/parameter-store-policies.html).",
				// AWS returns policies as structured metadata that does not round-trip byte-for-byte
				// to the submitted JSON string, so drift on this field is ignored to avoid false positives.
				IgnoreDrift: true,
			},
			"tags": {
				Type:                 provider.ResourceDefinitionsSchemaTypeMap,
				Description:          "A map of tags to attach to the parameter.",
				FormattedDescription: "A map of tags to attach to the parameter.",
				MapValues: &provider.ResourceDefinitionsSchema{
					Type: provider.ResourceDefinitionsSchemaTypeString,
				},
				Examples: []*core.MappingNode{
					core.MappingNodeFields(
						"Environment", core.ScalarFromString("production"),
					),
				},
			},

			// Computed fields.
			"arn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The ARN of the parameter.",
				Computed:    true,
			},
			"version": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The version of the parameter. Incremented each time the value is updated.",
				Computed:    true,
			},
		},
	}
}

func validateParameterTypeAndValue(
	_ string,
	value *core.MappingNode,
	resource *schema.Resource,
) []*core.Diagnostic {
	paramType := core.StringValue(value)

	valueNode, _ := core.GetPathValue("$.value", resource.Spec, core.MappingNodeMaxTraverseDepth)
	secureValueNode, _ := core.GetPathValue("$.secureValue", resource.Spec, core.MappingNodeMaxTraverseDepth)
	hasValue := !core.IsNilMappingNode(valueNode)
	hasSecureValue := !core.IsNilMappingNode(secureValueNode)

	var diagnostics []*core.Diagnostic
	if paramType == "SecureString" {
		if !hasSecureValue {
			diagnostics = append(diagnostics, errorDiagnostic(
				"\"secureValue\" is required when \"type\" is \"SecureString\".",
			))
		}
		if hasValue {
			diagnostics = append(diagnostics, errorDiagnostic(
				"\"value\" cannot be set when \"type\" is \"SecureString\"; use \"secureValue\" for encrypted parameters.",
			))
		}
		return diagnostics
	}

	if !hasValue {
		diagnostics = append(diagnostics, errorDiagnostic(fmt.Sprintf(
			"\"value\" is required when \"type\" is %q.", paramType,
		)))
	}
	if hasSecureValue {
		diagnostics = append(diagnostics, errorDiagnostic(
			"\"secureValue\" can only be set when \"type\" is \"SecureString\".",
		))
	}
	return diagnostics
}

func errorDiagnostic(message string) *core.Diagnostic {
	return &core.Diagnostic{
		Level:   core.DiagnosticLevelError,
		Message: message,
	}
}
