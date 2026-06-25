A AWS Events EventBusPolicy configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource eventBusPolicy: aws/events/eventBusPolicy {
    metadata {
        displayName = "AWS Events EventBusPolicy complete"
    }
    spec {
        action = "example-action"
        condition = {
            key = "example-key",
            type = "example-type",
            value = "example-value"
        }
        eventBusName = "example-event-bus-name"
        principal = "example-principal"
        statement = {
            action = [
                "s3:GetObject"
            ],
            effect = "Allow",
            resource = "arn:aws:s3:::example-bucket/*"
        }
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
            displayName: AWS Events EventBusPolicy complete
        spec:
            action: example-action
            condition:
                key: example-key
                type: example-type
                value: example-value
            eventBusName: example-event-bus-name
            principal: example-principal
            statement:
                action:
                    - s3:GetObject
                effect: Allow
                resource: arn:aws:s3:::example-bucket/*
            statementId: example-statement-id
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "eventBusPolicy": {
      "type": "aws/events/eventBusPolicy",
      "metadata": {
        "displayName": "AWS Events EventBusPolicy complete"
      },
      "spec": {
        "action": "example-action",
        "condition": {
          "key": "example-key",
          "type": "example-type",
          "value": "example-value"
        },
        "eventBusName": "example-event-bus-name",
        "principal": "example-principal",
        "statement": {
          "action": [
            "s3:GetObject"
          ],
          "effect": "Allow",
          "resource": "arn:aws:s3:::example-bucket/*"
        },
        "statementId": "example-statement-id"
      }
    }
  }
}
```
