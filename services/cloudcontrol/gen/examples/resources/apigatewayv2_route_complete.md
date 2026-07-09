A AWS ApiGatewayV2 Route configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource route: aws/apigatewayv2/route {
    metadata {
        displayName = "AWS ApiGatewayV2 Route complete"
    }
    spec {
        apiId = "example-api-id"
        apiKeyRequired = false
        authorizationScopes = [
            "example-authorization-scope"
        ]
        authorizationType = "example-authorization-type"
        authorizerId = "example-authorizer-id"
        modelSelectionExpression = "example-model-selection-expression"
        operationName = "example-operation-name"
        requestModels = {
            exampleKey = "example-value"
        }
        requestParameters = {
            exampleKey = "example-value"
        }
        routeKey = "example-route-key"
        routeResponseSelectionExpression = "example-route-response-selection-expression"
        target = "example-target"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    route:
        type: aws/apigatewayv2/route
        metadata:
            displayName: AWS ApiGatewayV2 Route complete
        spec:
            apiId: example-api-id
            apiKeyRequired: false
            authorizationScopes:
                - example-authorization-scope
            authorizationType: example-authorization-type
            authorizerId: example-authorizer-id
            modelSelectionExpression: example-model-selection-expression
            operationName: example-operation-name
            requestModels:
                exampleKey: example-value
            requestParameters:
                exampleKey: example-value
            routeKey: example-route-key
            routeResponseSelectionExpression: example-route-response-selection-expression
            target: example-target
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "route": {
      "type": "aws/apigatewayv2/route",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 Route complete"
      },
      "spec": {
        "apiId": "example-api-id",
        "apiKeyRequired": false,
        "authorizationScopes": [
          "example-authorization-scope"
        ],
        "authorizationType": "example-authorization-type",
        "authorizerId": "example-authorizer-id",
        "modelSelectionExpression": "example-model-selection-expression",
        "operationName": "example-operation-name",
        "requestModels": {
          "exampleKey": "example-value"
        },
        "requestParameters": {
          "exampleKey": "example-value"
        },
        "routeKey": "example-route-key",
        "routeResponseSelectionExpression": "example-route-response-selection-expression",
        "target": "example-target"
      }
    }
  }
}
```
