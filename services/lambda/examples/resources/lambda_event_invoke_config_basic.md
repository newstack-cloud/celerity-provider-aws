Create a basic Lambda event invoke config that sets retry attempts and the maximum event age for asynchronous invocations.

```blueprintlang
version "2025-11-02"

resource myEventInvokeConfig: aws/lambda/eventInvokeConfig {
    metadata {
        displayName = "My Event Invoke Config"
    }
    spec {
        functionName = "my-lambda-function"
        qualifier = "$LATEST"
        maximumRetryAttempts = 1
        maximumEventAgeInSeconds = 300
    }
}
```

```yaml
version: 2025-11-02

resources:
  myEventInvokeConfig:
    type: aws/lambda/eventInvokeConfig
    metadata:
      displayName: My Event Invoke Config
    spec:
      functionName: my-lambda-function
      qualifier: $LATEST
      maximumRetryAttempts: 1
      maximumEventAgeInSeconds: 300
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "myEventInvokeConfig": {
      "type": "aws/lambda/eventInvokeConfig",
      "metadata": {
        "displayName": "My Event Invoke Config"
      },
      "spec": {
        "functionName": "my-lambda-function",
        "qualifier": "$LATEST",
        "maximumRetryAttempts": 1,
        "maximumEventAgeInSeconds": 300
      }
    }
  }
}
```
