Create a Lambda event source mapping that consumes messages from an Amazon MQ (ActiveMQ/RabbitMQ) broker.

```blueprintlang
version "2025-11-02"

resource mqProcessorFunction: aws/lambda/function {
    metadata {
        displayName = "MQ Processor Function"
    }
    spec {
        functionName = "mq-processor"
        runtime = "nodejs18.x"
        handler = "index.handler"
        role = "arn:aws:iam::123456789012:role/lambda-execution-role"
        code = {
            zipFile = """
exports.handler = async (event) => {
  for (const record of event.Records) {
    console.log('Processing MQ message:', record);
  }
  return { statusCode: 200 };
};
"""
        }
    }
}

resource mqMapping: aws/lambda/eventSourceMapping {
    metadata {
        displayName = "MQ Mapping"
    }
    spec {
        functionName = resources.mqProcessorFunction.spec.functionName
        eventSourceArn = "arn:aws:mq:us-east-1:123456789012:broker/my-broker"
        queues = [ "order-queue", "notification-queue" ]
        batchSize = 10
        enabled = true
        sourceAccessConfigurations = [
            {
                type = "VPC_SUBNET"
                uri = "subnet-12345678"
            },
            {
                type = "VPC_SECURITY_GROUP"
                uri = "sg-12345678"
            },
            {
                type = "BASIC_AUTH"
                uri = "arn:aws:secretsmanager:us-east-1:123456789012:secret:mq-credentials"
            }
        ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  mqProcessorFunction:
    type: aws/lambda/function
    metadata:
      displayName: MQ Processor Function
    spec:
      functionName: mq-processor
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            for (const record of event.Records) {
              console.log('Processing MQ message:', record);
            }
            return { statusCode: 200 };
          };
  mqMapping:
    type: aws/lambda/eventSourceMapping
    metadata:
      displayName: MQ Mapping
    spec:
      functionName: ${resources.mqProcessorFunction.spec.functionName}
      eventSourceArn: arn:aws:mq:us-east-1:123456789012:broker/my-broker
      queues:
        - order-queue
        - notification-queue
      batchSize: 10
      enabled: true
      sourceAccessConfigurations:
        - type: VPC_SUBNET
          uri: subnet-12345678
        - type: VPC_SECURITY_GROUP
          uri: sg-12345678
        - type: BASIC_AUTH
          uri: arn:aws:secretsmanager:us-east-1:123456789012:secret:mq-credentials
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "mqProcessorFunction": {
      "type": "aws/lambda/function",
      "metadata": {
        "displayName": "MQ Processor Function"
      },
      "spec": {
        "functionName": "mq-processor",
        "runtime": "nodejs18.x",
        "handler": "index.handler",
        "role": "arn:aws:iam::123456789012:role/lambda-execution-role",
        "code": {
          "zipFile": "exports.handler = async (event) => {\n  for (const record of event.Records) {\n    console.log('Processing MQ message:', record);\n  }\n  return { statusCode: 200 };\n};"
        }
      }
    },
    "mqMapping": {
      "type": "aws/lambda/eventSourceMapping",
      "metadata": {
        "displayName": "MQ Mapping"
      },
      "spec": {
        "functionName": "${resources.mqProcessorFunction.spec.functionName}",
        "eventSourceArn": "arn:aws:mq:us-east-1:123456789012:broker/my-broker",
        "queues": ["order-queue", "notification-queue"],
        "batchSize": 10,
        "enabled": true,
        "sourceAccessConfigurations": [
          {
            "type": "VPC_SUBNET",
            "uri": "subnet-12345678"
          },
          {
            "type": "VPC_SECURITY_GROUP",
            "uri": "sg-12345678"
          },
          {
            "type": "BASIC_AUTH",
            "uri": "arn:aws:secretsmanager:us-east-1:123456789012:secret:mq-credentials"
          }
        ]
      }
    }
  }
}
```
