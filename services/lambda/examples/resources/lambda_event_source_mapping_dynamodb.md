Create a Lambda event source mapping that processes change records from a DynamoDB stream.

```blueprintlang
version "2025-11-02"

resource dynamodbProcessorFunction: aws/lambda/function {
    metadata {
        displayName = "DynamoDB Processor Function"
    }
    spec {
        functionName = "dynamodb-processor"
        runtime = "nodejs18.x"
        handler = "index.handler"
        role = "arn:aws:iam::123456789012:role/lambda-execution-role"
        code = {
            zipFile = """
exports.handler = async (event) => {
  for (const record of event.Records) {
    console.log('Processing DynamoDB change:', record.dynamodb);
  }
  return { statusCode: 200 };
};
"""
        }
    }
}

resource dynamodbMapping: aws/lambda/eventSourceMapping {
    metadata {
        displayName = "DynamoDB Stream Mapping"
    }
    spec {
        functionName = resources.dynamodbProcessorFunction.spec.functionName
        eventSourceArn = "arn:aws:dynamodb:us-east-1:123456789012:table/users/stream/2024-01-01T00:00:00.000"
        batchSize = 50
        startingPosition = "LATEST"
        maximumBatchingWindowInSeconds = 10
        maximumRecordAgeInSeconds = 7200
        maximumRetryAttempts = 5
        bisectBatchOnFunctionError = true
        parallelizationFactor = 1
        enabled = true
        functionResponseTypes = [ "ReportBatchItemFailures" ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  dynamodbProcessorFunction:
    type: aws/lambda/function
    metadata:
      displayName: DynamoDB Processor Function
    spec:
      functionName: dynamodb-processor
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            for (const record of event.Records) {
              console.log('Processing DynamoDB change:', record.dynamodb);
            }
            return { statusCode: 200 };
          };
  dynamodbMapping:
    type: aws/lambda/eventSourceMapping
    metadata:
      displayName: DynamoDB Stream Mapping
    spec:
      functionName: ${resources.dynamodbProcessorFunction.spec.functionName}
      eventSourceArn: arn:aws:dynamodb:us-east-1:123456789012:table/users/stream/2024-01-01T00:00:00.000
      batchSize: 50
      startingPosition: LATEST
      maximumBatchingWindowInSeconds: 10
      maximumRecordAgeInSeconds: 7200
      maximumRetryAttempts: 5
      bisectBatchOnFunctionError: true
      parallelizationFactor: 1
      enabled: true
      functionResponseTypes:
        - ReportBatchItemFailures
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "dynamodbProcessorFunction": {
      "type": "aws/lambda/function",
      "metadata": {
        "displayName": "DynamoDB Processor Function"
      },
      "spec": {
        "functionName": "dynamodb-processor",
        "runtime": "nodejs18.x",
        "handler": "index.handler",
        "role": "arn:aws:iam::123456789012:role/lambda-execution-role",
        "code": {
          "zipFile": "exports.handler = async (event) => {\n  for (const record of event.Records) {\n    console.log('Processing DynamoDB change:', record.dynamodb);\n  }\n  return { statusCode: 200 };\n};"
        }
      }
    },
    "dynamodbMapping": {
      "type": "aws/lambda/eventSourceMapping",
      "metadata": {
        "displayName": "DynamoDB Stream Mapping"
      },
      "spec": {
        "functionName": "${resources.dynamodbProcessorFunction.spec.functionName}",
        "eventSourceArn": "arn:aws:dynamodb:us-east-1:123456789012:table/users/stream/2024-01-01T00:00:00.000",
        "batchSize": 50,
        "startingPosition": "LATEST",
        "maximumBatchingWindowInSeconds": 10,
        "maximumRecordAgeInSeconds": 7200,
        "maximumRetryAttempts": 5,
        "bisectBatchOnFunctionError": true,
        "parallelizationFactor": 1,
        "enabled": true,
        "functionResponseTypes": ["ReportBatchItemFailures"]
      }
    }
  }
}
```
