**DynamoDB Table - JSONC**

This example shows how to create a DynamoDB table using JSONC format with comments.

```javascript
{
  "resources": {
    "sessionsTable": {
      "type": "aws/dynamodb/table",
      "metadata": {
        "displayName": "Sessions Table with JSONC"
      },
      "spec": {
        // Table name - must be unique within the region
        "tableName": "sessions",

        // Define attributes used in key schema and indexes
        "attributeDefinitions": [
          {
            "attributeName": "sessionId",
            "attributeType": "S"  // String
          },
          {
            "attributeName": "userId",
            "attributeType": "S"
          },
          {
            "attributeName": "createdAt",
            "attributeType": "N"  // Number (Unix timestamp)
          }
        ],

        // Primary key: sessionId as partition key
        "keySchema": [
          {
            "attributeName": "sessionId",
            "keyType": "HASH"  // Partition key
          }
        ],

        // Use on-demand billing for unpredictable workloads
        "billingMode": "PAY_PER_REQUEST",

        // GSI to query sessions by user
        "globalSecondaryIndexes": [
          {
            "indexName": "userId-createdAt-index",
            "keySchema": [
              {
                "attributeName": "userId",
                "keyType": "HASH"
              },
              {
                "attributeName": "createdAt",
                "keyType": "RANGE"  // Sort key for time-based queries
              }
            ],
            "projection": {
              // Include all attributes in the index
              "projectionType": "ALL"
            }
          }
        ],

        // Enable streams for real-time processing
        "streamSpecification": {
          "streamEnabled": true,
          // Capture both old and new item images
          "streamViewType": "NEW_AND_OLD_IMAGES"
        },

        // Auto-expire sessions after TTL
        "timeToLiveSpecification": {
          "enabled": true,
          "attributeName": "expiresAt"
        },

        // Enable point-in-time recovery for data protection
        "pointInTimeRecoverySpecification": {
          "pointInTimeRecoveryEnabled": true
        },

        // Use SQS-managed encryption
        "sseSpecification": {
          "enabled": true
        },

        // Tags for resource management
        "tags": [
          {
            "key": "Application",
            "value": "UserAuth"
          },
          {
            "key": "Environment",
            "value": "Production"
          }
        ]
      }
    }
  }
}
```
