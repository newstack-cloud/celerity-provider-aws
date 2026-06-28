A basic AWS Logs LogGroup with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource logGroup: aws/logs/logGroup {
    metadata {
        displayName = "AWS Logs LogGroup basic"
    }
    spec {
        logGroupName = "example-log-group-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    logGroup:
        type: aws/logs/logGroup
        metadata:
            displayName: AWS Logs LogGroup basic
        spec:
            logGroupName: example-log-group-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "logGroup": {
      "type": "aws/logs/logGroup",
      "metadata": {
        "displayName": "AWS Logs LogGroup basic"
      },
      "spec": {
        "logGroupName": "example-log-group-name"
      }
    }
  }
}
```
