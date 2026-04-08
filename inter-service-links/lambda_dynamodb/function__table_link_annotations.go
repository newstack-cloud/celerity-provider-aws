package lambdadynamodb

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func lambdaFunctionDynamoDBTableLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.lambda.dynamodb.populateEnvVars": {
			Name:  "aws.lambda.dynamodb.populateEnvVars",
			Label: "Populate Environment Variables",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from lambda function to reference the target DynamoDB table. " +
				"This will populate environment variables for all target DynamoDB tables that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.dynamodb.<targetTable>.populateEnvVars": {
			Name:  "aws.lambda.dynamodb.<targetTable>.populateEnvVars",
			Label: "Populate Environment Variables for Specific Target Table",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from lambda function for a specific target DynamoDB table. " +
				"The default behaviour is to populate environment variables for all target tables that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.dynamodb.<targetTable>.envVarName": {
			Name:  "aws.lambda.dynamodb.<targetTable>.envVarName",
			Label: "Environment Variable Name",
			Type:  core.ScalarTypeString,
			Description: "The name of the environment variable to populate in the linked from lambda function " +
				"to reference the target DynamoDB table. " +
				"The default format for table name environment variables is `DYNAMODB_TABLE_<targetTable>`.",
			Examples: []*core.ScalarValue{
				core.ScalarFromString("ORDERS_TABLE_NAME"),
				core.ScalarFromString("USERS_TABLE"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.lambda.dynamodb.<targetTable>.accessLevel": {
			Name:  "aws.lambda.dynamodb.<targetTable>.accessLevel",
			Label: "Access Level",
			Type:  core.ScalarTypeString,
			Description: "The level of access to grant to the DynamoDB table. " +
				"Valid values are 'read', 'write', or 'readwrite' (default).",
			DefaultValue: core.ScalarFromString("readwrite"),
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("read"),
				core.ScalarFromString("write"),
				core.ScalarFromString("readwrite"),
			},
			Required: false,
		},
	}
}
