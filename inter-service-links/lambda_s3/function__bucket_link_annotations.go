package lambdas3

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func functionBucketLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	return map[string]*provider.LinkAnnotationDefinition{
		"aws/lambda/function::aws.lambda.s3.populateEnvVars": {
			Name:  "aws.lambda.s3.populateEnvVars",
			Label: "Populate Environment Variables",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from lambda function to reference the target S3 bucket. " +
				"This will populate environment variables for all target S3 buckets that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.s3.<targetBucket>.populateEnvVars": {
			Name:  "aws.lambda.s3.<targetBucket>.populateEnvVars",
			Label: "Populate Environment Variables for Specific Target Bucket",
			Type:  core.ScalarTypeBool,
			Description: "A boolean flag to determine whether or not environment variables should be populated " +
				"in the linked from lambda function for a specific target S3 bucket. " +
				"The default behaviour is to populate environment variables for all target buckets that match the selector.",
			DefaultValue: core.ScalarFromBool(true),
			Required:     false,
		},
		"aws/lambda/function::aws.lambda.s3.<targetBucket>.envVarName": {
			Name:  "aws.lambda.s3.<targetBucket>.envVarName",
			Label: "Environment Variable Name",
			Type:  core.ScalarTypeString,
			Description: "The name of the environment variable to populate in the linked from lambda function " +
				"to reference the target S3 bucket. " +
				"The default format for bucket name environment variables is `S3_BUCKET_<targetBucket>`.",
			Examples: []*core.ScalarValue{
				core.ScalarFromString("UPLOADS_BUCKET_NAME"),
				core.ScalarFromString("ASSETS_BUCKET"),
			},
			Required: false,
		},
		"aws/lambda/function::aws.lambda.s3.<targetBucket>.accessLevel": {
			Name:  "aws.lambda.s3.<targetBucket>.accessLevel",
			Label: "Access Level",
			Type:  core.ScalarTypeString,
			Description: "The level of access to grant to the S3 bucket. " +
				"Valid values are 'read', 'write', or 'readwrite' (default).",
			DefaultValue: core.ScalarFromString("readwrite"),
			AllowedValues: []*core.ScalarValue{
				core.ScalarFromString("read"),
				core.ScalarFromString("write"),
				core.ScalarFromString("readwrite"),
			},
			Required: false,
		},
	}
}
