A basic AWS Lambda Url with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource url: aws/lambda/functionUrl {
    metadata {
        displayName = "AWS Lambda Url basic"
    }
    spec {
        authType = "AWS_IAM"
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
            displayName: AWS Lambda Url basic
        spec:
            authType: AWS_IAM
            targetFunctionArn: example-target-function-arn
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "url": {
      "type": "aws/lambda/functionUrl",
      "metadata": {
        "displayName": "AWS Lambda Url basic"
      },
      "spec": {
        "authType": "AWS_IAM",
        "targetFunctionArn": "example-target-function-arn"
      }
    }
  }
}
```
