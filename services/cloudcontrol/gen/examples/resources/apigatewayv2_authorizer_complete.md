A AWS ApiGatewayV2 Authorizer configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource authorizer: aws/apigatewayv2/authorizer {
    metadata {
        displayName = "AWS ApiGatewayV2 Authorizer complete"
    }
    spec {
        apiId = "example-api-id"
        authorizerCredentialsArn = "example-authorizer-credentials-arn"
        authorizerPayloadFormatVersion = "example-authorizer-payload-format-version"
        authorizerResultTtlInSeconds = 1
        authorizerType = "example-authorizer-type"
        authorizerUri = "example-authorizer-uri"
        enableSimpleResponses = false
        identitySource = [
            "example-identity-source"
        ]
        identityValidationExpression = "example-identity-validation-expression"
        jwtConfiguration = {
            audience = [
                "example-audience"
            ],
            issuer = "example-issuer"
        }
        name = "example-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    authorizer:
        type: aws/apigatewayv2/authorizer
        metadata:
            displayName: AWS ApiGatewayV2 Authorizer complete
        spec:
            apiId: example-api-id
            authorizerCredentialsArn: example-authorizer-credentials-arn
            authorizerPayloadFormatVersion: example-authorizer-payload-format-version
            authorizerResultTtlInSeconds: 1
            authorizerType: example-authorizer-type
            authorizerUri: example-authorizer-uri
            enableSimpleResponses: false
            identitySource:
                - example-identity-source
            identityValidationExpression: example-identity-validation-expression
            jwtConfiguration:
                audience:
                    - example-audience
                issuer: example-issuer
            name: example-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "authorizer": {
      "type": "aws/apigatewayv2/authorizer",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 Authorizer complete"
      },
      "spec": {
        "apiId": "example-api-id",
        "authorizerCredentialsArn": "example-authorizer-credentials-arn",
        "authorizerPayloadFormatVersion": "example-authorizer-payload-format-version",
        "authorizerResultTtlInSeconds": 1,
        "authorizerType": "example-authorizer-type",
        "authorizerUri": "example-authorizer-uri",
        "enableSimpleResponses": false,
        "identitySource": [
          "example-identity-source"
        ],
        "identityValidationExpression": "example-identity-validation-expression",
        "jwtConfiguration": {
          "audience": [
            "example-audience"
          ],
          "issuer": "example-issuer"
        },
        "name": "example-name"
      }
    }
  }
}
```
