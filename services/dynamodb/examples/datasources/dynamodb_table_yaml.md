Look up an existing DynamoDB table by name and export its ARN and latest stream ARN.

```blueprintlang
version "2025-11-02"

data applicationTable: aws/dynamodb/table {
    filter "tableName" == "my-application-table"

    export arn: string
    export latestStreamArn: string
}

export applicationTableArn: string {
    field = datasources.applicationTable.arn
}
```

```yaml
version: 2025-11-02

datasources:
  applicationTable:
    type: aws/dynamodb/table
    filter:
      field: tableName
      operator: "="
      search: my-application-table
    exports:
      arn:
        type: string
      latestStreamArn:
        type: string

exports:
  applicationTableArn:
    type: string
    field: datasources.applicationTable.arn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "applicationTable": {
      "type": "aws/dynamodb/table",
      "filter": {
        "field": "tableName",
        "operator": "=",
        "search": "my-application-table"
      },
      "exports": {
        "arn": { "type": "string" },
        "latestStreamArn": { "type": "string" }
      }
    }
  },
  "exports": {
    "applicationTableArn": {
      "type": "string",
      "field": "datasources.applicationTable.arn"
    }
  }
}
```
