Look up an existing AWS DynamoDB Table by tableName and export its arn.

```blueprintlang
version "2025-11-02"

data exampleTable: aws/dynamodb/table {
    filter "tableName" == "example-tablename"

    export arn: string
    export billingMode: string
    export deletionProtectionEnabled: boolean
}

export exampleTableArn: string {
    field = datasources.exampleTable.arn
}
```

```yaml
version: 2025-11-02

datasources:
  exampleTable:
    type: aws/dynamodb/table
    filter:
      field: tableName
      operator: "="
      search: example-tablename
    exports:
      arn:
        type: string
      billingMode:
        type: string
      deletionProtectionEnabled:
        type: boolean

exports:
  exampleTableArn:
    type: string
    field: datasources.exampleTable.arn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleTable": {
      "type": "aws/dynamodb/table",
      "filter": { "field": "tableName", "operator": "=", "search": "example-tablename" },
      "exports": {
        "arn": { "type": "string" },
        "billingMode": { "type": "string" },
        "deletionProtectionEnabled": { "type": "boolean" }
      }
    }
  }
}
```
