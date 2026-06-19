Look up an existing Lambda alias by function name and alias name, exporting its version and invoke ARN.

```blueprintlang
version "2025-11-02"

data productionAlias: aws/lambda/alias {
    filter "functionName" == "order-processor"
    filter "name" == "production"

    export functionVersion: string
    export description: string
    export invokeArn: string
}

export productionAliasInvokeArn: string {
    field = datasources.productionAlias.invokeArn
}
```

```yaml
version: 2025-11-02

datasources:
  productionAlias:
    type: aws/lambda/alias
    filter:
      - field: functionName
        operator: "="
        search: order-processor
      - field: name
        operator: "="
        search: production
    exports:
      functionVersion:
        type: string
      description:
        type: string
      invokeArn:
        type: string

exports:
  productionAliasInvokeArn:
    type: string
    field: datasources.productionAlias.invokeArn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "productionAlias": {
      "type": "aws/lambda/alias",
      "filter": [
        { "field": "functionName", "operator": "=", "search": "order-processor" },
        { "field": "name", "operator": "=", "search": "production" }
      ],
      "exports": {
        "functionVersion": { "type": "string" },
        "description": { "type": "string" },
        "invokeArn": { "type": "string" }
      }
    }
  },
  "exports": {
    "productionAliasInvokeArn": {
      "type": "string",
      "field": "datasources.productionAlias.invokeArn"
    }
  }
}
```
