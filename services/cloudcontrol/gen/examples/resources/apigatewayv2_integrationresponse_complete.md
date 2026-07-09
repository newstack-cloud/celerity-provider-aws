A AWS ApiGatewayV2 IntegrationResponse configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource integrationResponse: aws/apigatewayv2/integrationResponse {
    metadata {
        displayName = "AWS ApiGatewayV2 IntegrationResponse complete"
    }
    spec {
        apiId = "example-api-id"
        contentHandlingStrategy = "example-content-handling-strategy"
        integrationId = "example-integration-id"
        integrationResponseKey = "example-integration-response-key"
        responseParameters = {
            exampleKey = "example-value"
        }
        responseTemplates = {
            exampleKey = "example-value"
        }
        templateSelectionExpression = "example-template-selection-expression"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    integrationResponse:
        type: aws/apigatewayv2/integrationResponse
        metadata:
            displayName: AWS ApiGatewayV2 IntegrationResponse complete
        spec:
            apiId: example-api-id
            contentHandlingStrategy: example-content-handling-strategy
            integrationId: example-integration-id
            integrationResponseKey: example-integration-response-key
            responseParameters:
                exampleKey: example-value
            responseTemplates:
                exampleKey: example-value
            templateSelectionExpression: example-template-selection-expression
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "integrationResponse": {
      "type": "aws/apigatewayv2/integrationResponse",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 IntegrationResponse complete"
      },
      "spec": {
        "apiId": "example-api-id",
        "contentHandlingStrategy": "example-content-handling-strategy",
        "integrationId": "example-integration-id",
        "integrationResponseKey": "example-integration-response-key",
        "responseParameters": {
          "exampleKey": "example-value"
        },
        "responseTemplates": {
          "exampleKey": "example-value"
        },
        "templateSelectionExpression": "example-template-selection-expression"
      }
    }
  }
}
```
