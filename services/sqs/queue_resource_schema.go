package sqs

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func sqsQueueResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "SQSQueueDefinition",
		Description: "The definition of an AWS SQS queue.",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"queueName": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The name of the SQS queue.",
				FormattedDescription: "The name of the SQS queue. " +
					"Queue names must be unique within a region and can contain alphanumeric characters, hyphens, and underscores. " +
					"If not specified, a unique name will be generated.",
				Pattern:      `[a-zA-Z0-9_-]+`,
				MinLength:    1,
				MaxLength:    80,
				MustRecreate: true,
				Nullable:     true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("my-queue"),
					core.MappingNodeFromString("processing-queue"),
					core.MappingNodeFromString("dead-letter-queue"),
				},
			},
			"delaySeconds": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The number of seconds to delay the delivery of all messages in the queue.",
				FormattedDescription: "The number of seconds to delay the delivery of all messages in the queue. " +
					"Valid values: 0 to 900 seconds (15 minutes). Default is 0.",
				Minimum:  core.ScalarFromInt(0),
				Maximum:  core.ScalarFromInt(900),
				Default:  core.MappingNodeFromInt(0),
				Nullable: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromInt(0),
					core.MappingNodeFromInt(30),
					core.MappingNodeFromInt(300),
				},
			},
			"redrivePolicy": {
				Type:                 provider.ResourceDefinitionsSchemaTypeObject,
				Description:          "The redrive policy for the queue.",
				FormattedDescription: "The redrive policy that specifies the source queue, the dead-letter queue, and the conditions under which Amazon SQS moves messages from the former to the latter.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"deadLetterTargetArn": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The Amazon Resource Name (ARN) of the dead-letter queue to which Amazon SQS moves messages after the value of maxReceiveCount is exceeded.",
						FormattedDescription: "The ARN of the dead-letter queue where failed messages will be sent. " +
							"The dead letter queue must already exist and be in the same region.",
						Pattern: `^arn:aws:sqs:[a-z0-9-]+:[0-9]{12}:[a-zA-Z0-9_-]+(\.fifo)?$`,
						Examples: []*core.MappingNode{
							core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:my-dlq"),
							core.MappingNodeFromString("arn:aws:sqs:us-east-1:123456789012:failed-messages.fifo"),
						},
					},
					"maxReceiveCount": {
						Type:        provider.ResourceDefinitionsSchemaTypeInteger,
						Description: "The number of times a message is delivered to the source queue before being moved to the dead-letter queue.",
						FormattedDescription: "The number of times a message is delivered to the source queue before being moved to the dead-letter queue. " +
							"Valid values: 1 to 1000. Default is 10.",
						Minimum: core.ScalarFromInt(1),
						Maximum: core.ScalarFromInt(1000),
						Default: core.MappingNodeFromInt(10),
						Examples: []*core.MappingNode{
							core.MappingNodeFromInt(3),
							core.MappingNodeFromInt(10),
							core.MappingNodeFromInt(50),
						},
					},
				},
				Nullable: true,
			},
			"redriveAllowPolicy": {
				Type:                 provider.ResourceDefinitionsSchemaTypeObject,
				Description:          "The redrive allow policy for the queue.",
				FormattedDescription: "The redrive allow policy that specifies which source queues can use this queue as a dead-letter queue.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"redrivePermission": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The redrive permission for the queue.",
						FormattedDescription: "The redrive permission that specifies which source queues can use this queue as a dead-letter queue. " +
							"Valid values are 'allowAll' and 'denyAll'.",
						AllowedValues: []*core.MappingNode{
							core.MappingNodeFromString("allowAll"),
							core.MappingNodeFromString("denyAll"),
						},
						Examples: []*core.MappingNode{
							core.MappingNodeFromString("allowAll"),
							core.MappingNodeFromString("denyAll"),
						},
					},
					"sourceQueueArns": {
						Type:        provider.ResourceDefinitionsSchemaTypeArray,
						Description: "The ARNs of the source queues that can use this queue as a dead-letter queue.",
						FormattedDescription: "The ARNs of the source queues that can use this queue as a dead-letter queue. " +
							"This is only applicable when redrivePermission is 'denyAll' and you want to specify specific source queues.",
						Items: &provider.ResourceDefinitionsSchema{
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The ARN of a source queue.",
							Pattern:     `^arn:aws:sqs:[a-z0-9-]+:[0-9]{12}:[a-zA-Z0-9_-]+(\.fifo)?$`,
							Examples: []*core.MappingNode{
								core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:source-queue"),
								core.MappingNodeFromString("arn:aws:sqs:us-east-1:123456789012:processing-queue.fifo"),
							},
						},
						Nullable: true,
					},
				},
				Nullable: true,
			},
			"messageRetentionPeriod": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The number of seconds to retain a message in the queue.",
				FormattedDescription: "The number of seconds to retain a message in the queue. " +
					"Valid values: 60 seconds (1 minute) to 1,209,600 seconds (14 days). Default is 345,600 seconds (4 days).",
				Minimum:  core.ScalarFromInt(60),
				Maximum:  core.ScalarFromInt(1209600),
				Default:  core.MappingNodeFromInt(345600),
				Nullable: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromInt(3600),   // 1 hour
					core.MappingNodeFromInt(86400),  // 1 day
					core.MappingNodeFromInt(345600), // 4 days
				},
			},
			"receiveMessageWaitTimeSeconds": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The number of seconds to wait for a message to arrive in the queue before returning.",
				FormattedDescription: "The number of seconds to wait for a message to arrive in the queue before returning. " +
					"Valid values: 0 to 20 seconds. Default is 0 (short polling).",
				Minimum:  core.ScalarFromInt(0),
				Maximum:  core.ScalarFromInt(20),
				Default:  core.MappingNodeFromInt(0),
				Nullable: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromInt(0),
					core.MappingNodeFromInt(5),
					core.MappingNodeFromInt(20),
				},
			},
			"visibilityTimeout": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The number of seconds during which a message is hidden from other consumers after being received.",
				FormattedDescription: "The number of seconds during which a message is hidden from other consumers after being received. " +
					"Valid values: 0 to 43,200 seconds (12 hours). Default is 30 seconds.",
				Minimum:  core.ScalarFromInt(0),
				Maximum:  core.ScalarFromInt(43200),
				Default:  core.MappingNodeFromInt(30),
				Nullable: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromInt(30),   // 30 seconds
					core.MappingNodeFromInt(300),  // 5 minutes
					core.MappingNodeFromInt(3600), // 1 hour
				},
			},
			"fifoQueue": {
				Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
				Description: "Whether this is a FIFO (First-In-First-Out) queue.",
				FormattedDescription: "Whether this is a FIFO (First-In-First-Out) queue. " +
					"FIFO queues provide exactly-once processing and guarantee message ordering. " +
					"FIFO queue names must end with '.fifo'.",
				Default:  core.MappingNodeFromBool(false),
				Nullable: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromBool(false),
					core.MappingNodeFromBool(true),
				},
			},
			"contentBasedDeduplication": {
				Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
				Description: "Whether to enable content-based deduplication for FIFO queues.",
				FormattedDescription: "Whether to enable content-based deduplication for FIFO queues. " +
					"This feature automatically deduplicates messages with identical content. " +
					"Only applicable to FIFO queues.",
				Default:  core.MappingNodeFromBool(false),
				Nullable: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromBool(false),
					core.MappingNodeFromBool(true),
				},
			},
			"deduplicationScope": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The deduplication scope for FIFO queues.",
				FormattedDescription: "The deduplication scope for FIFO queues. " +
					"Valid values are 'messageGroup' and 'queue'. " +
					"Only applicable to FIFO queues.",
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("messageGroup"),
					core.MappingNodeFromString("queue"),
				},
				Nullable: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("messageGroup"),
					core.MappingNodeFromString("queue"),
				},
			},
			"fifoThroughputLimit": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "Specifies whether the FIFO queue throughput quota applies to the " +
					"entire queue or per message group. Valid values are 'perQueue' and 'perMessageGroupId'.",
				FormattedDescription: "Specifies whether the FIFO queue throughput quota applies to the " +
					"entire queue or per message group. Valid values are `perQueue` and `perMessageGroupId`.\n" +
					"If you set these attributes to anything other than these values, normal throughput is in effect and " +
					"deduplication occurs as specified. For more information see [High throughput for FIFO queues](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/high-throughput-fifo.html) and " +
					"[Quotas related to messages](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/quotas-messages.html) " +
					"in the _Amazon SQS Developer Guide_.",
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("perMessageGroupId"),
					core.MappingNodeFromString("perQueue"),
				},
				Nullable: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("perMessageGroupId"),
					core.MappingNodeFromString("perQueue"),
				},
			},
			"maximumMessageSize": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The maximum message size in bytes.",
				FormattedDescription: "The maximum message size in bytes. " +
					"Valid values: 1,024 bytes (1 KB) to 262,144 bytes (256 KB). Default is 262,144 bytes (256 KB).",
				Minimum:  core.ScalarFromInt(1024),
				Maximum:  core.ScalarFromInt(262144),
				Default:  core.MappingNodeFromInt(262144),
				Nullable: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromInt(1024),   // 1 KB
					core.MappingNodeFromInt(65536),  // 64 KB
					core.MappingNodeFromInt(262144), // 256 KB
				},
			},
			"kmsMasterKeyId": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The ID of an AWS-managed customer master key (CMK) for Amazon SQS or a custom CMK.",
				FormattedDescription: "The ID of an AWS-managed customer master key (CMK) for Amazon SQS or a custom CMK. " +
					"For more information, see Key Terms. For more information about AWS-managed CMKs, see the AWS Key Management Service Developer Guide.",
				Pattern:  `^(alias/[a-zA-Z0-9/_-]+|arn:aws:kms:[a-z0-9-]+:[0-9]{12}:(key/[a-f0-9-]{36}|alias/[a-zA-Z0-9/_-]+))$`,
				Nullable: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("alias/aws/sqs"),
					core.MappingNodeFromString("arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012"),
					core.MappingNodeFromString("arn:aws:kms:us-west-2:123456789012:alias/my-key"),
				},
			},
			"kmsDataKeyReusePeriodSeconds": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The number of seconds for which Amazon SQS can reuse a data key to encrypt or decrypt messages before calling AWS KMS again.",
				FormattedDescription: "The number of seconds for which Amazon SQS can reuse a data key to encrypt or decrypt messages before calling AWS KMS again. " +
					"Valid values: 60 seconds (1 minute) to 86,400 seconds (24 hours). Default is 300 seconds (5 minutes).",
				Minimum:  core.ScalarFromInt(60),
				Maximum:  core.ScalarFromInt(86400),
				Default:  core.MappingNodeFromInt(300),
				Nullable: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromInt(60),   // 1 minute
					core.MappingNodeFromInt(300),  // 5 minutes
					core.MappingNodeFromInt(3600), // 1 hour
				},
			},
			"sqsManagedSseEnabled": {
				Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
				Description: "Whether to enable server-side encryption using SQS-owned encryption keys.",
				FormattedDescription: "Whether to enable server-side encryption using SQS-owned encryption keys. " +
					"When enabled, SQS uses an AWS managed customer master key (CMK).",
				Default:  core.MappingNodeFromBool(false),
				Nullable: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromBool(false),
					core.MappingNodeFromBool(true),
				},
			},
			"policy": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "The resource-based policy document for the SQS queue.",
				FormattedDescription: "The resource-based policy document that defines who can access the queue and what actions they can perform. " +
					"This follows the standard IAM policy document format.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"Version": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The policy language version.",
						Default:     core.MappingNodeFromString("2012-10-17"),
						Examples: []*core.MappingNode{
							core.MappingNodeFromString("2012-10-17"),
						},
					},
					"Statement": {
						Type:        provider.ResourceDefinitionsSchemaTypeArray,
						Description: "An array of individual statements in the policy document.",
						Items: &provider.ResourceDefinitionsSchema{
							Type:     provider.ResourceDefinitionsSchemaTypeObject,
							Label:    "Statement",
							Required: []string{"Effect", "Action"},
							Attributes: map[string]*provider.ResourceDefinitionsSchema{
								"Effect": {
									Type:        provider.ResourceDefinitionsSchemaTypeString,
									Description: "The effect of the statement. Valid values are 'Allow' and 'Deny'.",
									Pattern:     `^(Allow|Deny)$`,
									Examples: []*core.MappingNode{
										core.MappingNodeFromString("Allow"),
										core.MappingNodeFromString("Deny"),
									},
								},
								"Action": {
									Type:        provider.ResourceDefinitionsSchemaTypeArray,
									Description: "The action or actions that will be allowed or denied.",
									Items: &provider.ResourceDefinitionsSchema{
										Type: provider.ResourceDefinitionsSchemaTypeString,
										Examples: []*core.MappingNode{
											core.MappingNodeFromString("sqs:SendMessage"),
											core.MappingNodeFromString("sqs:ReceiveMessage"),
											core.MappingNodeFromString("sqs:DeleteMessage"),
											core.MappingNodeFromString("sqs:GetQueueAttributes"),
										},
									},
								},
								"Resource": {
									Type:        provider.ResourceDefinitionsSchemaTypeArray,
									Description: "The resource or resources that the statement covers.",
									Items: &provider.ResourceDefinitionsSchema{
										Type: provider.ResourceDefinitionsSchemaTypeString,
										Examples: []*core.MappingNode{
											core.MappingNodeFromString("*"),
											core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:my-queue"),
										},
									},
									Nullable: true,
								},
								"Principal": {
									Type:        provider.ResourceDefinitionsSchemaTypeObject,
									Description: "The principal that the statement applies to.",
									Examples: []*core.MappingNode{
										{
											Fields: map[string]*core.MappingNode{
												"AWS": core.MappingNodeFromString("arn:aws:iam::123456789012:user/username"),
											},
										},
										{
											Fields: map[string]*core.MappingNode{
												"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
											},
										},
									},
									Nullable: true,
								},
								"Condition": {
									Type:        provider.ResourceDefinitionsSchemaTypeObject,
									Description: "The conditions under which the statement applies.",
									Nullable:    true,
								},
							},
						},
					},
				},
				Nullable: true,
			},
			"tags": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of tags to associate with the SQS queue.",
				FormattedDescription: "A list of tags to associate with the SQS queue. " +
					"Tags are key-value pairs that help you organize and identify your AWS resources.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:     provider.ResourceDefinitionsSchemaTypeObject,
					Label:    "Tag",
					Required: []string{"key", "value"},
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"key": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The tag key.",
							MinLength:   1,
							MaxLength:   128,
							Examples: []*core.MappingNode{
								core.MappingNodeFromString("Environment"),
								core.MappingNodeFromString("Project"),
								core.MappingNodeFromString("Owner"),
							},
						},
						"value": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The tag value.",
							MinLength:   0,
							MaxLength:   256,
							Examples: []*core.MappingNode{
								core.MappingNodeFromString("Production"),
								core.MappingNodeFromString("MyProject"),
								core.MappingNodeFromString("admin@example.com"),
							},
						},
					},
				},
				Nullable: true,
			},

			// Computed fields
			"arn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the SQS queue.",
				FormattedDescription: "The Amazon Resource Name (ARN) of the SQS queue. " +
					"This is automatically generated by AWS and uniquely identifies the queue.",
				Computed: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:my-queue"),
					core.MappingNodeFromString("arn:aws:sqs:us-east-1:123456789012:processing-queue.fifo"),
				},
			},
			"queueUrl": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The URL of the SQS queue.",
				FormattedDescription: "The URL of the SQS queue. " +
					"This is the endpoint used to send and receive messages from the queue.",
				Computed: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("https://sqs.us-west-2.amazonaws.com/123456789012/my-queue"),
					core.MappingNodeFromString("https://sqs.us-east-1.amazonaws.com/123456789012/processing-queue.fifo"),
				},
			},
		},
	}
}
