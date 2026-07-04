## S3 Bucket to SQS Queue

Delivers a message to an SQS queue when objects change in an S3 bucket.

When a bucket links to a queue, object events (by default `s3:ObjectCreated:*`) deliver a message to the queue. The link grants S3 permission to send messages to the queue and adds the notification to the bucket for you — you select the queue by label and the wiring is managed automatically.

This is the preferred, low-boilerplate way to wire bucket notifications. The bucket's notification configuration remains available for manual configuration; this link merges its own entry and preserves any entries you declare inline or that come from other links.

Queues encrypted with a customer managed KMS key require the key policy to allow `s3.amazonaws.com` (`kms:GenerateDataKey`, `kms:Decrypt`); that is outside this link's control.

### Example

```blueprintlang
version "2025-11-02"

resource ordersBucket: aws/s3/bucket {
    select by label {
        consumer = "orders"
    }

    spec {
        bucketName = "orders"
    }
}

resource orderEvents: aws/sqs/queue {
    metadata {
        labels = {
            consumer = "orders"
        }
        annotations = {
            "aws.s3.sqs.event.0" = "s3:ObjectCreated:*",
            "aws.s3.sqs.filterPrefix" = "incoming/"
        }
    }

    spec {
        queueName = "order-events"
        # ... other queue configuration
    }
}
```
