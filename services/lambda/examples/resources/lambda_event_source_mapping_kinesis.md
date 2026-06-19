Create a Lambda event source mapping for an Amazon Kinesis stream with advanced stream processing configurations.

```blueprintlang
version "2025-11-02"

resource kinesisProcessorFunction: aws/lambda/function {
    metadata {
        displayName = "Kinesis Processor Function"
    }
    spec {
        functionName = "kinesis-processor"
        runtime = "nodejs18.x"
        handler = "index.handler"
        role = "arn:aws:iam::123456789012:role/lambda-execution-role"
        code = {
            zipFile = """
exports.handler = async (event) => {
  for (const record of event.Records) {
    console.log('Processing Kinesis record:', record.kinesis.data);
  }
  return { statusCode: 200 };
};
"""
        }
    }
}

resource kinesisMapping: aws/lambda/eventSourceMapping {
    metadata {
        displayName = "Kinesis Mapping"
    }
    spec {
        functionName = resources.kinesisProcessorFunction.spec.functionName
        eventSourceArn = "arn:aws:kinesis:us-east-1:123456789012:stream/data-stream"
        batchSize = 100
        startingPosition = "TRIM_HORIZON"
        maximumBatchingWindowInSeconds = 5
        maximumRecordAgeInSeconds = 3600
        maximumRetryAttempts = 3
        bisectBatchOnFunctionError = true
        parallelizationFactor = 2
        tumblingWindowInSeconds = 300
        enabled = true
        functionResponseTypes = [ "ReportBatchItemFailures" ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  kinesisProcessorFunction:
    type: aws/lambda/function
    metadata:
      displayName: Kinesis Processor Function
    spec:
      functionName: kinesis-processor
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            for (const record of event.Records) {
              console.log('Processing Kinesis record:', record.kinesis.data);
            }
            return { statusCode: 200 };
          };
  kinesisMapping:
    type: aws/lambda/eventSourceMapping
    metadata:
      displayName: Kinesis Mapping
    spec:
      functionName: ${resources.kinesisProcessorFunction.spec.functionName}
      eventSourceArn: arn:aws:kinesis:us-east-1:123456789012:stream/data-stream
      batchSize: 100
      startingPosition: TRIM_HORIZON
      maximumBatchingWindowInSeconds: 5
      maximumRecordAgeInSeconds: 3600
      maximumRetryAttempts: 3
      bisectBatchOnFunctionError: true
      parallelizationFactor: 2
      tumblingWindowInSeconds: 300
      enabled: true
      functionResponseTypes:
        - ReportBatchItemFailures
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "kinesisProcessorFunction": {
      "type": "aws/lambda/function",
      "metadata": {
        "displayName": "Kinesis Processor Function"
      },
      "spec": {
        "functionName": "kinesis-processor",
        "runtime": "nodejs18.x",
        "handler": "index.handler",
        "role": "arn:aws:iam::123456789012:role/lambda-execution-role",
        "code": {
          "zipFile": "exports.handler = async (event) => {\n  for (const record of event.Records) {\n    console.log('Processing Kinesis record:', record.kinesis.data);\n  }\n  return { statusCode: 200 };\n};"
        }
      }
    },
    "kinesisMapping": {
      "type": "aws/lambda/eventSourceMapping",
      "metadata": {
        "displayName": "Kinesis Mapping"
      },
      "spec": {
        "functionName": "${resources.kinesisProcessorFunction.spec.functionName}",
        "eventSourceArn": "arn:aws:kinesis:us-east-1:123456789012:stream/data-stream",
        "batchSize": 100,
        "startingPosition": "TRIM_HORIZON",
        "maximumBatchingWindowInSeconds": 5,
        "maximumRecordAgeInSeconds": 3600,
        "maximumRetryAttempts": 3,
        "bisectBatchOnFunctionError": true,
        "parallelizationFactor": 2,
        "tumblingWindowInSeconds": 300,
        "enabled": true,
        "functionResponseTypes": ["ReportBatchItemFailures"]
      }
    }
  }
}
```
