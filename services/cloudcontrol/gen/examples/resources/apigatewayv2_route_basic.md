A basic AWS ApiGatewayV2 Route with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource route: aws/apigatewayv2/route {
    metadata {
        displayName = "AWS ApiGatewayV2 Route basic"
    }
    spec {
        apiId = "example-api-id"
        routeKey = "example-route-key"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    route:
        type: aws/apigatewayv2/route
        metadata:
            displayName: AWS ApiGatewayV2 Route basic
        spec:
            apiId: example-api-id
            routeKey: example-route-key
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "route": {
      "type": "aws/apigatewayv2/route",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 Route basic"
      },
      "spec": {
        "apiId": "example-api-id",
        "routeKey": "example-route-key"
      }
    }
  }
}
```
