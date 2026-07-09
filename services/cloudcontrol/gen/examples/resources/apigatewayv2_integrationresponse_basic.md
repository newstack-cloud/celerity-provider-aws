A basic AWS ApiGatewayV2 IntegrationResponse with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource integrationResponse: aws/apigatewayv2/integrationResponse {
    metadata {
        displayName = "AWS ApiGatewayV2 IntegrationResponse basic"
    }
    spec {
        apiId = "example-api-id"
        integrationId = "example-integration-id"
        integrationResponseKey = "example-integration-response-key"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    integrationResponse:
        type: aws/apigatewayv2/integrationResponse
        metadata:
            displayName: AWS ApiGatewayV2 IntegrationResponse basic
        spec:
            apiId: example-api-id
            integrationId: example-integration-id
            integrationResponseKey: example-integration-response-key
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "integrationResponse": {
      "type": "aws/apigatewayv2/integrationResponse",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 IntegrationResponse basic"
      },
      "spec": {
        "apiId": "example-api-id",
        "integrationId": "example-integration-id",
        "integrationResponseKey": "example-integration-response-key"
      }
    }
  }
}
```
