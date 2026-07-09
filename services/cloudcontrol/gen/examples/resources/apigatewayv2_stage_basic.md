A basic AWS ApiGatewayV2 Stage with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource stage: aws/apigatewayv2/stage {
    metadata {
        displayName = "AWS ApiGatewayV2 Stage basic"
    }
    spec {
        apiId = "example-api-id"
        stageName = "example-stage-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    stage:
        type: aws/apigatewayv2/stage
        metadata:
            displayName: AWS ApiGatewayV2 Stage basic
        spec:
            apiId: example-api-id
            stageName: example-stage-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "stage": {
      "type": "aws/apigatewayv2/stage",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 Stage basic"
      },
      "spec": {
        "apiId": "example-api-id",
        "stageName": "example-stage-name"
      }
    }
  }
}
```
