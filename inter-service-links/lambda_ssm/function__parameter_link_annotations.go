package lambdassm

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func functionParameterLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.lambda.ssm.populateEnvVars": {
			Name:  "aws.lambda.ssm.populateEnvVars",
			Label: "Populate Environment Variables",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from lambda function to reference the target SSM parameter. " +
				"This will populate environment variables for all target parameters that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.ssm.<targetParameter>.populateEnvVars": {
			Name:  "aws.lambda.ssm.<targetParameter>.populateEnvVars",
			Label: "Populate Environment Variables for Specific Target Parameter",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from lambda function for a specific target SSM parameter. " +
				"The default behaviour is to populate environment variables for all target parameters that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.ssm.<targetParameter>.envVarName": {
			Name:  "aws.lambda.ssm.<targetParameter>.envVarName",
			Label: "Environment Variable Name",
			Type:  core.ScalarTypeString,
			Description: "The name of the environment variable to populate in the linked from lambda function " +
				"to reference the target SSM parameter name. " +
				"The default format for parameter environment variables is `SSM_PARAMETER_<targetParameter>`.",
			Examples: []*core.ScalarValue{
				core.ScalarFromString("DB_HOST_PARAM"),
				core.ScalarFromString("FEATURE_FLAGS_PARAM_NAME"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.lambda.ssm.<targetParameter>.accessLevel": {
			Name:  "aws.lambda.ssm.<targetParameter>.accessLevel",
			Label: "Access Level",
			Type:  core.ScalarTypeString,
			Description: "The level of access to grant to the SSM parameter. " +
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
