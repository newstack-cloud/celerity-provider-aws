## DynamoDB Table to Lambda Function Link (Stream Trigger)

This link creates an Event Source Mapping that triggers a Lambda function when records are written to a DynamoDB table's stream. This enables real-time processing of data changes in DynamoDB.

The link automatically:
1. **Enables DynamoDB Streams** on the table if not already enabled, using the configured `streamViewType` annotation (defaults to `NEW_AND_OLD_IMAGES`)
2. **Adds stream read permissions** (`dynamodb:DescribeStream`, `dynamodb:GetRecords`, `dynamodb:GetShardIterator`, `dynamodb:ListStreams`) to the Lambda function's execution role
3. **Creates an Event Source Mapping** to trigger the Lambda function when records are written to the stream

### Requirements

The Lambda function's execution role must be defined in the same blueprint.

### Annotation Placement

- **Table annotations** (`aws.dynamodb.stream.viewType`): Configure the DynamoDB table's stream format (resource configuration). This affects all consumers of the stream.
- **Function annotations** (`aws.dynamodb.lambda.stream.batchSize`, `aws.dynamodb.lambda.stream.startingPosition`, etc.): Configure the event source mapping for this specific function (relationship configuration). Each function triggered by the same table can have different settings. Note that relationship annotations include both service names and the feature (`dynamodb.lambda.stream`).

### Example

```yaml
resources:
  ordersTable:
    type: aws/dynamodb/table
    metadata:
      labels:
        table: orders
      annotations:
        # viewType is on the table because it configures the stream format
        # (affects all functions that consume from this stream)
        aws.dynamodb.stream.viewType: NEW_AND_OLD_IMAGES
    linkSelector:
      byLabel:
        processor: orders
    spec:
      tableName: orders-table
      # No need to specify streamSpecification - the link enables it automatically

  orderProcessor:
    type: aws/lambda/function
    metadata:
      labels:
        processor: orders
      annotations:
        # Event source mapping config is on the function because it's specific
        # to how THIS function processes records from the stream.
        # Relationship annotations use aws.dynamodb.lambda.stream.* to indicate they
        # configure the DynamoDB→Lambda stream relationship.
        aws.dynamodb.lambda.stream.startingPosition: LATEST
        aws.dynamodb.lambda.stream.batchSize: 50
        aws.dynamodb.lambda.stream.batchWindow: 5
        aws.dynamodb.lambda.stream.enabled: true
    spec:
      functionName: order-processor
      role: ${resources.orderProcessorRole.state.arn}
      # ... other function configuration

  orderProcessorRole:
    type: aws/iam/role
    spec:
      name: order-processor-role
      # Stream read permissions are automatically added by the link
```

In this example:
- The DynamoDB table links to the Lambda function via label selector
- The link automatically enables DynamoDB Streams with `NEW_AND_OLD_IMAGES` view type (configured on the table)
- When records are written to the orders table, Lambda is triggered with batches of up to 50 records (configured on the function)
- The function processes only new records (LATEST starting position)
- Stream read permissions are automatically added to the function's execution role
