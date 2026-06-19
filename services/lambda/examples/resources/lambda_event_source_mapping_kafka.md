Create a Lambda event source mapping that consumes records from an Amazon MSK (Managed Streaming for Kafka) cluster.

```blueprintlang
version "2025-11-02"

resource kafkaProcessorFunction: aws/lambda/function {
    metadata {
        displayName = "Kafka Processor Function"
    }
    spec {
        functionName = "kafka-processor"
        runtime = "nodejs18.x"
        handler = "index.handler"
        role = "arn:aws:iam::123456789012:role/lambda-execution-role"
        code = {
            zipFile = """
exports.handler = async (event) => {
  for (const record of event.records) {
    console.log('Processing Kafka record:', record);
  }
  return { statusCode: 200 };
};
"""
        }
    }
}

resource kafkaMapping: aws/lambda/eventSourceMapping {
    metadata {
        displayName = "Kafka Mapping"
    }
    spec {
        functionName = resources.kafkaProcessorFunction.spec.functionName
        eventSourceArn = "arn:aws:kafka:us-east-1:123456789012:cluster/my-cluster"
        topics = [ "user-events", "order-events" ]
        batchSize = 100
        startingPosition = "TRIM_HORIZON"
        maximumBatchingWindowInSeconds = 5
        enabled = true
        amazonManagedKafkaEventSourceConfig = {
            consumerGroupId = "my-consumer-group"
        }
        sourceAccessConfigurations = [
            {
                type = "VPC_SUBNET"
                uri = "subnet-12345678"
            },
            {
                type = "VPC_SECURITY_GROUP"
                uri = "sg-12345678"
            }
        ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  kafkaProcessorFunction:
    type: aws/lambda/function
    metadata:
      displayName: Kafka Processor Function
    spec:
      functionName: kafka-processor
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            for (const record of event.records) {
              console.log('Processing Kafka record:', record);
            }
            return { statusCode: 200 };
          };
  kafkaMapping:
    type: aws/lambda/eventSourceMapping
    metadata:
      displayName: Kafka Mapping
    spec:
      functionName: ${resources.kafkaProcessorFunction.spec.functionName}
      eventSourceArn: arn:aws:kafka:us-east-1:123456789012:cluster/my-cluster
      topics:
        - user-events
        - order-events
      batchSize: 100
      startingPosition: TRIM_HORIZON
      maximumBatchingWindowInSeconds: 5
      enabled: true
      amazonManagedKafkaEventSourceConfig:
        consumerGroupId: my-consumer-group
      sourceAccessConfigurations:
        - type: VPC_SUBNET
          uri: subnet-12345678
        - type: VPC_SECURITY_GROUP
          uri: sg-12345678
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "kafkaProcessorFunction": {
      "type": "aws/lambda/function",
      "metadata": {
        "displayName": "Kafka Processor Function"
      },
      "spec": {
        "functionName": "kafka-processor",
        "runtime": "nodejs18.x",
        "handler": "index.handler",
        "role": "arn:aws:iam::123456789012:role/lambda-execution-role",
        "code": {
          "zipFile": "exports.handler = async (event) => {\n  for (const record of event.records) {\n    console.log('Processing Kafka record:', record);\n  }\n  return { statusCode: 200 };\n};"
        }
      }
    },
    "kafkaMapping": {
      "type": "aws/lambda/eventSourceMapping",
      "metadata": {
        "displayName": "Kafka Mapping"
      },
      "spec": {
        "functionName": "${resources.kafkaProcessorFunction.spec.functionName}",
        "eventSourceArn": "arn:aws:kafka:us-east-1:123456789012:cluster/my-cluster",
        "topics": ["user-events", "order-events"],
        "batchSize": 100,
        "startingPosition": "TRIM_HORIZON",
        "maximumBatchingWindowInSeconds": 5,
        "enabled": true,
        "amazonManagedKafkaEventSourceConfig": {
          "consumerGroupId": "my-consumer-group"
        },
        "sourceAccessConfigurations": [
          {
            "type": "VPC_SUBNET",
            "uri": "subnet-12345678"
          },
          {
            "type": "VPC_SECURITY_GROUP",
            "uri": "sg-12345678"
          }
        ]
      }
    }
  }
}
```
