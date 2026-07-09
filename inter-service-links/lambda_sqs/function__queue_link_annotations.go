package lambdasqs

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func functionQueueLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.lambda.sqs.populateEnvVars": {
			Name:  "aws.lambda.sqs.populateEnvVars",
			Label: "Populate Environment Variables",
			Type:  core.ScalarTypeBool,
			Description: "Whether to populate environment variables in the linked from Lambda function " +
				"to reference the target SQS queue URL. Populates for all target queues that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.sqs.<targetQueue>.populateEnvVars": {
			Name:  "aws.lambda.sqs.<targetQueue>.populateEnvVars",
			Label: "Populate Environment Variables for Specific Target Queue",
			Type:  core.ScalarTypeBool,
			Description: "Whether to populate environment variables in the linked from Lambda function " +
				"for a specific target queue. Overrides the global setting.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.sqs.<targetQueue>.envVarName": {
			Name:  "aws.lambda.sqs.<targetQueue>.envVarName",
			Label: "Environment Variable Name",
			Type:  core.ScalarTypeString,
			Description: "The name of the environment variable to populate in the linked from Lambda function " +
				"to reference the target queue URL. Defaults to `SQS_QUEUE_<targetQueue>`.",
			Examples: []*core.ScalarValue{
				core.ScalarFromString("ORDERS_QUEUE"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.lambda.sqs.<targetQueue>.accessLevel": {
			Name:  "aws.lambda.sqs.<targetQueue>.accessLevel",
			Label: "Access Level",
			Type:  core.ScalarTypeString,
			Description: "The access the function is granted to the queue: 'send' (sqs:SendMessage), 'receive' " +
				"(sqs:ReceiveMessage/DeleteMessage) or 'sendReceive' (both). Defaults to 'send'. For event-driven " +
				"consumption of a queue, use the aws/sqs/queue -> aws/lambda/function link (an event source mapping) " +
				"instead.",
			DefaultValue: core.ScalarFromString("send"),
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("send"),
				core.ScalarFromString("receive"),
				core.ScalarFromString("sendReceive"),
			},
			Required: false,
		},
	}
}
