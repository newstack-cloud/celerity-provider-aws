package sqslambda

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func sqsQueueLambdaFunctionLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		// resourceB (Lambda function) annotations - configure the event source mapping for this specific function
		// These are on resourceB because a queue can trigger multiple functions with different configurations
		// Annotation names include both services and feature (sqs.lambda) to indicate relationship configuration
		"aws/lambda/function::aws.sqs.lambda.batchSize": {
			Name:  "aws.sqs.lambda.batchSize",
			Label: "Batch Size",
			Type:  core.ScalarTypeInteger,
			Description: "The maximum number of messages to retrieve per batch. " +
				"Default: 10. Range: 1 to 10,000.",
			DefaultValue: core.ScalarFromInt(10),
			Required:     false,
		},
		"aws/lambda/function::aws.sqs.lambda.maximumBatchingWindowInSeconds": {
			Name:  "aws.sqs.lambda.maximumBatchingWindowInSeconds",
			Label: "Maximum Batching Window",
			Type:  core.ScalarTypeInteger,
			Description: "The maximum amount of time (in seconds) to gather messages before invoking the function. " +
				"Required to be greater than 0 when batchSize exceeds 10. Default: 0. Range: 0 to 300.",
			DefaultValue: core.ScalarFromInt(0),
			Required:     false,
		},
		"aws/lambda/function::aws.sqs.lambda.maximumConcurrency": {
			Name:  "aws.sqs.lambda.maximumConcurrency",
			Label: "Maximum Concurrency",
			Type:  core.ScalarTypeInteger,
			Description: "The maximum number of concurrent function invocations the event source mapping can create " +
				"for the queue. Range: 2 to 1,000. When unset, no scaling limit is applied.",
			Required: false,
		},
		"aws/lambda/function::aws.sqs.lambda.reportBatchItemFailures": {
			Name:  "aws.sqs.lambda.reportBatchItemFailures",
			Label: "Report Batch Item Failures",
			Type:  core.ScalarTypeBool,
			Description: "When true, allows the function to report partial batch failures (sets the " +
				"ReportBatchItemFailures function response type) so only failed messages are retried. Default: false.",
			DefaultValue: core.ScalarFromBool(false),
			Required:     false,
		},
		"aws/lambda/function::aws.sqs.lambda.filter.<index>": {
			Name:  "aws.sqs.lambda.filter.<index>",
			Label: "Filter Pattern",
			Type:  core.ScalarTypeString,
			Description: "An event filter pattern controlling which messages are sent to the function. " +
				"`<index>` is the zero-based position in the event source mapping's filter array; add multiple " +
				"filters with successive indices (0, 1, ...). The value is the AWS event-filtering pattern string " +
				"(e.g. `{\"body\":{\"action\":[\"process\"]}}`).",
			Required: false,
		},
		"aws/lambda/function::aws.sqs.lambda.enabled": {
			Name:  "aws.sqs.lambda.enabled",
			Label: "Enabled",
			Type:  core.ScalarTypeBool,
			Description: "When true, the event source mapping is active and Lambda will process messages. " +
				"Default: true.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
	}
}
