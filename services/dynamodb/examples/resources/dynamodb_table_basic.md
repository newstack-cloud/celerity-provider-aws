This example demonstrates creating a basic DynamoDB table with on-demand billing.

```blueprintlang
version "2025-11-02"

resource usersTable: aws/dynamodb/table {
    metadata {
        displayName = "Users Table"
    }
    spec {
        tableName = "users"
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
    }
}
```

```yaml
version: 2025-11-02

resources:
  usersTable:
    type: aws/dynamodb/table
    metadata:
      displayName: Users Table
    spec:
      tableName: users
      attributeDefinitions:
        - attributeName: id
          attributeType: S
      keySchema:
        - attributeName: id
          keyType: HASH
      billingMode: PAY_PER_REQUEST
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "usersTable": {
      "type": "aws/dynamodb/table",
      "metadata": {
        "displayName": "Users Table"
      },
      "spec": {
        // Table name - must be unique within the region
        "tableName": "users",
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
        // Use on-demand billing for unpredictable workloads
        "billingMode": "PAY_PER_REQUEST"
      }
    }
  }
}
```
