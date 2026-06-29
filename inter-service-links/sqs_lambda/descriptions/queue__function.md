## SQS Queue to Lambda Function Link (Trigger)

This link configures an SQS queue to trigger a Lambda function. When messages are sent to the queue, Lambda polls the queue and invokes the function with batches of messages, enabling event-driven processing of work items.

The link automatically:
1. **Grants the function's execution role permission to consume from the queue** (`sqs:ReceiveMessage`, `sqs:DeleteMessage`, `sqs:GetQueueAttributes`)
2. **Connects the queue to the function** so new messages trigger invocations, using the batching, concurrency and filtering settings configured via the function's annotations

### Requirements

The Lambda function's execution role must be defined in the same blueprint.

If the queue is encrypted with a customer managed KMS key, the key's policy must allow the function's execution role to use the key (`kms:Decrypt`). This is configured on the KMS key and is outside the scope of this link.

### Annotation Placement

Trigger settings are configured on the **function** (`aws.sqs.lambda.batchSize`, `aws.sqs.lambda.maximumConcurrency`, etc.) because they configure how this specific function consumes from the queue. A single queue can trigger multiple functions, each with its own settings. Relationship annotations include both service names and the feature (`sqs.lambda`).

### Example

```blueprintlang
version "2025-11-02"

resource ordersQueue: aws/sqs/queue {
    metadata {
        labels = {
            consumer = "orders"
        }
    }

    select by label {
        processor = "orders"
    }

    spec {
        queueName = "orders-queue"
    }
}

resource orderProcessor: aws/lambda/function {
    metadata {
        labels = {
            processor = "orders"
        }
        # Trigger config is on the function because it's specific to how THIS
        # function consumes from the queue. Relationship annotations use
        # aws.sqs.lambda.* to indicate they configure the SQS→Lambda trigger.
        annotations = {
            "aws.sqs.lambda.batchSize" = 10,
            "aws.sqs.lambda.maximumBatchingWindowInSeconds" = 5,
            "aws.sqs.lambda.reportBatchItemFailures" = true,
            "aws.sqs.lambda.enabled" = true
        }
    }

    spec {
        functionName = "order-processor"
        role = orderProcessorRole.spec.arn
        # ... other function configuration
    }
}

resource orderProcessorRole: aws/iam/role {
    spec {
        name = "order-processor-role"
        # Queue consume permissions are automatically added by the link
    }
}
```

In this example:
- The SQS queue links to the Lambda function via label selector
- When messages are sent to the orders queue, Lambda is triggered with batches of up to 10 messages
- The function reports partial batch failures so only failed messages are retried
- Queue consume permissions are automatically added to the function's execution role
