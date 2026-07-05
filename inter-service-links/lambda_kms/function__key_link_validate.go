package lambdakms

import (
	"context"
	"fmt"
	"strings"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Validate warns, at blueprint validation time, when the linked KMS key defines a custom key
// policy that does not appear to delegate authorisation to IAM (the default "Enable IAM
// policies" statement) and the link is not managing a KMS grant. In that configuration the
// link's IAM grant alone is insufficient and the function will fail at runtime with
// AccessDenied, so this turns a silent runtime failure into a plan-time nudge.
func (l *functionKeyLinkActions) Validate(
	ctx context.Context,
	input *provider.LinkValidateInput,
) (*provider.LinkValidateOutput, error) {
	// If the link is managing a KMS grant, the key-side authorisation is guaranteed
	// independently of the key policy, in this case there is nothing to warn about.
	manageGrantKey := fmt.Sprintf(
		"aws/lambda/function::aws.lambda.kms.%s.manageKeyGrant",
		input.ResourceBName,
	)
	if scalar, ok := input.Annotations[manageGrantKey]; ok && core.BoolValueFromScalar(scalar) {
		return &provider.LinkValidateOutput{}, nil
	}

	// No explicit key policy means AWS applies the default policy, which delegates to IAM.
	keyPolicy, hasKeyPolicy := pluginutils.GetValueByPath("$.keyPolicy", input.ResourceBSpec)
	if !hasKeyPolicy || keyPolicy == nil {
		return &provider.LinkValidateOutput{}, nil
	}

	if keyPolicyDelegatesToIAM(keyPolicy) {
		return &provider.LinkValidateOutput{}, nil
	}

	return &provider.LinkValidateOutput{
		Diagnostics: []*core.Diagnostic{
			{
				Level: core.DiagnosticLevelWarning,
				Message: fmt.Sprintf(
					"The linked KMS key %q defines a custom key policy that does not appear to "+
						"delegate authorisation to IAM (an Allow statement for the account root with "+
						"action \"kms:*\"). This link grants the function's execution role access via IAM "+
						"only, so it may be insufficient and the function could fail at runtime with "+
						"AccessDenied. Set \"aws.lambda.kms.%s.manageKeyGrant\" to true to have the link "+
						"create a KMS grant for the role, or add the role to the key policy.",
					input.ResourceBName,
					input.ResourceBName,
				),
			},
		},
	}, nil
}

// Reports whether the key policy contains the standard IAM-delegation
// statement (Effect Allow, account-root principal, action "kms:*"). It is a best-effort
// heuristic over the free-form policy document: it only returns true when the delegation is
// clearly present, so ambiguous/custom policies will lead to a warning being produced.
func keyPolicyDelegatesToIAM(keyPolicy *core.MappingNode) bool {
	if keyPolicy == nil || keyPolicy.Fields == nil {
		return false
	}
	statementNode := keyPolicy.Fields["Statement"]
	if statementNode == nil {
		return false
	}

	statements := statementNode.Items
	if len(statements) == 0 && statementNode.Fields != nil {
		statements = []*core.MappingNode{statementNode}
	}

	for _, statement := range statements {
		if statement == nil || statement.Fields == nil {
			continue
		}
		if !strings.EqualFold(core.StringValue(statement.Fields["Effect"]), "Allow") {
			continue
		}

		principal := statement.Fields["Principal"]
		if principal == nil || principal.Fields == nil {
			continue
		}
		hasRootPrincipal := false
		for _, awsPrincipal := range scalarOrArrayStrings(principal.Fields["AWS"]) {
			if strings.HasSuffix(awsPrincipal, ":root") {
				hasRootPrincipal = true
				break
			}
		}
		if !hasRootPrincipal {
			continue
		}

		for _, action := range scalarOrArrayStrings(statement.Fields["Action"]) {
			if action == "kms:*" {
				return true
			}
		}
	}

	return false
}

func scalarOrArrayStrings(node *core.MappingNode) []string {
	if node == nil {
		return nil
	}
	if len(node.Items) > 0 {
		values := make([]string, 0, len(node.Items))
		for _, item := range node.Items {
			values = append(values, core.StringValue(item))
		}
		return values
	}
	if value := core.StringValue(node); value != "" {
		return []string{value}
	}
	return nil
}
