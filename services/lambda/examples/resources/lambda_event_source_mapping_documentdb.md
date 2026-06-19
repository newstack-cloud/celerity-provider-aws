Create a Lambda event source mapping that processes change streams from an Amazon DocumentDB cluster.

```blueprintlang
version "2025-11-02"

resource documentdbProcessorFunction: aws/lambda/function {
    metadata {
        displayName = "DocumentDB Processor Function"
    }
    spec {
        functionName = "documentdb-processor"
        runtime = "nodejs18.x"
        handler = "index.handler"
        role = "arn:aws:iam::123456789012:role/lambda-execution-role"
        code = {
            zipFile = """
exports.handler = async (event) => {
  for (const record of event.Records) {
    console.log('Processing DocumentDB change:', record);
  }
  return { statusCode: 200 };
};
"""
        }
    }
}

resource documentdbMapping: aws/lambda/eventSourceMapping {
    metadata {
        displayName = "DocumentDB Mapping"
    }
    spec {
        functionName = resources.documentdbProcessorFunction.spec.functionName
        eventSourceArn = "arn:aws:docdb:us-east-1:123456789012:cluster/my-cluster"
        batchSize = 100
        enabled = true
        documentDbEventSourceConfig = {
            databaseName = "mydatabase"
            collectionName = "users"
            fullDocument = "UpdateLookup"
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
  documentdbProcessorFunction:
    type: aws/lambda/function
    metadata:
      displayName: DocumentDB Processor Function
    spec:
      functionName: documentdb-processor
      runtime: nodejs18.x
      handler: index.handler
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      code:
        zipFile: |
          exports.handler = async (event) => {
            for (const record of event.Records) {
              console.log('Processing DocumentDB change:', record);
            }
            return { statusCode: 200 };
          };
  documentdbMapping:
    type: aws/lambda/eventSourceMapping
    metadata:
      displayName: DocumentDB Mapping
    spec:
      functionName: ${resources.documentdbProcessorFunction.spec.functionName}
      eventSourceArn: arn:aws:docdb:us-east-1:123456789012:cluster/my-cluster
      batchSize: 100
      enabled: true
      documentDbEventSourceConfig:
        databaseName: mydatabase
        collectionName: users
        fullDocument: UpdateLookup
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
    "documentdbProcessorFunction": {
      "type": "aws/lambda/function",
      "metadata": {
        "displayName": "DocumentDB Processor Function"
      },
      "spec": {
        "functionName": "documentdb-processor",
        "runtime": "nodejs18.x",
        "handler": "index.handler",
        "role": "arn:aws:iam::123456789012:role/lambda-execution-role",
        "code": {
          "zipFile": "exports.handler = async (event) => {\n  for (const record of event.Records) {\n    console.log('Processing DocumentDB change:', record);\n  }\n  return { statusCode: 200 };\n};"
        }
      }
    },
    "documentdbMapping": {
      "type": "aws/lambda/eventSourceMapping",
      "metadata": {
        "displayName": "DocumentDB Mapping"
      },
      "spec": {
        "functionName": "${resources.documentdbProcessorFunction.spec.functionName}",
        "eventSourceArn": "arn:aws:docdb:us-east-1:123456789012:cluster/my-cluster",
        "batchSize": 100,
        "enabled": true,
        "documentDbEventSourceConfig": {
          "databaseName": "mydatabase",
          "collectionName": "users",
          "fullDocument": "UpdateLookup"
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
