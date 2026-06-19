Look up an existing Lambda function by ARN and export its name and qualified ARN.

```blueprintlang
version "2025-11-02"

data orderFunction: aws/lambda/function {
    filter "arn" == "arn:aws:lambda:us-east-1:123456789012:function:order-retrieval"

    export name: string
    export qualifiedArn: string
}

export orderFunctionQualifiedArn: string {
    field = datasources.orderFunction.qualifiedArn
}
```

```yaml
version: 2025-11-02

datasources:
  orderFunction:
    type: aws/lambda/function
    filter:
      field: arn
      operator: "="
      search: arn:aws:lambda:us-east-1:123456789012:function:order-retrieval
    exports:
      name:
        type: string
      qualifiedArn:
        type: string

exports:
  orderFunctionQualifiedArn:
    type: string
    field: datasources.orderFunction.qualifiedArn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "orderFunction": {
      "type": "aws/lambda/function",
      "filter": {
        "field": "arn",
        "operator": "=",
        "search": "arn:aws:lambda:us-east-1:123456789012:function:order-retrieval"
      },
      "exports": {
        "name": { "type": "string" },
        "qualifiedArn": { "type": "string" }
      }
    }
  },
  "exports": {
    "orderFunctionQualifiedArn": {
      "type": "string",
      "field": "datasources.orderFunction.qualifiedArn"
    }
  }
}
```
