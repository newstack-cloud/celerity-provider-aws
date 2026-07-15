package lambdassm

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// The global aws.lambda.ssm.populateEnvVars fallback is intentionally shared with the
// per-parameter and parameter path links so one annotation toggles env var population for
// every linked SSM resource on the function. The accessLevel allowed values must stay
// identical across the three links: each link's dynamic annotation pattern matches the
// others' concrete keys during validation, so diverging values would produce incorrect
// errors.
func functionParameterTreeLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.lambda.ssm.populateEnvVars": {
			Name:  "aws.lambda.ssm.populateEnvVars",
			Label: "Populate Environment Variables",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from lambda function to reference the target SSM resource. " +
				"This will populate environment variables for all target resources that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.ssm.<targetParameterTree>.populateEnvVars": {
			Name:  "aws.lambda.ssm.<targetParameterTree>.populateEnvVars",
			Label: "Populate Environment Variables for Specific Target Parameter Tree",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not an environment variable should be populated " +
				"in the linked from lambda function for a specific target SSM parameter tree. " +
				"The default behaviour is to populate environment variables for all target parameter trees " +
				"that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.ssm.<targetParameterTree>.envVarName": {
			Name:  "aws.lambda.ssm.<targetParameterTree>.envVarName",
			Label: "Environment Variable Name",
			Type:  core.ScalarTypeString,
			Description: "The name of the environment variable to populate in the linked from lambda function " +
				"holding the target SSM parameter tree's path prefix. " +
				"The default format for parameter tree environment variables is `SSM_PARAMETER_PATH_<targetParameterTree>`, " +
				"matching the parameter path link's convention so runtimes can consume either resource type.",
			Examples: []*core.ScalarValue{
				core.ScalarFromString("APP_CONFIG_STORE_PATH"),
				core.ScalarFromString("FEATURE_FLAGS_STORE_PATH"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.lambda.ssm.<targetParameterTree>.accessLevel": {
			Name:  "aws.lambda.ssm.<targetParameterTree>.accessLevel",
			Label: "Access Level",
			Type:  core.ScalarTypeString,
			Description: "The level of access to grant over the SSM parameter tree's path prefix and everything " +
				"beneath it. Valid values are 'read' (default), 'write', or 'readwrite'.",
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
