A AWS ApiGatewayV2 ApiMapping configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource apiMapping: aws/apigatewayv2/apiMapping {
    metadata {
        displayName = "AWS ApiGatewayV2 ApiMapping complete"
    }
    spec {
        apiId = "example-api-id"
        apiMappingKey = "example-api-mapping-key"
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
            displayName: AWS ApiGatewayV2 ApiMapping complete
        spec:
            apiId: example-api-id
            apiMappingKey: example-api-mapping-key
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
        "displayName": "AWS ApiGatewayV2 ApiMapping complete"
      },
      "spec": {
        "apiId": "example-api-id",
        "apiMappingKey": "example-api-mapping-key",
        "domainName": "example-domain-name",
        "stage": "example-stage"
      }
    }
  }
}
```
