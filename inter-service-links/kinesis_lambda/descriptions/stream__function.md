## Kinesis Data Stream to Lambda Function Link (Stream Trigger)

This link configures a Kinesis data stream to trigger a Lambda function. When records are written to the stream, Lambda reads them in batches and invokes the function, enabling real-time processing of streaming data.

The link automatically:
1. **Grants stream read permissions** (`kinesis:DescribeStream`, `kinesis:DescribeStreamSummary`, `kinesis:GetRecords`, `kinesis:GetShardIterator`, `kinesis:ListShards`, `kinesis:ListStreams`) to the Lambda function's execution role.
2. **Creates an event source mapping** that reads records from the stream and invokes the function, configured by the function's annotations.

### Requirements

The Lambda function's execution role must be defined in the same blueprint and referenced from the function's `role` field.

If the Kinesis stream is encrypted with a customer managed KMS key, the key policy must allow the function's execution role to call `kms:Decrypt`. This is outside the link's control and must be granted on the KMS key.

### Example

```blueprintlang
version "2025-11-02"

resource eventsStream: aws/kinesis/stream {
    metadata {
        labels = {
            stream = "events"
        }
    }

    select by label {
        processor = "events"
    }

    spec {
        name = "events-stream"
        shardCount = 1
    }
}

resource eventProcessor: aws/lambda/function {
    metadata {
        labels = {
            processor = "events"
        }
        # Event source mapping config is on the function because it is specific
        # to how THIS function processes records from the stream.
        annotations = {
            "aws.kinesis.lambda.startingPosition" = "LATEST",
            "aws.kinesis.lambda.batchSize" = 50,
            "aws.kinesis.lambda.maximumBatchingWindowInSeconds" = 5,
            "aws.kinesis.lambda.parallelizationFactor" = 2,
            "aws.kinesis.lambda.reportBatchItemFailures" = true,
            "aws.kinesis.lambda.enabled" = true
        }
    }

    spec {
        functionName = "event-processor"
        role = eventProcessorRole.spec.arn
        # ... other function configuration
    }
}

resource eventProcessorRole: aws/iam/role {
    spec {
        name = "event-processor-role"
        # Stream read permissions are automatically added by the link
    }
}
```
