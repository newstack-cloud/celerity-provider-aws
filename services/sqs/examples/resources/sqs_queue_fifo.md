**SQS FIFO Queue**

This example demonstrates creating an SQS FIFO queue with FIFO-specific configuration options.

```yaml
resources:
  fifoQueue:
    type: aws/sqs/queue
    metadata:
      displayName: SQS FIFO Queue Example
    spec:
      queueName: my-fifo-queue.fifo
      fifoQueue: true
      contentBasedDeduplication: true
      fifoThroughputLimit: perMessageGroupId
      deduplicationScope: messageGroup
      delaySeconds: 0
      messageRetentionPeriod: 1209600
      receiveMessageWaitTimeSeconds: 20
      visibilityTimeout: 30
      maximumMessageSize: 262144
      sqsManagedSseEnabled: true
      kmsDataKeyReusePeriodSeconds: 300
      redrivePolicy:
        deadLetterTargetArn: arn:aws:sqs:us-west-2:123456789012:my-fifo-dlq.fifo
        maxReceiveCount: 3
      tags:
        - key: QueueType
          value: FIFO
        - key: Environment
          value: Production
```
