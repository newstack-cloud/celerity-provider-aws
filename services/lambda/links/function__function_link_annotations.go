package lambdalinks

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func lambdaFunctionFunctionLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	// Intra-service link annotations use aws.lambda.invoke.* pattern
	// where "invoke" is the feature being configured
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.lambda.invoke.populateEnvVars": {
			Name:  "aws.lambda.invoke.populateEnvVars",
			Label: "Populate Environment Variables",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from lambda function in order to invoke the target lambda function. " +
				"This will populate environment variables for all target lambda functions that match the selector of the callee function.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
			AppliesTo:    provider.LinkAnnotationResourceA,
		},
		"aws/lambda/function::aws.lambda.invoke.<targetFunction>.populateEnvVars": {
			Name:  "aws.lambda.invoke.<targetFunction>.populateEnvVars",
			Label: "Populate Environment Variables for Specific Target Function",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from lambda function in order to invoke a specific target lambda function. " +
				"The default behaviour is to populate environment variables for all target lambda functions that match the selector of the callee function.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
			AppliesTo:    provider.LinkAnnotationResourceA,
		},
		"aws/lambda/function::aws.lambda.invoke.<targetFunction>.envVarName": {
			Name:  "aws.lambda.invoke.<targetFunction>.envVarName",
			Label: "Environment Variable Name",
			Type:  core.ScalarTypeString,
			Description: "The name of the environment variable to populate in the linked from lambda function " +
				"in order to invoke the target lambda function. " +
				"The default format for lambda function name environment variables is `AWS_LAMBDA_FUNCTION_<targetFunction>`.",
			DefaultValue: core.ScalarFromBool(true),
			Examples: []*core.ScalarValue{
				core.ScalarFromString("AWS_LAMBDA_FUNCTION_ORDERS"),
			},
			Required:  false,
			AppliesTo: provider.LinkAnnotationResourceA,
		},
	}
}
