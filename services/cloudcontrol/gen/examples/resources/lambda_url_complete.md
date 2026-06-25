A AWS Lambda Url configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource url: aws/lambda/functionUrl {
    metadata {
        displayName = "AWS Lambda Url complete"
    }
    spec {
        authType = "AWS_IAM"
        cors = {
            allowCredentials = false,
            allowHeaders = [
                "example-allow-header"
            ],
            allowMethods = [
                "GET"
            ],
            allowOrigins = [
                "example-allow-origin"
            ],
            exposeHeaders = [
                "example-expose-header"
            ],
            maxAge = 0
        }
        invokeMode = "BUFFERED"
        qualifier = "example-qualifier"
        targetFunctionArn = "example-target-function-arn"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    url:
        type: aws/lambda/functionUrl
        metadata:
            displayName: AWS Lambda Url complete
        spec:
            authType: AWS_IAM
            cors:
                allowCredentials: false
                allowHeaders:
                    - example-allow-header
                allowMethods:
                    - GET
                allowOrigins:
                    - example-allow-origin
                exposeHeaders:
                    - example-expose-header
                maxAge: 0
            invokeMode: BUFFERED
            qualifier: example-qualifier
            targetFunctionArn: example-target-function-arn
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "url": {
      "type": "aws/lambda/functionUrl",
      "metadata": {
        "displayName": "AWS Lambda Url complete"
      },
      "spec": {
        "authType": "AWS_IAM",
        "cors": {
          "allowCredentials": false,
          "allowHeaders": [
            "example-allow-header"
          ],
          "allowMethods": [
            "GET"
          ],
          "allowOrigins": [
            "example-allow-origin"
          ],
          "exposeHeaders": [
            "example-expose-header"
          ],
          "maxAge": 0
        },
        "invokeMode": "BUFFERED",
        "qualifier": "example-qualifier",
        "targetFunctionArn": "example-target-function-arn"
      }
    }
  }
}
```
