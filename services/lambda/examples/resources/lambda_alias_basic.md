Create a basic Lambda alias that points to a specific function version.

```blueprintlang
version "2025-11-02"

resource productionAlias: aws/lambda/alias {
    metadata {
        displayName = "Production Alias"
    }
    spec {
        functionName = "my-lambda-function"
        name = "PROD"
        functionVersion = "1"
        description = "Production alias for my Lambda function"
    }
}
```

```yaml
version: 2025-11-02

resources:
  productionAlias:
    type: aws/lambda/alias
    metadata:
      displayName: Production Alias
    spec:
      functionName: my-lambda-function
      name: PROD
      functionVersion: "1"
      description: Production alias for my Lambda function
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "productionAlias": {
      "type": "aws/lambda/alias",
      "metadata": {
        "displayName": "Production Alias"
      },
      "spec": {
        "functionName": "my-lambda-function",
        "name": "PROD",
        "functionVersion": "1",
        "description": "Production alias for my Lambda function"
      }
    }
  }
}
```
