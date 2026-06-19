Create a basic Lambda event source mapping that triggers a function from an SQS queue.

```blueprintlang
version "2025-11-02"

resource orderProcessorFunction: aws/lambda/function {
    metadata {
        displayName = "Order Processor Function"
    }
    spec {
        functionName = "order-processor"
        runtime = "nodejs18.x"
        handler = "index.handler"
        role = "arn:aws:iam::123456789012:role/lambda-execution-role"
        code = {
            zipFile = """
exports.handler = async (event) => {
  console.log('Processing orders:', JSON.stringify(event, null, 2));
  return { statusCode: 200 };
};
"""
        }
    }
}

resource orderQueueMapping: aws/lambda/eventSourceMapping {
    metadata {
        displayName = "Order Queue Mapping"
    }
    spec {
        functionName = resources.orderProcessorFunction.spec.functionName
        eventSourceArn = "arn:aws:sqs:us-east-1:123456789012:order-queue"
        batchSize = 10
        enabled = true
    }
}
```

```yaml
version: 2025-11-02

resources:
  orderProcessorFunction:
    type: aws/lambda/function
    metadata:
      displayName: Order Processor Function
    spec:
      functionName: order-processor
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            console.log('Processing orders:', JSON.stringify(event, null, 2));
            return { statusCode: 200 };
          };
  orderQueueMapping:
    type: aws/lambda/eventSourceMapping
    metadata:
      displayName: Order Queue Mapping
    spec:
      functionName: ${resources.orderProcessorFunction.spec.functionName}
      eventSourceArn: arn:aws:sqs:us-east-1:123456789012:order-queue
      batchSize: 10
      enabled: true
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "orderProcessorFunction": {
      "type": "aws/lambda/function",
      "metadata": {
        "displayName": "Order Processor Function"
      },
      "spec": {
        "functionName": "order-processor",
        "runtime": "nodejs18.x",
        "handler": "index.handler",
        "role": "arn:aws:iam::123456789012:role/lambda-execution-role",
        "code": {
          "zipFile": "exports.handler = async (event) => {\n  console.log('Processing orders:', JSON.stringify(event, null, 2));\n  return { statusCode: 200 };\n};"
        }
      }
    },
    "orderQueueMapping": {
      "type": "aws/lambda/eventSourceMapping",
      "metadata": {
        "displayName": "Order Queue Mapping"
      },
      "spec": {
        "functionName": "${resources.orderProcessorFunction.spec.functionName}",
        "eventSourceArn": "arn:aws:sqs:us-east-1:123456789012:order-queue",
        "batchSize": 10,
        "enabled": true
      }
    }
  }
}
```
