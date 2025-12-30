# DynamoDB Table Data Source (YAML)

Retrieves information about an existing DynamoDB table by name.

```yaml
dataSources:
  myTable:
    type: aws/dynamodb/table
    filter:
      tableName: my-application-table
    exports:
      tableArn: ${arn}
      streamArn: ${latestStreamArn}
```
