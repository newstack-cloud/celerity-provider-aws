A AWS SQS Queue configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource queue: aws/sqs/queue {
    metadata {
        displayName = "AWS SQS Queue complete"
    }
    spec {
        contentBasedDeduplication = false
        deduplicationScope = "example-deduplication-scope"
        delaySeconds = 1
        fifoQueue = false
        fifoThroughputLimit = "example-fifo-throughput-limit"
        kmsDataKeyReusePeriodSeconds = 1
        kmsMasterKeyId = "alias/aws/sqs"
        maximumMessageSize = 1
        messageRetentionPeriod = 1
        queueName = "orders-queue"
        receiveMessageWaitTimeSeconds = 1
        redriveAllowPolicy = {
            redrivePermission = "byQueue",
            sourceQueueArns = [
                "arn:aws:sqs:us-east-1:123456789012:orders-source-queue"
            ]
        }
        redrivePolicy = {
            deadLetterTargetArn = "arn:aws:sqs:us-east-1:123456789012:orders-dlq",
            maxReceiveCount = 10
        }
        sqsManagedSseEnabled = false
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        visibilityTimeout = 30
    }
}
```

```yaml
version: "2025-11-02"
resources:
    queue:
        type: aws/sqs/queue
        metadata:
            displayName: AWS SQS Queue complete
        spec:
            contentBasedDeduplication: false
            deduplicationScope: example-deduplication-scope
            delaySeconds: 1
            fifoQueue: false
            fifoThroughputLimit: example-fifo-throughput-limit
            kmsDataKeyReusePeriodSeconds: 1
            kmsMasterKeyId: alias/aws/sqs
            maximumMessageSize: 1
            messageRetentionPeriod: 1
            queueName: orders-queue
            receiveMessageWaitTimeSeconds: 1
            redriveAllowPolicy:
                redrivePermission: byQueue
                sourceQueueArns:
                    - arn:aws:sqs:us-east-1:123456789012:orders-source-queue
            redrivePolicy:
                deadLetterTargetArn: arn:aws:sqs:us-east-1:123456789012:orders-dlq
                maxReceiveCount: 10
            sqsManagedSseEnabled: false
            tags:
                - key: example-key
                  value: example-value
            visibilityTimeout: 30
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "queue": {
      "type": "aws/sqs/queue",
      "metadata": {
        "displayName": "AWS SQS Queue complete"
      },
      "spec": {
        "contentBasedDeduplication": false,
        "deduplicationScope": "example-deduplication-scope",
        "delaySeconds": 1,
        "fifoQueue": false,
        "fifoThroughputLimit": "example-fifo-throughput-limit",
        "kmsDataKeyReusePeriodSeconds": 1,
        "kmsMasterKeyId": "alias/aws/sqs",
        "maximumMessageSize": 1,
        "messageRetentionPeriod": 1,
        "queueName": "orders-queue",
        "receiveMessageWaitTimeSeconds": 1,
        "redriveAllowPolicy": {
          "redrivePermission": "byQueue",
          "sourceQueueArns": [
            "arn:aws:sqs:us-east-1:123456789012:orders-source-queue"
          ]
        },
        "redrivePolicy": {
          "deadLetterTargetArn": "arn:aws:sqs:us-east-1:123456789012:orders-dlq",
          "maxReceiveCount": 10
        },
        "sqsManagedSseEnabled": false,
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "visibilityTimeout": 30
      }
    }
  }
}
```
