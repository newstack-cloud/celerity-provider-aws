A basic AWS ApiGatewayV2 Api with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource api: aws/apigatewayv2/api {
    metadata {
        displayName = "AWS ApiGatewayV2 Api basic"
    }
    spec {
        apiKeySelectionExpression = "example-api-key-selection-expression"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    api:
        type: aws/apigatewayv2/api
        metadata:
            displayName: AWS ApiGatewayV2 Api basic
        spec:
            apiKeySelectionExpression: example-api-key-selection-expression
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "api": {
      "type": "aws/apigatewayv2/api",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 Api basic"
      },
      "spec": {
        "apiKeySelectionExpression": "example-api-key-selection-expression"
      }
    }
  }
}
```
