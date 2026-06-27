## EventBridge Rule to SQS Queue Link

This link grants an EventBridge rule permission to send messages to an SQS queue. SQS targets are delivered to by EventBridge using the queue's resource-based policy, so this link deploys a single link-owned managed `aws/sqs/queueInlinePolicy` intermediary granting the principal `events.amazonaws.com` permission to perform `sqs:SendMessage`, scoped by an `aws:SourceArn` condition to the rule's ARN. The intermediary's lifecycle (create, update, drift, destroy) is owned by the engine through Cloud Control; on removal the inline policy is destroyed.

The rule-to-queue wiring itself is modelled inline in the rule's `targets[]` array (the target entry, the FIFO `messageGroupId`, input transformation, retry configuration), with each target referencing the queue via its `arn`. That reference is what activates this link, `targets[].arn` is a link wiring slot, so no `linkSelector` is required. This link carries no user input.

> **Note (KMS-encrypted queues):** When the queue is encrypted with a customer managed KMS key, EventBridge additionally needs `kms:GenerateDataKey` and `kms:Decrypt` on that key. This link manages only the queue inline policy; granting those key-policy permissions to `events.amazonaws.com` will be handled automatically by a `rule → KMS key` link.

### Example

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "orderCreatedRule": {
      "type": "aws/events/rule",
      "spec": {
        "name": "order-created-rule",
        "eventPattern": { "source": ["app.orders"] },
        "targets": [
          {
            "id": "order-queue",
            "arn": "${orderQueue.spec.arn}"
          }
        ]
      }
    },
    "orderQueue": {
      "type": "aws/sqs/queue",
      "spec": { "queueName": "order-queue" }
    }
  }
}
```
