This example demonstrates creating a DynamoDB table with all available configuration options.

```blueprintlang
version "2025-11-02"

resource ordersTable: aws/dynamodb/table {
    metadata {
        displayName = "Orders Table with Full Configuration"
    }
    spec {
        tableName = "orders"
        attributeDefinitions = [
            {
                attributeName = "orderId"
                attributeType = "S"
            },
            {
                attributeName = "customerId"
                attributeType = "S"
            },
            {
                attributeName = "createdAt"
                attributeType = "N"
            },
            {
                attributeName = "status"
                attributeType = "S"
            }
        ]
        keySchema = [
            {
                attributeName = "orderId"
                keyType = "HASH"
            },
            {
                attributeName = "createdAt"
                keyType = "RANGE"
            }
        ]
        billingMode = "PROVISIONED"
        provisionedThroughput = {
            readCapacityUnits = 10
            writeCapacityUnits = 5
        }
        globalSecondaryIndexes = [
            {
                indexName = "customerId-index"
                keySchema = [
                    {
                        attributeName = "customerId"
                        keyType = "HASH"
                    },
                    {
                        attributeName = "createdAt"
                        keyType = "RANGE"
                    }
                ]
                projection = {
                    projectionType = "ALL"
                }
                provisionedThroughput = {
                    readCapacityUnits = 5
                    writeCapacityUnits = 2
                }
            },
            {
                indexName = "status-index"
                keySchema = [
                    {
                        attributeName = "status"
                        keyType = "HASH"
                    }
                ]
                projection = {
                    projectionType = "KEYS_ONLY"
                }
                provisionedThroughput = {
                    readCapacityUnits = 5
                    writeCapacityUnits = 2
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
            kmsMasterKeyId = "alias/my-dynamodb-key"
        }
        timeToLiveSpecification = {
            enabled = true
            attributeName = "expiresAt"
        }
        pointInTimeRecoverySpecification = {
            pointInTimeRecoveryEnabled = true
        }
        deletionProtectionEnabled = true
        tableClass = "STANDARD"
        tags = [
            {
                key = "Environment"
                value = "Production"
            },
            {
                key = "Service"
                value = "OrderManagement"
            },
            {
                key = "Team"
                value = "Commerce"
            }
        ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  ordersTable:
    type: aws/dynamodb/table
    metadata:
      displayName: Orders Table with Full Configuration
    spec:
      tableName: orders
      attributeDefinitions:
        - attributeName: orderId
          attributeType: S
        - attributeName: customerId
          attributeType: S
        - attributeName: createdAt
          attributeType: N
        - attributeName: status
          attributeType: S
      keySchema:
        - attributeName: orderId
          keyType: HASH
        - attributeName: createdAt
          keyType: RANGE
      billingMode: PROVISIONED
      provisionedThroughput:
        readCapacityUnits: 10
        writeCapacityUnits: 5
      globalSecondaryIndexes:
        - indexName: customerId-index
          keySchema:
            - attributeName: customerId
              keyType: HASH
            - attributeName: createdAt
              keyType: RANGE
          projection:
            projectionType: ALL
          provisionedThroughput:
            readCapacityUnits: 5
            writeCapacityUnits: 2
        - indexName: status-index
          keySchema:
            - attributeName: status
              keyType: HASH
          projection:
            projectionType: KEYS_ONLY
          provisionedThroughput:
            readCapacityUnits: 5
            writeCapacityUnits: 2
      streamSpecification:
        streamEnabled: true
        streamViewType: NEW_AND_OLD_IMAGES
      sseSpecification:
        enabled: true
        sseType: KMS
        kmsMasterKeyId: alias/my-dynamodb-key
      timeToLiveSpecification:
        enabled: true
        attributeName: expiresAt
      pointInTimeRecoverySpecification:
        pointInTimeRecoveryEnabled: true
      deletionProtectionEnabled: true
      tableClass: STANDARD
      tags:
        - key: Environment
          value: Production
        - key: Service
          value: OrderManagement
        - key: Team
          value: Commerce
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "ordersTable": {
      "type": "aws/dynamodb/table",
      "metadata": {
        "displayName": "Orders Table with Full Configuration"
      },
      "spec": {
        // Table name - must be unique within the region
        "tableName": "orders",
        // Define attributes used in key schema and indexes
        "attributeDefinitions": [
          {
            "attributeName": "orderId",
            "attributeType": "S"
          },
          {
            "attributeName": "customerId",
            "attributeType": "S"
          },
          {
            "attributeName": "createdAt",
            "attributeType": "N"
          },
          {
            "attributeName": "status",
            "attributeType": "S"
          }
        ],
        // Composite primary key: orderId (partition) + createdAt (sort)
        "keySchema": [
          {
            "attributeName": "orderId",
            "keyType": "HASH"
          },
          {
            "attributeName": "createdAt",
            "keyType": "RANGE"
          }
        ],
        "billingMode": "PROVISIONED",
        "provisionedThroughput": {
          "readCapacityUnits": 10,
          "writeCapacityUnits": 5
        },
        "globalSecondaryIndexes": [
          {
            "indexName": "customerId-index",
            "keySchema": [
              {
                "attributeName": "customerId",
                "keyType": "HASH"
              },
              {
                "attributeName": "createdAt",
                "keyType": "RANGE"
              }
            ],
            "projection": {
              "projectionType": "ALL"
            },
            "provisionedThroughput": {
              "readCapacityUnits": 5,
              "writeCapacityUnits": 2
            }
          },
          {
            "indexName": "status-index",
            "keySchema": [
              {
                "attributeName": "status",
                "keyType": "HASH"
              }
            ],
            "projection": {
              "projectionType": "KEYS_ONLY"
            },
            "provisionedThroughput": {
              "readCapacityUnits": 5,
              "writeCapacityUnits": 2
            }
          }
        ],
        // Enable streams for real-time processing
        "streamSpecification": {
          "streamEnabled": true,
          "streamViewType": "NEW_AND_OLD_IMAGES"
        },
        // Encrypt with a customer-managed KMS key
        "sseSpecification": {
          "enabled": true,
          "sseType": "KMS",
          "kmsMasterKeyId": "alias/my-dynamodb-key"
        },
        // Auto-expire items after TTL
        "timeToLiveSpecification": {
          "enabled": true,
          "attributeName": "expiresAt"
        },
        // Enable point-in-time recovery for data protection
        "pointInTimeRecoverySpecification": {
          "pointInTimeRecoveryEnabled": true
        },
        "deletionProtectionEnabled": true,
        "tableClass": "STANDARD",
        // Tags for resource management
        "tags": [
          {
            "key": "Environment",
            "value": "Production"
          },
          {
            "key": "Service",
            "value": "OrderManagement"
          },
          {
            "key": "Team",
            "value": "Commerce"
          }
        ]
      }
    }
  }
}
```
