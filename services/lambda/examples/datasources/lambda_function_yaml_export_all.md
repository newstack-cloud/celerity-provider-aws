Look up an existing Lambda function by ARN and export all of its fields for use elsewhere in the blueprint.

```blueprintlang
version "2025-11-02"

data orderFunction: aws/lambda/function {
    filter "arn" == "arn:aws:lambda:us-east-1:123456789012:function:order-retrieval"

    export *
}

export orderFunctionArn: string {
    field = datasources.orderFunction.arn
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
    exports: "*"

exports:
  orderFunctionArn:
    type: string
    field: datasources.orderFunction.arn
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
      "exports": "*"
    }
  },
  "exports": {
    "orderFunctionArn": {
      "type": "string",
      "field": "datasources.orderFunction.arn"
    }
  }
}
```
