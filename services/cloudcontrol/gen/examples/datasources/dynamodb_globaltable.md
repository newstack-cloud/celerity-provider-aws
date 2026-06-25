Look up an existing AWS DynamoDB GlobalTable by tableName and export its arn.

```blueprintlang
version "2025-11-02"

data exampleGlobalTable: aws/dynamodb/globalTable {
    filter "tableName" == "example-tablename"

    export arn: string
    export billingMode: string
    export multiRegionConsistency: string
}

export exampleGlobalTableArn: string {
    field = datasources.exampleGlobalTable.arn
}
```

```yaml
version: 2025-11-02

datasources:
  exampleGlobalTable:
    type: aws/dynamodb/globalTable
    filter:
      field: tableName
      operator: "="
      search: example-tablename
    exports:
      arn:
        type: string
      billingMode:
        type: string
      multiRegionConsistency:
        type: string

exports:
  exampleGlobalTableArn:
    type: string
    field: datasources.exampleGlobalTable.arn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleGlobalTable": {
      "type": "aws/dynamodb/globalTable",
      "filter": { "field": "tableName", "operator": "=", "search": "example-tablename" },
      "exports": {
        "arn": { "type": "string" },
        "billingMode": { "type": "string" },
        "multiRegionConsistency": { "type": "string" }
      }
    }
  }
}
```
