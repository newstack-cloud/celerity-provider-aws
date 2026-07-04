## S3 Bucket to Lambda Function

Invokes a Lambda function when objects change in an S3 bucket.

When a bucket links to a function, object events (by default `s3:ObjectCreated:*`) trigger the function. The link grants S3 permission to invoke the function and adds the notification to the bucket for you, you select the function by label and the wiring is managed automatically.

This is the preferred, low-boilerplate way to wire bucket notifications. The bucket's notification configuration remains available for manual configuration; this link merges its own entry and preserves any entries you declare inline or that come from other links.

### Example

```blueprintlang
version "2025-11-02"

resource ordersBucket: aws/s3/bucket {
    select by label {
        processor = "orders"
    }

    spec {
        bucketName = "orders"
    }
}

resource processOrder: aws/lambda/function {
    metadata {
        labels = {
            processor = "orders"
        }
        annotations = {
            "aws.s3.lambda.event.0" = "s3:ObjectCreated:*",
            "aws.s3.lambda.filterPrefix" = "incoming/"
        }
    }

    spec {
        functionName = "process-order"
        # ... other function configuration
    }
}
```

In this example:
- `ordersBucket` invokes `processOrder` when an object is created under the `incoming/` prefix.
- S3 is granted permission to invoke the function automatically.
