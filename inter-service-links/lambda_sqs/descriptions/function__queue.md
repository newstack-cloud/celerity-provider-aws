## Lambda Function to SQS Queue

Grants a Lambda function permission to send messages to (and optionally receive from) an SQS queue.

This is the **producer** side of a queue, a function that puts work onto it. When a function links to a queue, the link:

1. **Grants the execution role queue access** via an inline policy, scoped to the queue ARN. The access level comes from `aws.lambda.sqs.<targetQueue>.accessLevel` (default `send`):
   - `send` → `sqs:SendMessage`
   - `receive` → `sqs:ReceiveMessage`, `sqs:DeleteMessage`
   - `sendReceive` → both
   `sqs:GetQueueUrl` and `sqs:GetQueueAttributes` are always included so the SDK can resolve and inspect the queue.

2. **Populates an environment variable** (unless disabled) with the queue URL: `SQS_QUEUE_<targetQueue>` by default, or the name set with `aws.lambda.sqs.<targetQueue>.envVarName`.

3. **Activates an SQS interface VPC endpoint** when the function is VPC-isolated, so it can reach SQS without a NAT gateway. This is a no-op for functions that are not attached to a VPC.

For **event-driven consumption** of a queue (a function triggered by messages arriving), use the `aws/sqs/queue` → `aws/lambda/function` link instead, that will provision an event source mapping. This link is only for a function that actively sends to a queue.

### Example

```blueprintlang
version "2025-11-02"

resource submitOrder: aws/lambda/function {
    metadata {
        labels = {
            queue = "orders"
        }
        annotations = {
            "aws.lambda.sqs.ordersQueue.envVarName" = "ORDERS_QUEUE"
            "aws.lambda.sqs.ordersQueue.accessLevel" = "send"
        }
    }

    select by label {
        queue = "orders"
    }

    spec {
        functionName = "submit-order"
        # ... other function configuration
    }
}

resource ordersQueue: aws/sqs/queue {
    metadata {
        labels = {
            queue = "orders"
        }
    }

    spec {
        queueName = "orders"
    }
}
```
