A AWS ApiGatewayV2 Api configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource api: aws/apigatewayv2/api {
    metadata {
        displayName = "AWS ApiGatewayV2 Api complete"
    }
    spec {
        apiKeySelectionExpression = "example-api-key-selection-expression"
        basePath = "example-base-path"
        body = {
            exampleKey = "example-value"
        }
        bodyS3Location = {
            bucket = "example-bucket",
            etag = "example-etag",
            key = "example-key",
            version = "example-version"
        }
        corsConfiguration = {
            allowCredentials = false,
            allowHeaders = [
                "example-allow-header"
            ],
            allowMethods = [
                "example-allow-method"
            ],
            allowOrigins = [
                "example-allow-origin"
            ],
            exposeHeaders = [
                "example-expose-header"
            ],
            maxAge = 1
        }
        credentialsArn = "example-credentials-arn"
        description = "example-description"
        disableExecuteApiEndpoint = false
        disableSchemaValidation = false
        failOnWarnings = false
        ipAddressType = "example-ip-address-type"
        name = "example-name"
        protocolType = "example-protocol-type"
        routeKey = "example-route-key"
        routeSelectionExpression = "example-route-selection-expression"
        tags = {
            example = "example-tags"
        }
        target = "example-target"
        version = "example-version"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    api:
        type: aws/apigatewayv2/api
        metadata:
            displayName: AWS ApiGatewayV2 Api complete
        spec:
            apiKeySelectionExpression: example-api-key-selection-expression
            basePath: example-base-path
            body:
                exampleKey: example-value
            bodyS3Location:
                bucket: example-bucket
                etag: example-etag
                key: example-key
                version: example-version
            corsConfiguration:
                allowCredentials: false
                allowHeaders:
                    - example-allow-header
                allowMethods:
                    - example-allow-method
                allowOrigins:
                    - example-allow-origin
                exposeHeaders:
                    - example-expose-header
                maxAge: 1
            credentialsArn: example-credentials-arn
            description: example-description
            disableExecuteApiEndpoint: false
            disableSchemaValidation: false
            failOnWarnings: false
            ipAddressType: example-ip-address-type
            name: example-name
            protocolType: example-protocol-type
            routeKey: example-route-key
            routeSelectionExpression: example-route-selection-expression
            tags:
                example: example-tags
            target: example-target
            version: example-version
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "api": {
      "type": "aws/apigatewayv2/api",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 Api complete"
      },
      "spec": {
        "apiKeySelectionExpression": "example-api-key-selection-expression",
        "basePath": "example-base-path",
        "body": {
          "exampleKey": "example-value"
        },
        "bodyS3Location": {
          "bucket": "example-bucket",
          "etag": "example-etag",
          "key": "example-key",
          "version": "example-version"
        },
        "corsConfiguration": {
          "allowCredentials": false,
          "allowHeaders": [
            "example-allow-header"
          ],
          "allowMethods": [
            "example-allow-method"
          ],
          "allowOrigins": [
            "example-allow-origin"
          ],
          "exposeHeaders": [
            "example-expose-header"
          ],
          "maxAge": 1
        },
        "credentialsArn": "example-credentials-arn",
        "description": "example-description",
        "disableExecuteApiEndpoint": false,
        "disableSchemaValidation": false,
        "failOnWarnings": false,
        "ipAddressType": "example-ip-address-type",
        "name": "example-name",
        "protocolType": "example-protocol-type",
        "routeKey": "example-route-key",
        "routeSelectionExpression": "example-route-selection-expression",
        "tags": {
          "example": "example-tags"
        },
        "target": "example-target",
        "version": "example-version"
      }
    }
  }
}
```
