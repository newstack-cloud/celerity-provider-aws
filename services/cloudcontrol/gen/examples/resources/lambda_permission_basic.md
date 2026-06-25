A basic AWS Lambda Permission with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource permission: aws/lambda/permission {
    metadata {
        displayName = "AWS Lambda Permission basic"
    }
    spec {
        action = "example-action"
        functionName = "example-function-name"
        principal = "example-principal"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    permission:
        type: aws/lambda/permission
        metadata:
            displayName: AWS Lambda Permission basic
        spec:
            action: example-action
            functionName: example-function-name
            principal: example-principal
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "permission": {
      "type": "aws/lambda/permission",
      "metadata": {
        "displayName": "AWS Lambda Permission basic"
      },
      "spec": {
        "action": "example-action",
        "functionName": "example-function-name",
        "principal": "example-principal"
      }
    }
  }
}
```
