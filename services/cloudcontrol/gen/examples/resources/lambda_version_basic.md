A basic AWS Lambda Version with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource "version": aws/lambda/functionVersion {
    metadata {
        displayName = "AWS Lambda Version basic"
    }
    spec {
        functionName = "example-function-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    version:
        type: aws/lambda/functionVersion
        metadata:
            displayName: AWS Lambda Version basic
        spec:
            functionName: example-function-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "version": {
      "type": "aws/lambda/functionVersion",
      "metadata": {
        "displayName": "AWS Lambda Version basic"
      },
      "spec": {
        "functionName": "example-function-name"
      }
    }
  }
}
```
