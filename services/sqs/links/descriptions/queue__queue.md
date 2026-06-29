The link type used to configure a dead-letter queue relationship between two SQS queues.
This will automatically configure the source queue with a redrive policy that sends failed messages
to the dead-letter queue after a specified number of receive attempts.

The link supports customizing the maximum receive count through annotations. The default value is 10.

**Example**

```blueprintlang
version "2025-11-02"

resource processingQueue: aws/sqs/queue {
    metadata {
        displayName = "Processing Queue"
        labels = {
            app = "order-processor"
        }
        annotations = {
            "aws.sqs.redrive.maxReceiveCount" = 5
        }
    }

    select by label {
        app = "order-processor"
    }

    spec {
        queueName = "processing-queue"
        visibilityTimeout = 300
    }
}

resource deadLetterQueue: aws/sqs/queue {
    metadata {
        displayName = "Dead Letter Queue"
        labels = {
            app = "order-processor"
        }
    }

    spec {
        queueName = "dead-letter-queue"
        messageRetentionPeriod = 1209600  # 14 days
    }
}
```

In this example, messages that fail to be processed after 5 attempts in the `processingQueue`
will be automatically moved to the `deadLetterQueue`.
