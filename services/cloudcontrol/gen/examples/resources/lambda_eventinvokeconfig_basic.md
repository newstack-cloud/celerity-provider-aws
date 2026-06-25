A basic AWS Lambda EventInvokeConfig with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource eventInvokeConfig: aws/lambda/eventInvokeConfig {
    metadata {
        displayName = "AWS Lambda EventInvokeConfig basic"
    }
    spec {
        functionName = "example-function-name"
        qualifier = "example-qualifier"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    eventInvokeConfig:
        type: aws/lambda/eventInvokeConfig
        metadata:
            displayName: AWS Lambda EventInvokeConfig basic
        spec:
            functionName: example-function-name
            qualifier: example-qualifier
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "eventInvokeConfig": {
      "type": "aws/lambda/eventInvokeConfig",
      "metadata": {
        "displayName": "AWS Lambda EventInvokeConfig basic"
      },
      "spec": {
        "functionName": "example-function-name",
        "qualifier": "example-qualifier"
      }
    }
  }
}
```
