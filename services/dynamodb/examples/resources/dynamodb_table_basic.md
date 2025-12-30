**DynamoDB Table - Basic**

This example demonstrates creating a basic DynamoDB table with on-demand billing.

```yaml
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
