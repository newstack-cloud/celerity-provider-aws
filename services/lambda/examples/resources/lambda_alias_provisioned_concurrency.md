Create a Lambda alias with provisioned concurrency for high-performance workloads.

```blueprintlang
version "2025-11-02"

resource highPerfAlias: aws/lambda/alias {
    metadata {
        displayName = "High Performance Alias"
    }
    spec {
        functionName = "my-lambda-function"
        name = "HIGHPERF"
        functionVersion = "3"
        description = "High performance alias with provisioned concurrency"
        provisionedConcurrencyConfig = {
            provisionedConcurrentExecutions = 10
        }
    }
}
```

```yaml
version: 2025-11-02

resources:
  highPerfAlias:
    type: aws/lambda/alias
    metadata:
      displayName: High Performance Alias
    spec:
      functionName: my-lambda-function
      name: HIGHPERF
      functionVersion: "3"
      description: High performance alias with provisioned concurrency
      provisionedConcurrencyConfig:
        provisionedConcurrentExecutions: 10
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "highPerfAlias": {
      "type": "aws/lambda/alias",
      "metadata": {
        "displayName": "High Performance Alias"
      },
      "spec": {
        "functionName": "my-lambda-function",
        "name": "HIGHPERF",
        "functionVersion": "3",
        "description": "High performance alias with provisioned concurrency",
        "provisionedConcurrencyConfig": {
          "provisionedConcurrentExecutions": 10
        }
      }
    }
  }
}
```
