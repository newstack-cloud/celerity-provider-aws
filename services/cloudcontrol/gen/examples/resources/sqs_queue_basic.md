A basic AWS SQS Queue with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource queue: aws/sqs/queue {
    metadata {
        displayName = "AWS SQS Queue basic"
    }
    spec {
        queueName = "orders-queue"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    queue:
        type: aws/sqs/queue
        metadata:
            displayName: AWS SQS Queue basic
        spec:
            queueName: orders-queue
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "queue": {
      "type": "aws/sqs/queue",
      "metadata": {
        "displayName": "AWS SQS Queue basic"
      },
      "spec": {
        "queueName": "orders-queue"
      }
    }
  }
}
```
