This example demonstrates creating a basic DynamoDB global table replicated across two regions.

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
            }
        ]
        keySchema = [
            {
                attributeName = "id"
                keyType = "HASH"
            }
        ]
        billingMode = "PAY_PER_REQUEST"
        streamSpecification = {
            streamEnabled = true
            streamViewType = "NEW_AND_OLD_IMAGES"
        }
        replicas = [
            {
                regionName = "us-east-1"
            },
            {
                regionName = "eu-west-1"
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
      keySchema:
        - attributeName: id
          keyType: HASH
      billingMode: PAY_PER_REQUEST
      streamSpecification:
        streamEnabled: true
        streamViewType: NEW_AND_OLD_IMAGES
      replicas:
        - regionName: us-east-1
        - regionName: eu-west-1
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
        // Table name - shared across all replicas
        "tableName": "orders",
        "attributeDefinitions": [
          {
            "attributeName": "id",
            "attributeType": "S"
          }
        ],
        "keySchema": [
          {
            "attributeName": "id",
            "keyType": "HASH"
          }
        ],
        "billingMode": "PAY_PER_REQUEST",
        // Global tables require streams with the NEW_AND_OLD_IMAGES view type
        "streamSpecification": {
          "streamEnabled": true,
          "streamViewType": "NEW_AND_OLD_IMAGES"
        },
        // Regional replicas that make up the global table
        "replicas": [
          {
            "regionName": "us-east-1"
          },
          {
            "regionName": "eu-west-1"
          }
        ]
      }
    }
  }
}
```
