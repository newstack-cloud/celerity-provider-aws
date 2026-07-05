package lambdasecretsmanager

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func functionSecretLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.lambda.secretsmanager.populateEnvVars": {
			Name:  "aws.lambda.secretsmanager.populateEnvVars",
			Label: "Populate Environment Variables",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from lambda function to reference the target Secrets Manager secret. " +
				"This will populate environment variables for all target secrets that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.secretsmanager.<targetSecret>.populateEnvVars": {
			Name:  "aws.lambda.secretsmanager.<targetSecret>.populateEnvVars",
			Label: "Populate Environment Variables for Specific Target Secret",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from lambda function for a specific target Secrets Manager secret. " +
				"The default behaviour is to populate environment variables for all target secrets that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.secretsmanager.<targetSecret>.envVarName": {
			Name:  "aws.lambda.secretsmanager.<targetSecret>.envVarName",
			Label: "Environment Variable Name",
			Type:  core.ScalarTypeString,
			Description: "The name of the environment variable to populate in the linked from lambda function " +
				"to reference the target Secrets Manager secret ARN. " +
				"The default format for secret environment variables is `SECRET_<targetSecret>`.",
			Examples: []*core.ScalarValue{
				core.ScalarFromString("DB_CREDENTIALS_SECRET"),
				core.ScalarFromString("API_KEY_SECRET_ARN"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.lambda.secretsmanager.<targetSecret>.accessLevel": {
			Name:  "aws.lambda.secretsmanager.<targetSecret>.accessLevel",
			Label: "Access Level",
			Type:  core.ScalarTypeString,
			Description: "The level of access to grant to the Secrets Manager secret. " +
				"Valid values are 'read' (default), 'write', or 'readwrite'.",
			DefaultValue: core.ScalarFromString("read"),
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("read"),
				core.ScalarFromString("write"),
				core.ScalarFromString("readwrite"),
			},
			Required: false,
		},
	}
}
