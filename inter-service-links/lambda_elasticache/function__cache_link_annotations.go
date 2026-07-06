package lambdaelasticache

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func functionCacheLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.lambda.elasticache.populateEnvVars": {
			Name:  "aws.lambda.elasticache.populateEnvVars",
			Label: "Populate Environment Variables",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not connection environment variables should be " +
				"populated in the linked from lambda function for all linked caches.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.elasticache.<targetCache>.populateEnvVars": {
			Name:  "aws.lambda.elasticache.<targetCache>.populateEnvVars",
			Label: "Populate Environment Variables for Specific Target Cache",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not connection environment variables should be " +
				"populated in the linked from lambda function for a specific target cache. Overrides the global setting.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.elasticache.<targetCache>.envVarPrefix": {
			Name:  "aws.lambda.elasticache.<targetCache>.envVarPrefix",
			Label: "Environment Variable Prefix",
			Type:  core.ScalarTypeString,
			Description: "Prefix for the connection environment variables the link populates: " +
				"<PREFIX>_HOST and <PREFIX>_PORT. Defaults to an auto-generated prefix based on the cache name.",
			Examples: []*core.ScalarValue{
				core.ScalarFromString("SESSION_CACHE"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.lambda.elasticache.<targetCache>.authMode": {
			Name:  "aws.lambda.elasticache.<targetCache>.authMode",
			Label: "Authentication Mode",
			Type:  core.ScalarTypeString,
			Description: "How the function authenticates to the cache. 'password' relies on an AUTH token the " +
				"function reads from a Secrets Manager secret it is linked to separately; 'iam' grants the execution " +
				"role elasticache:Connect for token-based auth (requires an IAM-enabled ElastiCache user).",
			DefaultValue: core.ScalarFromString("password"),
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("password"),
				core.ScalarFromString("iam"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.lambda.elasticache.<targetCache>.userId": {
			Name:  "aws.lambda.elasticache.<targetCache>.userId",
			Label: "ElastiCache User Id",
			Type:  core.ScalarTypeString,
			Description: "The IAM-enabled ElastiCache user id the function authenticates as, used to scope the " +
				"elasticache:Connect grant when authMode is 'iam'. Defaults to '*' (any user attached to the cache's user group).",
			DefaultValue: core.ScalarFromString("*"),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.elasticache.<targetCache>.port": {
			Name:  "aws.lambda.elasticache.<targetCache>.port",
			Label: "Cache Port",
			Type:  core.ScalarTypeInteger,
			Description: "The cache port opened in the security-group rule and populated in the <PREFIX>_PORT " +
				"environment variable.",
			DefaultValue: core.ScalarFromInt(6379),
			Required:     false,
		},
	}
}
