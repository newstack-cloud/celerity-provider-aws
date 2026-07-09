A basic AWS ApiGatewayV2 ApiMapping with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource apiMapping: aws/apigatewayv2/apiMapping {
    metadata {
        displayName = "AWS ApiGatewayV2 ApiMapping basic"
    }
    spec {
        apiId = "example-api-id"
        domainName = "example-domain-name"
        stage = "example-stage"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    apiMapping:
        type: aws/apigatewayv2/apiMapping
        metadata:
            displayName: AWS ApiGatewayV2 ApiMapping basic
        spec:
            apiId: example-api-id
            domainName: example-domain-name
            stage: example-stage
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "apiMapping": {
      "type": "aws/apigatewayv2/apiMapping",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 ApiMapping basic"
      },
      "spec": {
        "apiId": "example-api-id",
        "domainName": "example-domain-name",
        "stage": "example-stage"
      }
    }
  }
}
```
