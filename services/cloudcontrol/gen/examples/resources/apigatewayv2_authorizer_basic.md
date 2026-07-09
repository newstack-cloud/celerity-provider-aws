A basic AWS ApiGatewayV2 Authorizer with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource authorizer: aws/apigatewayv2/authorizer {
    metadata {
        displayName = "AWS ApiGatewayV2 Authorizer basic"
    }
    spec {
        apiId = "example-api-id"
        authorizerType = "example-authorizer-type"
        name = "example-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    authorizer:
        type: aws/apigatewayv2/authorizer
        metadata:
            displayName: AWS ApiGatewayV2 Authorizer basic
        spec:
            apiId: example-api-id
            authorizerType: example-authorizer-type
            name: example-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "authorizer": {
      "type": "aws/apigatewayv2/authorizer",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 Authorizer basic"
      },
      "spec": {
        "apiId": "example-api-id",
        "authorizerType": "example-authorizer-type",
        "name": "example-name"
      }
    }
  }
}
```
