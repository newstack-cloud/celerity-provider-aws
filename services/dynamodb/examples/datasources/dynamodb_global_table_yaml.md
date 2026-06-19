Look up an existing DynamoDB global table by name and export its ARN and billing mode.

```blueprintlang
version "2025-11-02"

data ordersGlobalTable: aws/dynamodb/globalTable {
    filter "tableName" == "my-global-table"

    export arn: string
    export billingMode: string
}

export ordersGlobalTableArn: string {
    field = datasources.ordersGlobalTable.arn
}
```

```yaml
version: 2025-11-02

datasources:
  ordersGlobalTable:
    type: aws/dynamodb/globalTable
    filter:
      field: tableName
      operator: "="
      search: my-global-table
    exports:
      arn:
        type: string
      billingMode:
        type: string

exports:
  ordersGlobalTableArn:
    type: string
    field: datasources.ordersGlobalTable.arn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "ordersGlobalTable": {
      "type": "aws/dynamodb/globalTable",
      "filter": {
        "field": "tableName",
        "operator": "=",
        "search": "my-global-table"
      },
      "exports": {
        "arn": { "type": "string" },
        "billingMode": { "type": "string" }
      }
    }
  },
  "exports": {
    "ordersGlobalTableArn": {
      "type": "string",
      "field": "datasources.ordersGlobalTable.arn"
    }
  }
}
```
