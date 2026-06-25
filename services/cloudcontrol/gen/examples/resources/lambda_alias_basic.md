A basic AWS Lambda Alias with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource alias: aws/lambda/alias {
    metadata {
        displayName = "AWS Lambda Alias basic"
    }
    spec {
        functionName = "example-function-name"
        functionVersion = "example-function-version"
        name = "example-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    alias:
        type: aws/lambda/alias
        metadata:
            displayName: AWS Lambda Alias basic
        spec:
            functionName: example-function-name
            functionVersion: example-function-version
            name: example-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "alias": {
      "type": "aws/lambda/alias",
      "metadata": {
        "displayName": "AWS Lambda Alias basic"
      },
      "spec": {
        "functionName": "example-function-name",
        "functionVersion": "example-function-version",
        "name": "example-name"
      }
    }
  }
}
```
