package lambdasns

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func functionTopicLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.lambda.sns.populateEnvVars": {
			Name:  "aws.lambda.sns.populateEnvVars",
			Label: "Populate Environment Variables",
			Type:  core.ScalarTypeBool,
			Description: "Whether to populate environment variables in the linked from Lambda function " +
				"to reference the target SNS topic. Populates for all target topics that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.sns.<targetTopic>.populateEnvVars": {
			Name:  "aws.lambda.sns.<targetTopic>.populateEnvVars",
			Label: "Populate Environment Variables for Specific Target Topic",
			Type:  core.ScalarTypeBool,
			Description: "Whether to populate environment variables in the linked from Lambda function " +
				"for a specific target topic. Overrides the global setting.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.sns.<targetTopic>.envVarName": {
			Name:  "aws.lambda.sns.<targetTopic>.envVarName",
			Label: "Environment Variable Name",
			Type:  core.ScalarTypeString,
			Description: "The name of the environment variable to populate in the linked from Lambda function " +
				"to reference the target topic ARN. Defaults to `SNS_TOPIC_<targetTopic>`.",
			Examples: []*core.ScalarValue{
				core.ScalarFromString("ORDERS_TOPIC"),
			},
			Required: false,
		},
	}
}
