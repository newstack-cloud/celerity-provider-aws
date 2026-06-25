A basic AWS Events Rule with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource rule: aws/events/rule {
    metadata {
        displayName = "AWS Events Rule basic"
    }
    spec {
        eventBusName = "example-event-bus-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    rule:
        type: aws/events/rule
        metadata:
            displayName: AWS Events Rule basic
        spec:
            eventBusName: example-event-bus-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "rule": {
      "type": "aws/events/rule",
      "metadata": {
        "displayName": "AWS Events Rule basic"
      },
      "spec": {
        "eventBusName": "example-event-bus-name"
      }
    }
  }
}
```
