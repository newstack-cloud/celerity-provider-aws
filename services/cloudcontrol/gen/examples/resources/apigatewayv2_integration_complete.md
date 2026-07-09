A AWS ApiGatewayV2 Integration configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource integration: aws/apigatewayv2/integration {
    metadata {
        displayName = "AWS ApiGatewayV2 Integration complete"
    }
    spec {
        apiId = "example-api-id"
        connectionId = "example-connection-id"
        connectionType = "example-connection-type"
        contentHandlingStrategy = "example-content-handling-strategy"
        credentialsArn = "example-credentials-arn"
        description = "example-description"
        integrationMethod = "example-integration-method"
        integrationSubtype = "example-integration-subtype"
        integrationType = "example-integration-type"
        integrationUri = "example-integration-uri"
        passthroughBehavior = "example-passthrough-behavior"
        payloadFormatVersion = "example-payload-format-version"
        requestParameters = {
            example = "example-request-parameters"
        }
        requestTemplates = {
            example = "example-request-templates"
        }
        responseParameters = {
            example = {
                responseParameters = [
                    {
                        destination = "example-destination",
                        source = "example-source"
                    }
                ]
            }
        }
        templateSelectionExpression = "example-template-selection-expression"
        timeoutInMillis = 1
        tlsConfig = {
            serverNameToVerify = "example-server-name-to-verify"
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    integration:
        type: aws/apigatewayv2/integration
        metadata:
            displayName: AWS ApiGatewayV2 Integration complete
        spec:
            apiId: example-api-id
            connectionId: example-connection-id
            connectionType: example-connection-type
            contentHandlingStrategy: example-content-handling-strategy
            credentialsArn: example-credentials-arn
            description: example-description
            integrationMethod: example-integration-method
            integrationSubtype: example-integration-subtype
            integrationType: example-integration-type
            integrationUri: example-integration-uri
            passthroughBehavior: example-passthrough-behavior
            payloadFormatVersion: example-payload-format-version
            requestParameters:
                example: example-request-parameters
            requestTemplates:
                example: example-request-templates
            responseParameters:
                example:
                    responseParameters:
                        - destination: example-destination
                          source: example-source
            templateSelectionExpression: example-template-selection-expression
            timeoutInMillis: 1
            tlsConfig:
                serverNameToVerify: example-server-name-to-verify
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "integration": {
      "type": "aws/apigatewayv2/integration",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 Integration complete"
      },
      "spec": {
        "apiId": "example-api-id",
        "connectionId": "example-connection-id",
        "connectionType": "example-connection-type",
        "contentHandlingStrategy": "example-content-handling-strategy",
        "credentialsArn": "example-credentials-arn",
        "description": "example-description",
        "integrationMethod": "example-integration-method",
        "integrationSubtype": "example-integration-subtype",
        "integrationType": "example-integration-type",
        "integrationUri": "example-integration-uri",
        "passthroughBehavior": "example-passthrough-behavior",
        "payloadFormatVersion": "example-payload-format-version",
        "requestParameters": {
          "example": "example-request-parameters"
        },
        "requestTemplates": {
          "example": "example-request-templates"
        },
        "responseParameters": {
          "example": {
            "responseParameters": [
              {
                "destination": "example-destination",
                "source": "example-source"
              }
            ]
          }
        },
        "templateSelectionExpression": "example-template-selection-expression",
        "timeoutInMillis": 1,
        "tlsConfig": {
          "serverNameToVerify": "example-server-name-to-verify"
        }
      }
    }
  }
}
```
