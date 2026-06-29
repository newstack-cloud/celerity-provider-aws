package kinesislambda

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func kinesisStreamLambdaFunctionLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.kinesis.lambda.batchSize": {
			Name:  "aws.kinesis.lambda.batchSize",
			Label: "Batch Size",
			Type:  core.ScalarTypeInteger,
			Description: "The maximum number of records to retrieve per batch. " +
				"Default: 100. Range: 1 to 10,000.",
			DefaultValue: core.ScalarFromInt(100),
			Required:     false,
		},
		"aws/lambda/function::aws.kinesis.lambda.maximumBatchingWindowInSeconds": {
			Name:  "aws.kinesis.lambda.maximumBatchingWindowInSeconds",
			Label: "Maximum Batching Window",
			Type:  core.ScalarTypeInteger,
			Description: "The maximum amount of time (in seconds) to gather records before invoking the function. " +
				"Default: 0. Range: 0 to 300.",
			DefaultValue: core.ScalarFromInt(0),
			Required:     false,
		},
		"aws/lambda/function::aws.kinesis.lambda.startingPosition": {
			Name:  "aws.kinesis.lambda.startingPosition",
			Label: "Starting Position",
			Type:  core.ScalarTypeString,
			Description: "The position in the stream where Lambda should start reading. " +
				"TRIM_HORIZON starts from the oldest record. LATEST starts from the newest record. " +
				"AT_TIMESTAMP starts from the time given by startingPositionTimestamp.",
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("TRIM_HORIZON"),
				core.ScalarFromString("LATEST"),
				core.ScalarFromString("AT_TIMESTAMP"),
			},
			Required: true,
		},
		"aws/lambda/function::aws.kinesis.lambda.startingPositionTimestamp": {
			Name:  "aws.kinesis.lambda.startingPositionTimestamp",
			Label: "Starting Position Timestamp",
			Type:  core.ScalarTypeString,
			Description: "The time from which to start reading, in Unix epoch seconds. " +
				"Used only when startingPosition is AT_TIMESTAMP.",
			Required: false,
		},
		"aws/lambda/function::aws.kinesis.lambda.parallelizationFactor": {
			Name:  "aws.kinesis.lambda.parallelizationFactor",
			Label: "Parallelization Factor",
			Type:  core.ScalarTypeInteger,
			Description: "The number of batches to process from each shard concurrently. " +
				"Default: 1. Range: 1 to 10.",
			DefaultValue: core.ScalarFromInt(1),
			Required:     false,
		},
		"aws/lambda/function::aws.kinesis.lambda.maximumRetryAttempts": {
			Name:  "aws.kinesis.lambda.maximumRetryAttempts",
			Label: "Maximum Retry Attempts",
			Type:  core.ScalarTypeInteger,
			Description: "The maximum number of retry attempts. " +
				"Default: -1 (infinite retries). Range: -1 to 10,000.",
			DefaultValue: core.ScalarFromInt(-1),
			Required:     false,
		},
		"aws/lambda/function::aws.kinesis.lambda.maximumRecordAgeInSeconds": {
			Name:  "aws.kinesis.lambda.maximumRecordAgeInSeconds",
			Label: "Maximum Record Age",
			Type:  core.ScalarTypeInteger,
			Description: "The maximum age (in seconds) of a record that Lambda sends to a function. " +
				"Default: -1 (infinite). Range: -1 to 604,800 (7 days).",
			DefaultValue: core.ScalarFromInt(-1),
			Required:     false,
		},
		"aws/lambda/function::aws.kinesis.lambda.bisectBatchOnFunctionError": {
			Name:  "aws.kinesis.lambda.bisectBatchOnFunctionError",
			Label: "Bisect Batch on Error",
			Type:  core.ScalarTypeBool,
			Description: "When true, splits the batch and retries on a smaller batch when a function returns an error. " +
				"Default: false.",
			DefaultValue: core.ScalarFromBool(false),
			Required:     false,
		},
		"aws/lambda/function::aws.kinesis.lambda.tumblingWindowInSeconds": {
			Name:  "aws.kinesis.lambda.tumblingWindowInSeconds",
			Label: "Tumbling Window",
			Type:  core.ScalarTypeInteger,
			Description: "The duration (in seconds) of a tumbling window for aggregating records before invoking the function. " +
				"0 disables tumbling windows. Default: 0. Range: 0 to 900.",
			DefaultValue: core.ScalarFromInt(0),
			Required:     false,
		},
		"aws/lambda/function::aws.kinesis.lambda.reportBatchItemFailures": {
			Name:  "aws.kinesis.lambda.reportBatchItemFailures",
			Label: "Report Batch Item Failures",
			Type:  core.ScalarTypeBool,
			Description: "When true, allows the function to report partial batch failures (sets the " +
				"ReportBatchItemFailures function response type) so processing resumes after the last " +
				"successful record. Default: false.",
			DefaultValue: core.ScalarFromBool(false),
			Required:     false,
		},
		"aws/lambda/function::aws.kinesis.lambda.filter.<index>": {
			Name:  "aws.kinesis.lambda.filter.<index>",
			Label: "Filter Pattern",
			Type:  core.ScalarTypeString,
			Description: "An event filter pattern controlling which records are sent to the function. " +
				"`<index>` is the zero-based position in the event source mapping's filter array; add multiple " +
				"filters with successive indices (0, 1, ...). The value is the AWS event-filtering pattern string " +
				"(e.g. `{\"data\":{\"eventType\":[\"created\"]}}`).",
			Required: false,
		},
		"aws/lambda/function::aws.kinesis.lambda.enabled": {
			Name:  "aws.kinesis.lambda.enabled",
			Label: "Enabled",
			Type:  core.ScalarTypeBool,
			Description: "When true, the event source mapping is active and Lambda will process records. " +
				"Default: true.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
	}
}
