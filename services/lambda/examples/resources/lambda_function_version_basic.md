Create a basic Lambda function version for an existing function.

```blueprintlang
version "2025-11-02"

resource version1: aws/lambda/functionVersion {
    metadata {
        displayName = "Function Version 1"
    }
    spec {
        functionName = "my-lambda-function"
        description = "Initial version of the function"
    }
}
```

```yaml
version: 2025-11-02

resources:
  version1:
    type: aws/lambda/functionVersion
    metadata:
      displayName: Function Version 1
    spec:
      functionName: my-lambda-function
      description: "Initial version of the function"
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "version1": {
      "type": "aws/lambda/functionVersion",
      "metadata": {
        "displayName": "Function Version 1"
      },
      "spec": {
        "functionName": "my-lambda-function",
        "description": "Initial version of the function"
      }
    }
  }
}
```
