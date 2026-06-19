Create a Lambda function URL with IAM authentication, response streaming, and a full CORS configuration.

```blueprintlang
version "2025-11-02"

resource functionUrl: aws/lambda/functionUrl {
    metadata {
        displayName = "Function URL"
    }
    spec {
        targetFunctionArn = "my-function"
        authType = "AWS_IAM"
        qualifier = "PROD"
        invokeMode = "RESPONSE_STREAM"
        cors = {
            allowCredentials = true
            allowHeaders = [
                "Content-Type",
                "Authorization"
            ]
            allowMethods = [
                "GET",
                "POST",
                "PUT",
                "DELETE"
            ]
            allowOrigins = [
                "https://example.com",
                "https://app.example.com"
            ]
            exposeHeaders = [
                "X-Custom-Header"
            ]
            maxAge = 3600
        }
    }
}
```

```yaml
version: 2025-11-02

resources:
  functionUrl:
    type: aws/lambda/functionUrl
    metadata:
      displayName: Function URL
    spec:
      targetFunctionArn: my-function
      authType: AWS_IAM
      qualifier: PROD
      invokeMode: RESPONSE_STREAM
      cors:
        allowCredentials: true
        allowHeaders:
          - Content-Type
          - Authorization
        allowMethods:
          - GET
          - POST
          - PUT
          - DELETE
        allowOrigins:
          - https://example.com
          - https://app.example.com
        exposeHeaders:
          - X-Custom-Header
        maxAge: 3600
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "functionUrl": {
      "type": "aws/lambda/functionUrl",
      "metadata": {
        "displayName": "Function URL"
      },
      "spec": {
        "targetFunctionArn": "my-function",
        "authType": "AWS_IAM",
        "qualifier": "PROD",
        "invokeMode": "RESPONSE_STREAM",
        "cors": {
          "allowCredentials": true,
          "allowHeaders": [
            "Content-Type",
            "Authorization"
          ],
          "allowMethods": [
            "GET",
            "POST",
            "PUT",
            "DELETE"
          ],
          "allowOrigins": [
            "https://example.com",
            "https://app.example.com"
          ],
          "exposeHeaders": [
            "X-Custom-Header"
          ],
          "maxAge": 3600
        }
      }
    }
  }
}
```
