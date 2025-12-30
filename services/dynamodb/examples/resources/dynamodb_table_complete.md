**DynamoDB Table - Complete**

This example demonstrates creating a DynamoDB table with all available configuration options.

```yaml
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
