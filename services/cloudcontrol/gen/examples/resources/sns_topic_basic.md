A basic AWS SNS Topic with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource topic: aws/sns/topic {
    metadata {
        displayName = "AWS SNS Topic basic"
    }
    spec {
        displayName = "example-display-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    topic:
        type: aws/sns/topic
        metadata:
            displayName: AWS SNS Topic basic
        spec:
            displayName: example-display-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "topic": {
      "type": "aws/sns/topic",
      "metadata": {
        "displayName": "AWS SNS Topic basic"
      },
      "spec": {
        "displayName": "example-display-name"
      }
    }
  }
}
```
