## EventBridge Rule to SQS Queue

Lets an EventBridge rule deliver matched events to an SQS queue.

When a rule targets a queue, the queue's access policy is configured to allow the rule to send messages to it, so events matched by the rule are delivered to the queue.

Add the queue as a target on the rule by referencing the queue's `arn` in the rule's `targets`; the connection takes effect automatically, with no link selector required. Per-target options including the target id, the FIFO message group id, input transformation and retry configuration are set on the rule's target entry.

### Encryption

If the queue is encrypted with a customer managed KMS key, that key's policy must allow EventBridge (`events.amazonaws.com`) to use it (`kms:GenerateDataKey` and `kms:Decrypt`).

### Example

```blueprintlang
version "2025-11-02"

resource orderCreatedRule: aws/events/rule {
    spec {
        name = "order-created-rule"
        eventPattern = {
            source = ["app.orders"]
        }
        targets = [
            {
                id = "order-queue",
                arn = orderQueue.spec.arn
            }
        ]
    }
}

resource orderQueue: aws/sqs/queue {
    spec {
        queueName = "order-queue"
    }
}
```
