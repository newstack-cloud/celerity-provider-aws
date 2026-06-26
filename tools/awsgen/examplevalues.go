package main

import "github.com/newstack-cloud/bluelink/libs/blueprint/core"

// Per-type, per-field example value overrides (overlay
// Layer 1). They replace the generator's generic placeholders where a field needs
// a semantically valid value (an ARN, a region, an enum choice) or is a free-form
// object (an IAM policy document, an EventBridge event pattern) for which a
// meaningful sample is better than a generic placeholder. Keys are the Bluelink
// resource type; inner keys are camelCase spec field names.
var exampleSeedValues = map[string]map[string]*core.MappingNode{
	"aws/sqs/queue": {
		"queueName":         core.MappingNodeFromString("orders-queue"),
		"visibilityTimeout": core.MappingNodeFromInt(30),
		"kmsMasterKeyId":    core.MappingNodeFromString("alias/aws/sqs"),
		"redrivePolicy":     exampleRedrivePolicy(),
		"redriveAllowPolicy": core.MappingNodeFields(
			"redrivePermission", core.MappingNodeFromString("byQueue"),
			"sourceQueueArns", core.MappingNodeItems(
				core.MappingNodeFromString("arn:aws:sqs:us-east-1:123456789012:orders-source-queue"),
			),
		),
	},
	"aws/sqs/queueInlinePolicy": {
		"policyDocument": examplePolicyDocument(),
	},
	"aws/dynamodb/table": {
		"tableName":      core.MappingNodeFromString("orders"),
		"billingMode":    core.MappingNodeFromString("PAY_PER_REQUEST"),
		"policyDocument": examplePolicyDocument(),
	},
	"aws/dynamodb/globalTable": {
		"policyDocument": examplePolicyDocument(),
	},
	"aws/events/rule": {
		"eventPattern": exampleEventPattern(),
	},
	"aws/events/archive": {
		"eventPattern": exampleEventPattern(),
	},
	"aws/events/eventBus": {
		"policy": examplePolicyDocument(),
	},
	"aws/events/eventBusPolicy": {
		"statement": examplePolicyStatement(),
	},
	"aws/iam/managedPolicy": {
		"policyDocument": examplePolicyDocument(),
	},
	"aws/iam/rolePolicy": {
		"policyDocument": examplePolicyDocument(),
	},
	"aws/iam/userPolicy": {
		"policyDocument": examplePolicyDocument(),
	},
	"aws/iam/groupPolicy": {
		"policyDocument": examplePolicyDocument(),
	},
	"aws/iam/group": {
		"policyDocument": examplePolicyDocument(),
	},
	"aws/iam/user": {
		"policyDocument": examplePolicyDocument(),
	},
}

// A sample IAM-style policy document. Policy documents are
// re-typed as structured objects by the policy-document overlay, so they are
// authored with camelCase fields (translated to the PascalCase shape Cloud Control
// expects).
func examplePolicyDocument() *core.MappingNode {
	return core.MappingNodeFields(
		"version", core.MappingNodeFromString("2012-10-17"),
		"statement", core.MappingNodeItems(examplePolicyStatement()),
	)
}

func examplePolicyStatement() *core.MappingNode {
	return core.MappingNodeFields(
		"effect", core.MappingNodeFromString("Allow"),
		"action", core.MappingNodeItems(core.MappingNodeFromString("s3:GetObject")),
		"resource", core.MappingNodeFromString("arn:aws:s3:::example-bucket/*"),
	)
}

// A sample EventBridge event pattern (a free-form object,
// authored with verbatim lowercase keys as EventBridge expects).
func exampleEventPattern() *core.MappingNode {
	return core.MappingNodeFields(
		"source", core.MappingNodeItems(core.MappingNodeFromString("com.example.orders")),
	)
}

// A sample SQS redrive policy (a free-form object).
func exampleRedrivePolicy() *core.MappingNode {
	return core.MappingNodeFields(
		"deadLetterTargetArn", core.MappingNodeFromString("arn:aws:sqs:us-east-1:123456789012:orders-dlq"),
		"maxReceiveCount", core.MappingNodeFromInt(10),
	)
}

func seededExampleValue(blueprintType, field string) *core.MappingNode {
	fields, ok := exampleSeedValues[blueprintType]
	if !ok {
		return nil
	}
	return fields[field]
}
