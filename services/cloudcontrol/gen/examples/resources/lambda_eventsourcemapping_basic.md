A basic AWS Lambda EventSourceMapping with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource eventSourceMapping: aws/lambda/eventSourceMapping {
    metadata {
        displayName = "AWS Lambda EventSourceMapping basic"
    }
    spec {
        functionName = "example-function-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    eventSourceMapping:
        type: aws/lambda/eventSourceMapping
        metadata:
            displayName: AWS Lambda EventSourceMapping basic
        spec:
            functionName: example-function-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "eventSourceMapping": {
      "type": "aws/lambda/eventSourceMapping",
      "metadata": {
        "displayName": "AWS Lambda EventSourceMapping basic"
      },
      "spec": {
        "functionName": "example-function-name"
      }
    }
  }
}
```
