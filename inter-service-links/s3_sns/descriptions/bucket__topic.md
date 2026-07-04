## S3 Bucket to SNS Topic

Publishes a message to an SNS topic when objects change in an S3 bucket.

When a bucket links to a topic, object events (by default `s3:ObjectCreated:*`) publish to the topic. The link grants S3 permission to publish to the topic and adds the notification to the bucket for you — you select the topic by label and the wiring is managed automatically.

This is the preferred, low-boilerplate way to wire bucket notifications. The bucket's notification configuration remains available for manual configuration; this link merges its own entry and preserves any entries you declare inline or that come from other links.

Topics encrypted with a customer managed KMS key require the key policy to allow `s3.amazonaws.com` (`kms:GenerateDataKey`, `kms:Decrypt`); that is outside this link's control.

### Example

```blueprintlang
version "2025-11-02"

resource ordersBucket: aws/s3/bucket {
    select by label {
        notify = "orders"
    }

    spec {
        bucketName = "orders"
    }
}

resource orderEvents: aws/sns/topic {
    metadata {
        labels = {
            notify = "orders"
        }
        annotations = {
            "aws.s3.sns.event.0" = "s3:ObjectCreated:*",
            "aws.s3.sns.filterPrefix" = "incoming/"
        }
    }

    spec {
        name = "order-events"
    }
}
```
