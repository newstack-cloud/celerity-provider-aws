**SQS Queue - Basic**

This example demonstrates creating a basic SQS queue with minimal configuration.

```yaml
resources:
  myQueue:
    type: aws/sqs/queue
    metadata:
      displayName: My Basic SQS Queue
    spec:
      queueName: my-basic-queue
```
