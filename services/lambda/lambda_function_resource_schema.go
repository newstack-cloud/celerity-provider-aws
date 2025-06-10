package lambda

import (
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func lambdaFunctionResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "LambdaFunctionDefinition",
		Description: "The definition of an AWS Lambda function.",
		Required:    []string{"functionName", "code", "role"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"architecture": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The instruction set architecture that the function supports.",
				Default:     core.MappingNodeFromString("x86_64"),
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("x86_64"),
					core.MappingNodeFromString("arm64"),
				},
				Nullable: true,
			},
			"code": {
				Type:  provider.ResourceDefinitionsSchemaTypeObject,
				Label: "FunctionCode",
				Description: "The code for the Lambda function. You can either specify an object in Amazon S3," +
					" upload a .zip file archive deployment package directly, or specify the URI of a container image.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"imageUri": {
						Type:                 provider.ResourceDefinitionsSchemaTypeString,
						Description:          "The URI of the container image in the Amazon ECR registry.",
						FormattedDescription: "The URI of a [container image](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html) in the Amazon ECR registry.",
					},
					"s3Bucket": {
						Type: provider.ResourceDefinitionsSchemaTypeString,
						Description: "An Amazon S3 bucket in the same AWS Region as the function. " +
							"The bucket can be in a different AWS account.",
						// We can't do a negative lookbehind with Go's regexp engine, the regexp
						// for the bucket name in the official AWS API documentation for Lambda
						// includes a negative lookbehind to ensure that a bucket name that consists
						// of only a single period (".") is not allowed.
						// Due to this, bucket names that start with "." will fail validation
						// with the provider.
						Pattern:   "^[0-9A-Za-z\\-_][0-9A-Za-z\\.\\-_]+$",
						MinLength: 3,
						MaxLength: 63,
					},
					"s3Key": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The Amazon S3 key of the deployment package.",
						MinLength:   1,
						MaxLength:   1024,
					},
					"s3ObjectVersion": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "For versioned objects, the version of the deployment package object to use.",
						MinLength:   1,
						MaxLength:   1024,
					},
					"sourceKMSKeyArn": {
						Type: provider.ResourceDefinitionsSchemaTypeString,
						Description: "The ARN of the AWS Key Management Service (AWS KMS) customer managed key that's used to " +
							"encrypt your function's .zip deployment package. If you do not provide a custom managed key, " +
							"Lambda uses an AWS owned key.",
						FormattedDescription: "The ARN of the AWS Key Management Service (AWS KMS) customer managed key that's used to " +
							"encrypt your function's .zip deployment package. If you do not provide a custom managed key, " +
							"Lambda uses an [AWS owned key](https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#aws-owned-cmk).",
						Pattern: "(arn:(aws[a-zA-Z-]*)?:[a-z0-9-.]+:.*)|()",
					},
					"zipFile": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The inline code for the Lambda function. This will be converted into a base-64 encoded zip archive format by the provider.",
					},
				},
			},
			"codeSigningConfigArn": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "To enable code signing for this function, specify the ARN of code-signing configuration. " +
					"A code-signing configuration includes a set of signing profiles, which define the trusted publishers " +
					"for this function.",
				Pattern: "arn:(aws[a-zA-Z-]*)?:lambda:[a-z]{2}((-gov)|(-iso(b?)))?-[a-z]+-\\d{1}:\\d{12}:code-signing-config:csc-[a-z0-9]{17}",
			},
			"deadLetterConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Label:       "DeadLetterConfig",
				Description: "The dead-letter queue for failed asynchronous invocations.",
				FormattedDescription: "The [dead-letter queue](https://docs.aws.amazon.com/lambda/latest/dg/invocation-async-retain-records.html#invocation-dlq) " +
					"for failed asynchronous invocations.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"targetArn": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The Amazon Resource Name (ARN) of an Amazon SQS queue or Amazon SNS topic.",
						Pattern:     "(arn:(aws[a-zA-Z-]*)?:[a-z0-9-.]+:.*)|()",
					},
				},
			},
			"description": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "A description of the function.",
				MaxLength:   256,
			},
			"environment": {
				Type:  provider.ResourceDefinitionsSchemaTypeObject,
				Label: "Environment",
				Description: "A function's environment variable settings. You can use environment variables to adjust your " +
					"function's behavior without updating code. An environment variable is a pair of strings that are stored in a function's version-specific configuration.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"variables": {
						Type: provider.ResourceDefinitionsSchemaTypeMap,
						MapValues: &provider.ResourceDefinitionsSchema{
							Type: provider.ResourceDefinitionsSchemaTypeString,
						},
						Description: "Environment variable key-value pairs. Keys must follow the pattern \"[a-zA-Z]([a-zA-Z0-9_])+\".",
						FormattedDescription: "Environment variable key-value pairs. For more information, see " +
							"[Using Lambda environment variables](https://docs.aws.amazon.com/lambda/latest/dg/configuration-envvars.html). " +
							"Keys must follow the pattern `[a-zA-Z]([a-zA-Z0-9_])+`.",
					},
				},
			},
			"ephemeralStorage": {
				Type:  provider.ResourceDefinitionsSchemaTypeObject,
				Label: "EphemeralStorage",
				Description: "The size of the function's \"tmp\" directory in MB. The default value is 512," +
					" but can be any whole number between 512 and 10,240 MB.",
				FormattedDescription: "The size of the function's `tmp` directory in MB. The default value is 512," +
					" but can be any whole number between 512 and 10,240 MB. For more information, see " +
					"[Configuring ephemeral storage (console)](https://docs.aws.amazon.com/lambda/latest/dg/lambda-functions.html#configuration-ephemeral-storage)",
				Required: []string{"size"},
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"size": {
						Type: provider.ResourceDefinitionsSchemaTypeInteger,
						Description: "The size of the function's `tmp` directory in MB. " +
							"Can have a minimum value of 512 and a maximum value of 10240.",
						Minimum: core.ScalarFromInt(512),
						Maximum: core.ScalarFromInt(10240),
					},
				},
			},
			"fileSystemConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Label:       "FileSystemConfig",
				Description: "Details about the connection between a Lambda function and an Amazon EFS file system.",
				FormattedDescription: "Details about the connection between a Lambda function and " +
					"an [Amazon EFS file system](https://docs.aws.amazon.com/lambda/latest/dg/configuration-filesystem.html).",
				Required: []string{"arn", "localMountPath"},
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"arn": {
						Type: provider.ResourceDefinitionsSchemaTypeString,
						Description: "The Amazon Resource Name (ARN) of the Amazon EFS access" +
							" point that provides access to the file system.",
						Pattern:   "arn:aws[a-zA-Z-]*:elasticfilesystem:[a-z]{2}((-gov)|(-iso(b?)))?-[a-z]+-\\d{1}:\\d{12}:access-point/fsap-[a-f0-9]{17}",
						MaxLength: 200,
					},
					"localMountPath": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The path where the function can access the file system, starting with /mnt/.",
						Pattern:     "^/mnt/[a-zA-Z0-9-_.]+$",
						MaxLength:   160,
					},
				},
			},
			"functionName": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The name of the Lambda function stored in the AWS system.",
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("MyFunction"),
				},
				MinLength:    1,
				MaxLength:    64,
				MustRecreate: true,
			},
			"handler": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The name of the method within your code that Lambda calls to execute your function." +
					" This is required if the deployment package is a .zip file archive. The format includes the file name." +
					" It can also include namespaces and other qualifiers, depending on the runtime.",
				FormattedDescription: "The name of the method within your code that Lambda calls to execute your function." +
					" This is required if the deployment package is a .zip file archive. The format includes the file name." +
					" It can also include namespaces and other qualifiers, depending on the runtime." +
					" For more information, see [Lambda programming model](https://docs.aws.amazon.com/lambda/latest/dg/foundation-progmodel.html)",
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("index.handler"),
					core.MappingNodeFromString("lambda_function.lambda_handler"),
				},
				MaxLength: 128,
				Pattern:   "^[^\\s]+$",
			},
			"imageConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Label:       "ImageConfig",
				Description: "Configuration values that override the container image Dockerfile settings.",
				FormattedDescription: "Configuration values that override the container image Dockerfile settings. " +
					"For more information, see [Container image settings](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html#images-parms)",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"command": {
						Type:        provider.ResourceDefinitionsSchemaTypeArray,
						Description: "Specifies the parameters that you want to pass in with ENTRYPOINT. You can specify a maximum of 1,500 parameters in the list.",
						Items: &provider.ResourceDefinitionsSchema{
							Type: provider.ResourceDefinitionsSchemaTypeString,
						},
						MaxLength: 1500,
					},
					"entryPoint": {
						Type: provider.ResourceDefinitionsSchemaTypeArray,
						Description: "Specifies the entry point to the application, which is typically the location of the runtime executable. " +
							"You can specify a maximum of 1,500 entries in the list.",
						Items: &provider.ResourceDefinitionsSchema{
							Type: provider.ResourceDefinitionsSchemaTypeString,
						},
						MaxLength: 1500,
					},
					"workingDirectory": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "Specifies the working directory. The length of the directory string cannot exceed 1,000 characters.",
						MaxLength:   1000,
					},
				},
			},
			"kmsKeyArn": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The ARN of the AWS Key Management Service (AWS KMS) customer managed key that's used to " +
					"encrypt your function's environment variables.",
				FormattedDescription: "The ARN of the AWS Key Management Service (AWS KMS) customer managed key that's used to " +
					"encrypt your function's environment variables. When this configuration is not provided, AWS Lambda uses " +
					"a default service key.",
				Pattern: "^(arn:(aws[a-zA-Z-]*)?:[a-z0-9-.]+:.*)|()$",
			},
			"layers": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of function layers to add to the function's execution environment.",
				FormattedDescription: "A list of [function layers](https://docs.aws.amazon.com/lambda/latest/dg/configuration-layers.html) " +
					"to add to the function's execution environment. Specify each layer by its ARN, including the version.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeString,
					Description: "The ARN of a layer version.",
				},
			},
			"loggingConfig": {
				Type:                 provider.ResourceDefinitionsSchemaTypeObject,
				Label:                "LoggingConfig",
				Description:          "The function's Amazon CloudWatch logging configuration.",
				FormattedDescription: "The function's [Amazon CloudWatch logging configuration](https://docs.aws.amazon.com/lambda/latest/dg/monitoring-cloudwatchlogs.html).",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"applicationLogLevel": {
						Type: provider.ResourceDefinitionsSchemaTypeString,
						Description: "A property to filter the application logs for your function that Lambda sends to CloudWatch. " +
							"Lambda only sends application logs at the selected level of detail and lower, where TRACE is the highest level and FATAL is the lowest.",
						FormattedDescription: "A property to filter the application logs for your function that Lambda sends to CloudWatch. " +
							"Lambda only sends application logs at the selected level of detail and lower, where `TRACE` is the highest level and `FATAL` is the lowest.",
						AllowedValues: []*core.MappingNode{
							core.MappingNodeFromString("TRACE"),
							core.MappingNodeFromString("DEBUG"),
							core.MappingNodeFromString("INFO"),
							core.MappingNodeFromString("WARN"),
							core.MappingNodeFromString("ERROR"),
							core.MappingNodeFromString("FATAL"),
						},
					},
					"logFormat": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The format of the log records that the function sends to CloudWatch Logs.",
						AllowedValues: []*core.MappingNode{
							core.MappingNodeFromString("JSON"),
							core.MappingNodeFromString("Text"),
						},
					},
					"logGroup": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The name of the CloudWatch Logs group the function sends logs to.",
						Pattern:     "[\\.\\-_/#A-Za-z0-9]+",
					},
					"systemLogLevel": {
						Type: provider.ResourceDefinitionsSchemaTypeString,
						Description: "A property to filter the system logs for your function that Lambda sends to CloudWatch." +
							"Lambda only sends system logs at the selected level of detail and lower, where DEBUG is the highest level and WARN is the lowest.",
						FormattedDescription: "A property to filter the system logs for your function that Lambda sends to CloudWatch." +
							"Lambda only sends system logs at the selected level of detail and lower, where `DEBUG` is the highest level and `WARN` is the lowest.",
						AllowedValues: []*core.MappingNode{
							core.MappingNodeFromString("DEBUG"),
							core.MappingNodeFromString("INFO"),
							core.MappingNodeFromString("WARN"),
						},
					},
				},
			},
			"memorySize": {
				Type: provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The amount of memory available to the function at runtime. " +
					"Increasing the function memory also increases its CPU allocation. " +
					"The default value is 128 MB. The value can be any multiple of 1 MB. " +
					"Note that new AWS accounts have reduced concurrency and memory quotas. " +
					"AWS raises these quotas automatically based on your usage. " +
					"You can also request a quota increase.",
				FormattedDescription: "The amount of [memory available to the function](https://docs.aws.amazon.com/lambda/latest/dg/configuration-function-common.html#configuration-memory-console) at runtime. " +
					"Increasing the function memory also increases its CPU allocation. " +
					"The default value is 128 MB. The value can be any multiple of 1 MB. " +
					"Note that new AWS accounts have reduced concurrency and memory quotas. " +
					"AWS raises these quotas automatically based on your usage. " +
					"You can also request a quota increase.",
				Default: core.MappingNodeFromInt(128),
				Examples: []*core.MappingNode{
					core.MappingNodeFromInt(128),
					core.MappingNodeFromInt(512),
					core.MappingNodeFromInt(1024),
				},
			},
			"packageType": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The type of deployment package.",
				Default:     core.MappingNodeFromString("Zip"),
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("Zip"),
					core.MappingNodeFromString("Image"),
				},
				MustRecreate: true,
			},
			"recursiveLoop": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The status of your function's recursive loop detection configuration.\n\n" +
					"When this value is set to Allow and Lambda detcts your function being invoked as part of a recursive loop, it doesn't take any action.\n\n" +
					"When this value is set to Terminate and Lambda detects your function being invoked as part of a recursive loop, it stops your function being invoked and notifies you.",
				FormattedDescription: "The status of your function's recursive loop detection configuration.\n\n" +
					"When this value is set to `Allow` and Lambda detcts your function being invoked as part of a recursive loop, it doesn't take any action.\n\n" +
					"When this value is set to `Terminate` and Lambda detects your function being invoked as part of a recursive loop, it stops your function being invoked and notifies you.",
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("Allow"),
					core.MappingNodeFromString("Terminate"),
				},
			},
			"reservedConcurrentExecutions": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The number of simultaneous executions to reserve for the function.",
				Minimum:     core.ScalarFromInt(0),
			},
			"role": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the function's execution role that grants the function " +
					"permission to access AWS services and resources.",
				Pattern: "^arn:(aws[a-zA-Z-]*)?:iam::\\d{12}:role/?[a-zA-Z_0-9+=,.@\\-_/]+$",
			},
			"runtime": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The identifier of the function's runtime. " +
					"Runtime is required if the deployment package is a .zip file archive. Specifying a runtime results in an error " +
					"if you're deploying a function using a container image.",
				FormattedDescription: "The identifier of the function's [runtime](https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html). " +
					"Runtime is required if the deployment package is a .zip file archive. Specifying a runtime results in an error " +
					"if you're deploying a function using a container image.\n\n" +
					"The following list includes deprecated runtimes. Lambda blocks creating new functions and updating existing functions " +
					"shortly after each runtime is deprecated. For more information, see [Runtime use after deprecation](https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html#runtime-deprecation-levels).\n\n" +
					"For a list of all currently supported runtimes, see [Supported runtimes](https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html#runtimes-supported)",
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("nodejs"),
					core.MappingNodeFromString("nodejs4.3"),
					core.MappingNodeFromString("nodejs4.3-edge"),
					core.MappingNodeFromString("nodejs6.10"),
					core.MappingNodeFromString("nodejs8.10"),
					core.MappingNodeFromString("nodejs10.x"),
					core.MappingNodeFromString("nodejs12.x"),
					core.MappingNodeFromString("nodejs14.x"),
					core.MappingNodeFromString("nodejs16.x"),
					core.MappingNodeFromString("nodejs18.x"),
					core.MappingNodeFromString("nodejs20.x"),
					core.MappingNodeFromString("nodejs22.x"),
					core.MappingNodeFromString("java8"),
					core.MappingNodeFromString("java8.al2"),
					core.MappingNodeFromString("java11"),
					core.MappingNodeFromString("java17"),
					core.MappingNodeFromString("java21"),
					core.MappingNodeFromString("python2.7"),
					core.MappingNodeFromString("python3.6"),
					core.MappingNodeFromString("python3.7"),
					core.MappingNodeFromString("python3.8"),
					core.MappingNodeFromString("python3.9"),
					core.MappingNodeFromString("python3.10"),
					core.MappingNodeFromString("python3.11"),
					core.MappingNodeFromString("python3.12"),
					core.MappingNodeFromString("python3.13"),
					core.MappingNodeFromString("dotnetcore1.0"),
					core.MappingNodeFromString("dotnetcore2.0"),
					core.MappingNodeFromString("dotnetcore2.1"),
					core.MappingNodeFromString("dotnetcore3.1"),
					core.MappingNodeFromString("dotnet6"),
					core.MappingNodeFromString("dotnet7"),
					core.MappingNodeFromString("dotnet8"),
					core.MappingNodeFromString("go1.x"),
					core.MappingNodeFromString("ruby2.5"),
					core.MappingNodeFromString("ruby2.7"),
					core.MappingNodeFromString("ruby3.2"),
					core.MappingNodeFromString("ruby3.3"),
					core.MappingNodeFromString("ruby3.4"),
					core.MappingNodeFromString("provided"),
					core.MappingNodeFromString("provided.al2"),
					core.MappingNodeFromString("provided.al2023"),
				},
			},
			"runtimeManagementConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Label:       "RuntimeManagementConfig",
				Required:    []string{"updateRuntimeOn"},
				Description: "Sets the runtime management configuration for a function's version.",
				FormattedDescription: "Sets the runtime management configuration for a function's version. " +
					"For more information, see [Runtime updates](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-update.html).",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"runtimeVersionArn": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The ARN of the runtime version you want the function to use.",
						FormattedDescription: "The ARN of the runtime version you want the function to use.\n\n" +
							"> [!NOTE]\n" +
							"> This is only required if you're using the **Manual** runtime update mode.",
						Pattern:   "^arn:(aws[a-zA-Z-]*):lambda:[a-z]{2}((-gov)|(-iso(b?)))?-[a-z]+-\\d{1}::runtime:.+$",
						MinLength: 26,
						MaxLength: 2048,
					},
					"updateRuntimeOn": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The runtime update mode to use.",
						FormattedDescription: "The runtime update mode to use.\n\n" +
							"- **Auto (default)** - Automatically update to the most recent and secure runtime " +
							"version using a [Two-phase runtime version rollout](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-update.html#runtime-management-two-phase)." +
							" This is the best choice for most cases as it ensures you will always benefit from runtime updates.\n" +
							"- **FunctionUpdate** - Lambda updates the runtime of your function to the most recent and secure runtime version " +
							"when you update your function. This approach synchronizes runtime updates with function deployments, " +
							"giving you control over when runtime updates are applied and allowing you to detect and mitigate rare runtime update incompatibilities early. " +
							"When using this setting, you need to regularly update your functions to keep their runtime up-to-date.\n" +
							"- **Manual** - You specify a runtime version in your function configuration. " +
							"The function will use this runtime version indefinitely. In the rare case where a runtime version is incompatible with an existing function, " +
							"this allows you to roll back your function to an earlier runtime version. " +
							"For more information, see [Roll back a runtime version](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-update.html#runtime-management-rollback).",
						AllowedValues: []*core.MappingNode{
							core.MappingNodeFromString("Auto"),
							core.MappingNodeFromString("FunctionUpdate"),
							core.MappingNodeFromString("Manual"),
						},
					},
				},
			},
			"snapStart": {
				Type:                 provider.ResourceDefinitionsSchemaTypeObject,
				Label:                "SnapStart",
				Description:          "The function's AWS Lambda SnapStart setting.",
				FormattedDescription: "The function's [AWS Lambda SnapStart](https://docs.aws.amazon.com/lambda/latest/dg/snapstart.html) setting.",
				Required:             []string{"applyOn"},
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"applyOn": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "Set ApplyOn to PublishedVersions to create a snapshot of the initialized execution environment when you publish a function version.",
						AllowedValues: []*core.MappingNode{
							core.MappingNodeFromString("PublishedVersions"),
							core.MappingNodeFromString("None"),
						},
					},
				},
			},
			"tags": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of tags to apply to the function.",
				FormattedDescription: "A list of [tags](https://docs.aws.amazon.com/lambda/latest/dg/tagging.html) " +
					"to apply to the function.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:                 provider.ResourceDefinitionsSchemaTypeObject,
					Label:                "Tag",
					Description:          "A tag to apply to the function.",
					FormattedDescription: "A [tag](https://docs.aws.amazon.com/lambda/latest/dg/tagging.html) to apply to the function.",
					Required:             []string{"key", "value"},
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"key": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The key of the tag.",
							MinLength:   1,
							MaxLength:   128,
						},
						"value": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The value of the tag.",
							MinLength:   0,
							MaxLength:   256,
						},
					},
				},
			},
			"timeout": {
				Type: provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The amount of time (in seconds) that Lambda allows a function to run before stopping it. " +
					"The default is 3 seconds. The maximum allowed value is 900 seconds.",
				FormattedDescription: "The amount of time (in seconds) that Lambda allows a function to run before stopping it. " +
					"The default is 3 seconds. The maximum allowed value is 900 seconds. " +
					"For more information, see [Lambda execution environment](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-context.html).",
				Default: core.MappingNodeFromInt(3),
				Minimum: core.ScalarFromInt(1),
				Examples: []*core.MappingNode{
					core.MappingNodeFromInt(3),
					core.MappingNodeFromInt(30),
					core.MappingNodeFromInt(900),
				},
			},
			"tracingConfig": {
				Type:                 provider.ResourceDefinitionsSchemaTypeObject,
				Label:                "TracingConfig",
				Description:          "The function's AWS X-Ray tracing configuration.",
				FormattedDescription: "The function's [AWS X-Ray tracing](https://docs.aws.amazon.com/lambda/latest/dg/services-xray.html) configuration.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"mode": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The tracing mode.",
						AllowedValues: []*core.MappingNode{
							core.MappingNodeFromString("Active"),
							core.MappingNodeFromString("PassThrough"),
						},
					},
				},
			},
			"vpcConfig": {
				Type:                 provider.ResourceDefinitionsSchemaTypeObject,
				Label:                "VpcConfig",
				Description:          "The VPC configuration for the function.",
				FormattedDescription: "The [VPC configuration](https://docs.aws.amazon.com/lambda/latest/dg/configuration-vpc.html) for the function.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"securityGroupIds": {
						Type:        provider.ResourceDefinitionsSchemaTypeArray,
						Description: "A list of VPC security group IDs.",
						Items: &provider.ResourceDefinitionsSchema{
							Type: provider.ResourceDefinitionsSchemaTypeString,
						},
						MaxLength: 5,
					},
					"subnetIds": {
						Type:        provider.ResourceDefinitionsSchemaTypeArray,
						Description: "A list of VPC subnet IDs.",
						Items: &provider.ResourceDefinitionsSchema{
							Type: provider.ResourceDefinitionsSchemaTypeString,
						},
						MaxLength: 16,
					},
					"ipv6AllowedForDualStack": {
						Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
						Description: "Allows outbound IPv6 traffic on VPC functions that are connected to dual-stack subnets.",
					},
				},
			},

			// Computed fields
			"arn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the Lambda function.",
				Computed:    true,
			},
			"snapStartResponseApplyOn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "When SnapStart is set to PublishedVersions, this field indicates the apply setting.",
				Computed:    true,
			},
			"snapStartResponseOptimizationStatus": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The status of the SnapStart optimization.",
				Computed:    true,
			},
		},
	}
}
