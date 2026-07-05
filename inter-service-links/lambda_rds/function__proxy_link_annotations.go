package lambdards

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func functionProxyLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.lambda.rds.populateEnvVars": {
			Name:  "aws.lambda.rds.populateEnvVars",
			Label: "Populate Environment Variables",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not connection environment variables should be " +
				"populated in the linked from lambda function for all linked RDS proxies.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.rds.<targetProxy>.populateEnvVars": {
			Name:  "aws.lambda.rds.<targetProxy>.populateEnvVars",
			Label: "Populate Environment Variables for Specific Target Proxy",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not connection environment variables should be " +
				"populated in the linked from lambda function for a specific target proxy. Overrides the global setting.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.rds.<targetProxy>.envVarPrefix": {
			Name:  "aws.lambda.rds.<targetProxy>.envVarPrefix",
			Label: "Environment Variable Prefix",
			Type:  core.ScalarTypeString,
			Description: "Prefix for the connection environment variables the link populates: " +
				"<PREFIX>_HOST, <PREFIX>_PORT and <PREFIX>_DATABASE. Defaults to an auto-generated prefix based on the proxy name.",
			Examples: []*core.ScalarValue{
				core.ScalarFromString("ORDERS_DB"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.lambda.rds.<targetProxy>.authMode": {
			Name:  "aws.lambda.rds.<targetProxy>.authMode",
			Label: "Authentication Mode",
			Type:  core.ScalarTypeString,
			Description: "How the function authenticates to the database. 'password' relies on a Secrets Manager " +
				"secret the function is linked to separately; 'iam' grants the execution role rds-db:connect for token-based auth.",
			DefaultValue: core.ScalarFromString("password"),
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("password"),
				core.ScalarFromString("iam"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.lambda.rds.<targetProxy>.dbUser": {
			Name:  "aws.lambda.rds.<targetProxy>.dbUser",
			Label: "Database User",
			Type:  core.ScalarTypeString,
			Description: "The database username the function authenticates as, used to scope the rds-db:connect grant " +
				"when authMode is 'iam'. Defaults to '*' (any IAM-enabled user on the linked database).",
			DefaultValue: core.ScalarFromString("*"),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.rds.<targetProxy>.port": {
			Name:  "aws.lambda.rds.<targetProxy>.port",
			Label: "Database Port",
			Type:  core.ScalarTypeInteger,
			Description: "The database port opened in the security-group rule and populated in the <PREFIX>_PORT " +
				"environment variable.",
			DefaultValue: core.ScalarFromInt(5432),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.rds.<targetProxy>.databaseName": {
			Name:        "aws.lambda.rds.<targetProxy>.databaseName",
			Label:       "Database Name",
			Type:        core.ScalarTypeString,
			Description: "The database name populated in the <PREFIX>_DATABASE environment variable.",
			Required:    false,
		},
	}
}
