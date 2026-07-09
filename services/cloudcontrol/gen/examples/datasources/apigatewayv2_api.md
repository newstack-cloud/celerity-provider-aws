Look up an existing AWS ApiGatewayV2 Api by apiId and export its apiEndpoint.

```blueprintlang
version "2025-11-02"

data exampleApi: aws/apigatewayv2/api {
    filter "apiId" == "example-apiid"

    export apiEndpoint: string
    export apiId: string
    export apiKeySelectionExpression: string
}

export exampleApiApiEndpoint: string {
    field = datasources.exampleApi.apiEndpoint
}
```

```yaml
version: 2025-11-02

datasources:
  exampleApi:
    type: aws/apigatewayv2/api
    filter:
      field: apiId
      operator: "="
      search: example-apiid
    exports:
      apiEndpoint:
        type: string
      apiId:
        type: string
      apiKeySelectionExpression:
        type: string

exports:
  exampleApiApiEndpoint:
    type: string
    field: datasources.exampleApi.apiEndpoint
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleApi": {
      "type": "aws/apigatewayv2/api",
      "filter": { "field": "apiId", "operator": "=", "search": "example-apiid" },
      "exports": {
        "apiEndpoint": { "type": "string" },
        "apiId": { "type": "string" },
        "apiKeySelectionExpression": { "type": "string" }
      }
    }
  }
}
```
