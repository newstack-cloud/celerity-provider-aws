package lambdakms

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func functionKeyLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.lambda.kms.populateEnvVars": {
			Name:  "aws.lambda.kms.populateEnvVars",
			Label: "Populate Environment Variables",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from lambda function to reference the target KMS key. " +
				"This will populate environment variables for all target keys that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.kms.<targetKey>.populateEnvVars": {
			Name:  "aws.lambda.kms.<targetKey>.populateEnvVars",
			Label: "Populate Environment Variables for Specific Target Key",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from lambda function for a specific target KMS key. " +
				"The default behaviour is to populate environment variables for all target keys that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.kms.<targetKey>.envVarName": {
			Name:  "aws.lambda.kms.<targetKey>.envVarName",
			Label: "Environment Variable Name",
			Type:  core.ScalarTypeString,
			Description: "The name of the environment variable to populate in the linked from lambda function " +
				"to reference the target KMS key ARN. " +
				"The default format for key environment variables is `KMS_KEY_<targetKey>`.",
			Examples: []*core.ScalarValue{
				core.ScalarFromString("DATA_ENCRYPTION_KEY"),
				core.ScalarFromString("SIGNING_KEY_ARN"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.lambda.kms.<targetKey>.accessLevel": {
			Name:  "aws.lambda.kms.<targetKey>.accessLevel",
			Label: "Access Level",
			Type:  core.ScalarTypeString,
			Description: "The level of cryptographic access to grant to the KMS key. " +
				"Valid values are 'decrypt' (default) or 'encryptDecrypt'.",
			DefaultValue: core.ScalarFromString("decrypt"),
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("decrypt"),
				core.ScalarFromString("encryptDecrypt"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.lambda.kms.<targetKey>.manageKeyGrant": {
			Name:  "aws.lambda.kms.<targetKey>.manageKeyGrant",
			Label: "Manage Key Grant",
			Type:  core.ScalarTypeBool,
			Description: "When true, the link additionally creates a KMS grant for the function's " +
				"execution role covering the granted operations, guaranteeing key-side authorisation " +
				"regardless of whether the key policy delegates to IAM. Use this when the key policy has " +
				"been locked down (the default 'Enable IAM policies' delegation removed). The grant is " +
				"revoked when the link is destroyed or this is set back to false. Defaults to false.",
			DefaultValue: core.ScalarFromBool(false),
			Required:     false,
		},
	}
}
