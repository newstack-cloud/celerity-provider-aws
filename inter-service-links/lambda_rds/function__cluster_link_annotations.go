package lambdards

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func functionClusterLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		// Shared global toggle (same annotation as the dbProxy link -- one switch for all RDS links).
		"aws/lambda/function::aws.lambda.rds.populateEnvVars": {
			Name:  "aws.lambda.rds.populateEnvVars",
			Label: "Populate Environment Variables",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not connection environment variables should be " +
				"populated in the linked from lambda function for all linked RDS databases.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.rds.<targetCluster>.populateEnvVars": {
			Name:  "aws.lambda.rds.<targetCluster>.populateEnvVars",
			Label: "Populate Environment Variables for Specific Target Cluster",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not connection environment variables should be " +
				"populated in the linked from lambda function for a specific target cluster. Overrides the global setting.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.rds.<targetCluster>.envVarPrefix": {
			Name:  "aws.lambda.rds.<targetCluster>.envVarPrefix",
			Label: "Environment Variable Prefix",
			Type:  core.ScalarTypeString,
			Description: "Prefix for the connection environment variables the link populates: " +
				"<PREFIX>_HOST, <PREFIX>_PORT, <PREFIX>_DATABASE and (when readerEndpoint is enabled) " +
				"<PREFIX>_READER_HOST. Defaults to an auto-generated prefix based on the cluster name.",
			Examples: []*core.ScalarValue{
				core.ScalarFromString("ORDERS_DB"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.lambda.rds.<targetCluster>.authMode": {
			Name:  "aws.lambda.rds.<targetCluster>.authMode",
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
		"aws/lambda/function::aws.lambda.rds.<targetCluster>.dbUser": {
			Name:  "aws.lambda.rds.<targetCluster>.dbUser",
			Label: "Database User",
			Type:  core.ScalarTypeString,
			Description: "The database username the function authenticates as, used to scope the rds-db:connect grant " +
				"when authMode is 'iam'. Defaults to '*' (any IAM-enabled user on the linked cluster).",
			DefaultValue: core.ScalarFromString("*"),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.rds.<targetCluster>.port": {
			Name:  "aws.lambda.rds.<targetCluster>.port",
			Label: "Database Port",
			Type:  core.ScalarTypeInteger,
			Description: "The database port opened in the security-group rule and populated in the <PREFIX>_PORT " +
				"environment variable. Defaults to 5432 (Aurora PostgreSQL); set to 3306 for Aurora MySQL.",
			DefaultValue: core.ScalarFromInt(5432),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.rds.<targetCluster>.databaseName": {
			Name:        "aws.lambda.rds.<targetCluster>.databaseName",
			Label:       "Database Name",
			Type:        core.ScalarTypeString,
			Description: "The database name populated in the <PREFIX>_DATABASE environment variable.",
			Required:    false,
		},
		"aws/lambda/function::aws.lambda.rds.<targetCluster>.readerEndpoint": {
			Name:  "aws.lambda.rds.<targetCluster>.readerEndpoint",
			Label: "Populate Reader Endpoint",
			Type:  core.ScalarTypeBool,
			Description: "When true, also populates <PREFIX>_READER_HOST with the cluster's reader endpoint for " +
				"read scaling across Aurora replicas.",
			DefaultValue: core.ScalarFromBool(false),
			Required:     false,
		},
	}
}
