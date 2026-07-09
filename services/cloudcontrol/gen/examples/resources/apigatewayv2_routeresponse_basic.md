A basic AWS ApiGatewayV2 RouteResponse with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource routeResponse: aws/apigatewayv2/routeResponse {
    metadata {
        displayName = "AWS ApiGatewayV2 RouteResponse basic"
    }
    spec {
        apiId = "example-api-id"
        routeId = "example-route-id"
        routeResponseKey = "example-route-response-key"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    routeResponse:
        type: aws/apigatewayv2/routeResponse
        metadata:
            displayName: AWS ApiGatewayV2 RouteResponse basic
        spec:
            apiId: example-api-id
            routeId: example-route-id
            routeResponseKey: example-route-response-key
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "routeResponse": {
      "type": "aws/apigatewayv2/routeResponse",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 RouteResponse basic"
      },
      "spec": {
        "apiId": "example-api-id",
        "routeId": "example-route-id",
        "routeResponseKey": "example-route-response-key"
      }
    }
  }
}
```
