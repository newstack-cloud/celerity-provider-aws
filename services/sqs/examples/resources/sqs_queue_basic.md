Create a basic SQS queue with minimal configuration.

```blueprintlang
version "2025-11-02"

resource myQueue: aws/sqs/queue {
    metadata {
        displayName = "My Basic SQS Queue"
    }
    spec {
        queueName = "my-basic-queue"
    }
}
```

```yaml
version: 2025-11-02

resources:
  myQueue:
    type: aws/sqs/queue
    metadata:
      displayName: My Basic SQS Queue
    spec:
      queueName: my-basic-queue
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "myQueue": {
      "type": "aws/sqs/queue",
      "metadata": {
        "displayName": "My Basic SQS Queue"
      },
      "spec": {
        "queueName": "my-basic-queue"
      }
    }
  }
}
```
