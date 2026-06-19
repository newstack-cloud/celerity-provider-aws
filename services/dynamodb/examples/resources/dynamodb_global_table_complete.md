This example demonstrates a fully configured DynamoDB global table with provisioned throughput, a global secondary index, encryption, per-replica overrides, and tags.

```blueprintlang
version "2025-11-02"

resource ordersGlobalTable: aws/dynamodb/globalTable {
    metadata {
        displayName = "Orders Global Table"
    }
    spec {
        tableName = "orders"
        attributeDefinitions = [
            {
                attributeName = "id"
                attributeType = "S"
            },
            {
                attributeName = "customerId"
                attributeType = "S"
            }
        ]
        keySchema = [
            {
                attributeName = "id"
                keyType = "HASH"
            }
        ]
        billingMode = "PROVISIONED"
        provisionedThroughput = {
            readCapacityUnits = 5
            writeCapacityUnits = 5
        }
        globalSecondaryIndexes = [
            {
                indexName = "customerId-index"
                keySchema = [
                    {
                        attributeName = "customerId"
                        keyType = "HASH"
                    }
                ]
                projection = {
                    projectionType = "ALL"
                }
                provisionedThroughput = {
                    readCapacityUnits = 5
                    writeCapacityUnits = 5
                }
            }
        ]
        streamSpecification = {
            streamEnabled = true
            streamViewType = "NEW_AND_OLD_IMAGES"
        }
        sseSpecification = {
            enabled = true
            sseType = "KMS"
        }
        tableClass = "STANDARD"
        deletionProtectionEnabled = true
        replicas = [
            {
                regionName = "us-east-1"
                tableClass = "STANDARD"
                provisionedThroughputOverride = {
                    readCapacityUnits = 10
                }
                pointInTimeRecovery = {
                    pointInTimeRecoveryEnabled = true
                }
            },
            {
                regionName = "eu-west-1"
                globalSecondaryIndexes = [
                    {
                        indexName = "customerId-index"
                        provisionedThroughputOverride = {
                            readCapacityUnits = 5
                        }
                    }
                ]
            }
        ]
        tags = [
            {
                key = "Environment"
                value = "production"
            },
            {
                key = "Team"
                value = "orders"
            }
        ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  ordersGlobalTable:
    type: aws/dynamodb/globalTable
    metadata:
      displayName: Orders Global Table
    spec:
      tableName: orders
      attributeDefinitions:
        - attributeName: id
          attributeType: S
        - attributeName: customerId
          attributeType: S
      keySchema:
        - attributeName: id
          keyType: HASH
      billingMode: PROVISIONED
      provisionedThroughput:
        readCapacityUnits: 5
        writeCapacityUnits: 5
      globalSecondaryIndexes:
        - indexName: customerId-index
          keySchema:
            - attributeName: customerId
              keyType: HASH
          projection:
            projectionType: ALL
          provisionedThroughput:
            readCapacityUnits: 5
            writeCapacityUnits: 5
      streamSpecification:
        streamEnabled: true
        streamViewType: NEW_AND_OLD_IMAGES
      sseSpecification:
        enabled: true
        sseType: KMS
      tableClass: STANDARD
      deletionProtectionEnabled: true
      replicas:
        - regionName: us-east-1
          tableClass: STANDARD
          provisionedThroughputOverride:
            readCapacityUnits: 10
          pointInTimeRecovery:
            pointInTimeRecoveryEnabled: true
        - regionName: eu-west-1
          globalSecondaryIndexes:
            - indexName: customerId-index
              provisionedThroughputOverride:
                readCapacityUnits: 5
      tags:
        - key: Environment
          value: production
        - key: Team
          value: orders
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "ordersGlobalTable": {
      "type": "aws/dynamodb/globalTable",
      "metadata": {
        "displayName": "Orders Global Table"
      },
      "spec": {
        "tableName": "orders",
        "attributeDefinitions": [
          {
            "attributeName": "id",
            "attributeType": "S"
          },
          {
            "attributeName": "customerId",
            "attributeType": "S"
          }
        ],
        "keySchema": [
          {
            "attributeName": "id",
            "keyType": "HASH"
          }
        ],
        "billingMode": "PROVISIONED",
        "provisionedThroughput": {
          "readCapacityUnits": 5,
          "writeCapacityUnits": 5
        },
        "globalSecondaryIndexes": [
          {
            "indexName": "customerId-index",
            "keySchema": [
              {
                "attributeName": "customerId",
                "keyType": "HASH"
              }
            ],
            "projection": {
              "projectionType": "ALL"
            },
            "provisionedThroughput": {
              "readCapacityUnits": 5,
              "writeCapacityUnits": 5
            }
          }
        ],
        "streamSpecification": {
          "streamEnabled": true,
          "streamViewType": "NEW_AND_OLD_IMAGES"
        },
        "sseSpecification": {
          "enabled": true,
          "sseType": "KMS"
        },
        "tableClass": "STANDARD",
        "deletionProtectionEnabled": true,
        // Per-replica overrides (table class, throughput, PITR, GSI throughput)
        "replicas": [
          {
            "regionName": "us-east-1",
            "tableClass": "STANDARD",
            "provisionedThroughputOverride": {
              "readCapacityUnits": 10
            },
            "pointInTimeRecovery": {
              "pointInTimeRecoveryEnabled": true
            }
          },
          {
            "regionName": "eu-west-1",
            "globalSecondaryIndexes": [
              {
                "indexName": "customerId-index",
                "provisionedThroughputOverride": {
                  "readCapacityUnits": 5
                }
              }
            ]
          }
        ],
        "tags": [
          {
            "key": "Environment",
            "value": "production"
          },
          {
            "key": "Team",
            "value": "orders"
          }
        ]
      }
    }
  }
}
```
