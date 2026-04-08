## Lambda Function to DynamoDB Table Link

This link configures a Lambda function to access a DynamoDB table by:

1. **Environment Variables** (optional): Populates environment variables in the Lambda function with the table name/ARN. This allows the function code to easily reference the table without hardcoding values.

2. **IAM Permissions**: Adds an inline policy to the Lambda function's execution role with DynamoDB permissions based on the configured access level (read, write, or readwrite).

### Requirements

The Lambda function's execution role must be defined in the same blueprint. The link will look up the role using the function's `spec.role` field and add an inline policy with the required DynamoDB permissions.

### Example

```yaml
resources:
  myFunction:
    type: aws/lambda/function
    metadata:
      labels:
        app: my-app
      annotations:
        aws.lambda.dynamodb.ordersTable.accessLevel: read
        aws.lambda.dynamodb.ordersTable.envVarName: ORDERS_TABLE_NAME
    linkSelector:
      byLabel:
        table: orders
    spec:
      functionName: my-function
      role: ${resources.myFunctionRole.state.arn}
      # ... other function configuration

  ordersTable:
    type: aws/dynamodb/table
    metadata:
      labels:
        table: orders
    spec:
      tableName: orders-table
      # ... other table configuration

  myFunctionRole:
    type: aws/iam/role
    spec:
      name: my-function-role
      # ... role configuration
```

In this example:
- The Lambda function links to the DynamoDB table via label selector
- The function gets read-only access to the table
- An environment variable `ORDERS_TABLE_NAME` is populated with the table name
- The function's execution role receives an inline policy with DynamoDB read permissions
