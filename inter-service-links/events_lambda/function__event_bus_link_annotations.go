package eventslambda

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func functionEventBusLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.lambda.events.populateEnvVars": {
			Name:  "aws.lambda.events.populateEnvVars",
			Label: "Populate Environment Variables",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from Lambda function to reference the target event bus. " +
				"This populates environment variables for all target event buses that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.events.<targetBus>.populateEnvVars": {
			Name:  "aws.lambda.events.<targetBus>.populateEnvVars",
			Label: "Populate Environment Variables for Specific Target Event Bus",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from Lambda function for a specific target event bus. " +
				"The default behaviour is to populate environment variables for all target event buses that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.events.<targetBus>.envVarName": {
			Name:  "aws.lambda.events.<targetBus>.envVarName",
			Label: "Environment Variable Name",
			Type:  core.ScalarTypeString,
			Description: "The name of the environment variable to populate in the linked from Lambda function " +
				"to reference the target event bus. " +
				"The default format for event bus environment variables is `EVENT_BUS_<targetBus>`.",
			Examples: []*core.ScalarValue{
				core.ScalarFromString("ORDER_EVENT_BUS"),
			},
			Required: false,
		},
	}
}
