A basic AWS Events EventBusPolicy with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource eventBusPolicy: aws/events/eventBusPolicy {
    metadata {
        displayName = "AWS Events EventBusPolicy basic"
    }
    spec {
        statementId = "example-statement-id"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    eventBusPolicy:
        type: aws/events/eventBusPolicy
        metadata:
            displayName: AWS Events EventBusPolicy basic
        spec:
            statementId: example-statement-id
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "eventBusPolicy": {
      "type": "aws/events/eventBusPolicy",
      "metadata": {
        "displayName": "AWS Events EventBusPolicy basic"
      },
      "spec": {
        "statementId": "example-statement-id"
      }
    }
  }
}
```
