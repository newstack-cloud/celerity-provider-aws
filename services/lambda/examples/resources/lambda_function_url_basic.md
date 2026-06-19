Create a basic Lambda function URL with no authentication required.

```blueprintlang
version "2025-11-02"

resource functionUrl: aws/lambda/functionUrl {
    metadata {
        displayName = "Function URL"
    }
    spec {
        targetFunctionArn = "my-function"
        authType = "NONE"
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
      authType: NONE
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
        "authType": "NONE"
      }
    }
  }
}
```
