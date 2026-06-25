A basic AWS Events EventBus with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource eventBus: aws/events/eventBus {
    metadata {
        displayName = "AWS Events EventBus basic"
    }
    spec {
        name = "example-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    eventBus:
        type: aws/events/eventBus
        metadata:
            displayName: AWS Events EventBus basic
        spec:
            name: example-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "eventBus": {
      "type": "aws/events/eventBus",
      "metadata": {
        "displayName": "AWS Events EventBus basic"
      },
      "spec": {
        "name": "example-name"
      }
    }
  }
}
```
