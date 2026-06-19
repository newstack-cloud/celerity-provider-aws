Look up the function URL for an existing Lambda function and export its URL and auth type.

```blueprintlang
version "2025-11-02"

data functionUrl: aws/lambda/function_url {
    filter "functionName" == "order-processor"

    export functionUrl: string
    export authType: string
}

export orderProcessorUrl: string {
    field = datasources.functionUrl.functionUrl
}
```

```yaml
version: 2025-11-02

datasources:
  functionUrl:
    type: aws/lambda/function_url
    filter:
      field: functionName
      operator: "="
      search: order-processor
    exports:
      functionUrl:
        type: string
      authType:
        type: string

exports:
  orderProcessorUrl:
    type: string
    field: datasources.functionUrl.functionUrl
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "functionUrl": {
      "type": "aws/lambda/function_url",
      "filter": {
        "field": "functionName",
        "operator": "=",
        "search": "order-processor"
      },
      "exports": {
        "functionUrl": { "type": "string" },
        "authType": { "type": "string" }
      }
    }
  },
  "exports": {
    "orderProcessorUrl": {
      "type": "string",
      "field": "datasources.functionUrl.functionUrl"
    }
  }
}
```
