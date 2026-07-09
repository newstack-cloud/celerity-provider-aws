A AWS ApiGatewayV2 RouteResponse configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource routeResponse: aws/apigatewayv2/routeResponse {
    metadata {
        displayName = "AWS ApiGatewayV2 RouteResponse complete"
    }
    spec {
        apiId = "example-api-id"
        modelSelectionExpression = "example-model-selection-expression"
        responseModels = {
            exampleKey = "example-value"
        }
        responseParameters = "example-response-parameters"
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
            displayName: AWS ApiGatewayV2 RouteResponse complete
        spec:
            apiId: example-api-id
            modelSelectionExpression: example-model-selection-expression
            responseModels:
                exampleKey: example-value
            responseParameters: example-response-parameters
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
        "displayName": "AWS ApiGatewayV2 RouteResponse complete"
      },
      "spec": {
        "apiId": "example-api-id",
        "modelSelectionExpression": "example-model-selection-expression",
        "responseModels": {
          "exampleKey": "example-value"
        },
        "responseParameters": "example-response-parameters",
        "routeId": "example-route-id",
        "routeResponseKey": "example-route-response-key"
      }
    }
  }
}
```
