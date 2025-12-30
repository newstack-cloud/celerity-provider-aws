# DynamoDB Table Data Source (JSONC)

Retrieves information about an existing DynamoDB table by ARN.

```jsonc
{
  "dataSources": {
    "myTable": {
      "type": "aws/dynamodb/table",
      "filter": {
        "arn": "arn:aws:dynamodb:us-east-1:123456789012:table/my-table"
      },
      "exports": {
        "tableName": "${tableName}",
        "tableArn": "${arn}",
        "billingMode": "${billingMode}"
      }
    }
  }
}
```
