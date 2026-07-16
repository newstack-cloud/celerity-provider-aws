package lambdadynamodb

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func dynamoDBTableLambdaFunctionLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		// resourceA (DynamoDB table) annotation - configures the table's stream format
		// This stays on resourceA because it affects the table regardless of which functions consume it
		"aws/dynamodb/table::aws.dynamodb.stream.viewType": {
			Name:  "aws.dynamodb.stream.viewType",
			Label: "Stream View Type",
			Type:  core.ScalarTypeString,
			Description: "The information to include in stream records. " +
				"KEYS_ONLY: Only key attributes. " +
				"NEW_IMAGE: Entire item after modification. " +
				"OLD_IMAGE: Entire item before modification. " +
				"NEW_AND_OLD_IMAGES: Both old and new images. " +
				"Default: NEW_AND_OLD_IMAGES.",
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("KEYS_ONLY"),
				core.ScalarFromString("NEW_IMAGE"),
				core.ScalarFromString("OLD_IMAGE"),
				core.ScalarFromString("NEW_AND_OLD_IMAGES"),
			},
			DefaultValue: core.ScalarFromString("NEW_AND_OLD_IMAGES"),
			Required:     false,
		},
		// resourceB (Lambda function) annotations - configure the event source mapping for this specific function
		// These are on resourceB because a table can trigger multiple functions with different configurations
		// Annotation names include both services and feature (dynamodb.lambda.stream) to indicate relationship configuration
		"aws/lambda/function::aws.dynamodb.lambda.stream.batchSize": {
			Name:  "aws.dynamodb.lambda.stream.batchSize",
			Label: "Batch Size",
			Type:  core.ScalarTypeInteger,
			Description: "The maximum number of records to retrieve per batch. " +
				"Default: 100. Range: 1 to 10,000.",
			DefaultValue: core.ScalarFromInt(100),
			Required:     false,
		},
		"aws/lambda/function::aws.dynamodb.lambda.stream.batchWindow": {
			Name:  "aws.dynamodb.lambda.stream.batchWindow",
			Label: "Batch Window",
			Type:  core.ScalarTypeInteger,
			Description: "The maximum amount of time (in seconds) to gather records before invoking the function. " +
				"Default: 0. Range: 0 to 300.",
			DefaultValue: core.ScalarFromInt(0),
			Required:     false,
		},
		"aws/lambda/function::aws.dynamodb.lambda.stream.startingPosition": {
			Name:  "aws.dynamodb.lambda.stream.startingPosition",
			Label: "Starting Position",
			Type:  core.ScalarTypeString,
			Description: "The position in the stream where AWS Lambda should start reading. " +
				"TRIM_HORIZON starts from the oldest record. LATEST starts from the newest record.",
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("TRIM_HORIZON"),
				core.ScalarFromString("LATEST"),
			},
			Required: true,
		},
		"aws/lambda/function::aws.dynamodb.lambda.stream.parallelizationFactor": {
			Name:  "aws.dynamodb.lambda.stream.parallelizationFactor",
			Label: "Parallelization Factor",
			Type:  core.ScalarTypeInteger,
			Description: "The number of batches to process from each shard concurrently. " +
				"Default: 1. Range: 1 to 10.",
			DefaultValue: core.ScalarFromInt(1),
			Required:     false,
		},
		"aws/lambda/function::aws.dynamodb.lambda.stream.maximumRetryAttempts": {
			Name:  "aws.dynamodb.lambda.stream.maximumRetryAttempts",
			Label: "Maximum Retry Attempts",
			Type:  core.ScalarTypeInteger,
			Description: "The maximum number of retry attempts. " +
				"Default: -1 (infinite retries). Range: -1 to 10,000.",
			DefaultValue: core.ScalarFromInt(-1),
			Required:     false,
		},
		"aws/lambda/function::aws.dynamodb.lambda.stream.maximumRecordAgeInSeconds": {
			Name:  "aws.dynamodb.lambda.stream.maximumRecordAgeInSeconds",
			Label: "Maximum Record Age",
			Type:  core.ScalarTypeInteger,
			Description: "The maximum age (in seconds) of a record that Lambda sends to a function. " +
				"Default: -1 (infinite). Range: -1 to 604,800 (7 days).",
			DefaultValue: core.ScalarFromInt(-1),
			Required:     false,
		},
		"aws/lambda/function::aws.dynamodb.lambda.stream.bisectBatchOnFunctionError": {
			Name:  "aws.dynamodb.lambda.stream.bisectBatchOnFunctionError",
			Label: "Bisect Batch on Error",
			Type:  core.ScalarTypeBool,
			Description: "When true, splits the batch and retries on a smaller batch when a function returns an error. " +
				"Default: false.",
			DefaultValue: core.ScalarFromBool(false),
			Required:     false,
		},
		"aws/lambda/function::aws.dynamodb.lambda.stream.reportBatchItemFailures": {
			Name:  "aws.dynamodb.lambda.stream.reportBatchItemFailures",
			Label: "Report Batch Item Failures",
			Type:  core.ScalarTypeBool,
			Description: "When true, allows the function to report partial batch failures (sets the " +
				"ReportBatchItemFailures function response type) so only failed records are retried. Default: false.",
			DefaultValue: core.ScalarFromBool(false),
			Required:     false,
		},
		"aws/lambda/function::aws.dynamodb.lambda.stream.filter.<index>": {
			Name:  "aws.dynamodb.lambda.stream.filter.<index>",
			Label: "Filter Pattern",
			Type:  core.ScalarTypeString,
			Description: "An event filter pattern controlling which stream records are sent to the function. " +
				"`<index>` is the zero-based position in the event source mapping's filter array; add multiple " +
				"filters with successive indices (0, 1, ...). The value is the AWS event-filtering pattern string " +
				"(e.g. `{\"eventName\":[\"INSERT\"]}`).",
			Required: false,
		},
		"aws/lambda/function::aws.dynamodb.lambda.stream.enabled": {
			Name:  "aws.dynamodb.lambda.stream.enabled",
			Label: "Enabled",
			Type:  core.ScalarTypeBool,
			Description: "When true, the event source mapping is active and Lambda will process records. " +
				"Default: true.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
	}
}
