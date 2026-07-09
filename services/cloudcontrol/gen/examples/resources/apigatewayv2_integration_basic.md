A basic AWS ApiGatewayV2 Integration with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource integration: aws/apigatewayv2/integration {
    metadata {
        displayName = "AWS ApiGatewayV2 Integration basic"
    }
    spec {
        apiId = "example-api-id"
        integrationType = "example-integration-type"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    integration:
        type: aws/apigatewayv2/integration
        metadata:
            displayName: AWS ApiGatewayV2 Integration basic
        spec:
            apiId: example-api-id
            integrationType: example-integration-type
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "integration": {
      "type": "aws/apigatewayv2/integration",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 Integration basic"
      },
      "spec": {
        "apiId": "example-api-id",
        "integrationType": "example-integration-type"
      }
    }
  }
}
```
