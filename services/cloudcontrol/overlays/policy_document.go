package overlays

import (
	"fmt"
	"strings"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// AWS models policy documents (and EventBridge event-bus policies / statements) as
// free-form objects in its CloudFormation schemas, so the generator emits them as
// free-form maps. Their structure is well known, though, so these overlays re-type them
// as structured objects with camelCase fields. This gives every policy-bearing
// resource consistent, validated, camelCase authoring (the same as the rest of the
// spec), with the engine translating the camelCase fields to the PascalCase shape
// Cloud Control expects.
func init() {
	registerPolicyDocumentOverlay("aws/iam/role", "assumeRolePolicyDocument", "policies.*.policyDocument")
	registerPolicyDocumentOverlay("aws/iam/managedPolicy", "policyDocument")
	registerPolicyDocumentOverlay("aws/iam/rolePolicy", "policyDocument")
	registerPolicyDocumentOverlay("aws/iam/userPolicy", "policyDocument")
	registerPolicyDocumentOverlay("aws/iam/groupPolicy", "policyDocument")
	registerPolicyDocumentOverlay("aws/iam/group", "policies.*.policyDocument")
	registerPolicyDocumentOverlay("aws/iam/user", "policies.*.policyDocument")
	registerPolicyDocumentOverlay("aws/sqs/queueInlinePolicy", "policyDocument")
	registerPolicyDocumentOverlay("aws/sns/topicInlinePolicy", "policyDocument")
	registerPolicyDocumentOverlay("aws/s3/bucketPolicy", "policyDocument")
	registerPolicyDocumentOverlay("aws/events/eventBus", "policy")
	registerPolicyDocumentOverlay(
		"aws/dynamodb/table",
		"resourcePolicy.policyDocument",
		"streamSpecification.resourcePolicy.policyDocument",
	)
	registerPolicyDocumentOverlay(
		"aws/dynamodb/globalTable",
		"replicas.*.resourcePolicy.policyDocument",
		"replicas.*.replicaStreamSpecification.resourcePolicy.policyDocument",
	)
	registerStatementOverlay("aws/events/eventBusPolicy", "statement")
}

func registerPolicyDocumentOverlay(blueprintType string, fieldPaths ...string) {
	Register(blueprintType, func(
		schema *provider.ResourceDefinitionsSchema,
	) *provider.ResourceDefinitionsSchema {
		for _, path := range fieldPaths {
			retypeAtPath(schema, path, func(
				existing *provider.ResourceDefinitionsSchema,
			) *provider.ResourceDefinitionsSchema {
				return policyDocumentSchema(existing.Label, existing.Description, existing.Nullable)
			})
		}
		return schema
	})

	overrides := map[string]string{}
	for _, path := range fieldPaths {
		key := fmt.Sprintf("%s.statement.*.principal.aws", path)
		overrides[key] = "AWS"
	}
	RegisterMetaAdjustment(blueprintType, &MetaAdjustment{AddFieldNameOverrides: overrides})
}

func registerStatementOverlay(blueprintType, fieldPath string) {
	Register(blueprintType, func(
		schema *provider.ResourceDefinitionsSchema,
	) *provider.ResourceDefinitionsSchema {
		retypeAtPath(schema, fieldPath, func(
			existing *provider.ResourceDefinitionsSchema,
		) *provider.ResourceDefinitionsSchema {
			stmt := statementSchema()
			stmt.Label = existing.Label
			stmt.Description = existing.Description
			stmt.Nullable = existing.Nullable
			return stmt
		})
		return schema
	})
	RegisterMetaAdjustment(blueprintType, &MetaAdjustment{
		AddFieldNameOverrides: map[string]string{fieldPath + ".principal.aws": "AWS"},
	})
}

// A "*" path segment descends into an array's items or a map's values.
func retypeAtPath(
	schema *provider.ResourceDefinitionsSchema,
	path string,
	retype func(*provider.ResourceDefinitionsSchema) *provider.ResourceDefinitionsSchema,
) {
	retypeSegments(schema, strings.Split(path, "."), retype)
}

func retypeSegments(
	schema *provider.ResourceDefinitionsSchema,
	segments []string,
	retype func(*provider.ResourceDefinitionsSchema) *provider.ResourceDefinitionsSchema,
) {
	if schema == nil || len(segments) == 0 {
		return
	}
	segment, rest := segments[0], segments[1:]

	if segment == "*" {
		retypeSegments(schema.Items, rest, retype)
		retypeSegments(schema.MapValues, rest, retype)
		return
	}

	if schema.Attributes == nil {
		return
	}
	child, ok := schema.Attributes[segment]
	if !ok {
		return
	}
	if len(rest) == 0 {
		schema.Attributes[segment] = retype(child)
		return
	}
	retypeSegments(child, rest, retype)
}

func policyDocumentSchema(label, description string, nullable bool) *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       label,
		Description: description,
		Nullable:    nullable,
		Required:    []string{"version", "statement"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"version": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Label:       "Version",
				Description: "The policy language version. The only valid value is '2012-10-17'.",
			},
			"statement": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Label:       "Statement",
				Description: "An array of individual statements in the policy document.",
				Items:       statementSchema(),
			},
		},
	}
}

func statementSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:     provider.ResourceDefinitionsSchemaTypeObject,
		Label:    "Statement",
		Required: []string{"effect"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"sid": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Label:       "Sid",
				Description: "An optional identifier for the statement.",
				Nullable:    true,
			},
			"effect": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Label:       "Effect",
				Description: "Whether the statement allows or denies access ('Allow' or 'Deny').",
			},
			"action":   stringOrStringArray("Action", "The action or actions the statement covers."),
			"resource": stringOrStringArray("Resource", "The resource or resources the statement covers."),
			"principal": {
				Type:        provider.ResourceDefinitionsSchemaTypeUnion,
				Label:       "Principal",
				Description: "The principal the statement applies to (a wildcard string or an object).",
				Nullable:    true,
				OneOf: []*provider.ResourceDefinitionsSchema{
					{Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Principal"},
					principalObjectSchema(),
				},
			},
			"condition": conditionSchema(),
		},
	}
}

func principalObjectSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:     provider.ResourceDefinitionsSchemaTypeObject,
		Label:    "Principal",
		Nullable: true,
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"service":       stringOrStringArray("Service", "AWS service principals."),
			"aws":           stringOrStringArray("AWS", "AWS account or IAM principals."),
			"federated":     stringOrStringArray("Federated", "Federated identity providers."),
			"canonicalUser": {Type: provider.ResourceDefinitionsSchemaTypeString, Label: "Canonical User", Nullable: true},
		},
	}
}

func conditionSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeMap,
		Label:       "Condition",
		Description: "The conditions under which the statement is in effect.",
		Nullable:    true,
		MapValues: &provider.ResourceDefinitionsSchema{
			Type:      provider.ResourceDefinitionsSchemaTypeMap,
			Label:     "Condition Operator",
			MapValues: stringOrStringArray("Condition Value", "A condition value or values."),
		},
	}
}

func stringOrStringArray(label, description string) *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeUnion,
		Label:       label,
		Description: description,
		Nullable:    true,
		OneOf: []*provider.ResourceDefinitionsSchema{
			{Type: provider.ResourceDefinitionsSchemaTypeString, Label: label},
			{
				Type:  provider.ResourceDefinitionsSchemaTypeArray,
				Label: label,
				Items: &provider.ResourceDefinitionsSchema{Type: provider.ResourceDefinitionsSchemaTypeString, Label: label},
			},
		},
	}
}
