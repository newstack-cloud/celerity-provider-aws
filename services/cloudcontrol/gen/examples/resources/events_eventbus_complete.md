A AWS Events EventBus configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource eventBus: aws/events/eventBus {
    metadata {
        displayName = "AWS Events EventBus complete"
    }
    spec {
        deadLetterConfig = {
            arn = "example-arn"
        }
        description = "example-description"
        eventSourceName = "example-event-source-name"
        kmsKeyIdentifier = "example-kms-key-identifier"
        logConfig = {
            includeDetail = "FULL",
            level = "INFO"
        }
        name = "example-name"
        policy = {
            statement = [
                {
                    action = [
                        "s3:GetObject"
                    ],
                    effect = "Allow",
                    resource = "arn:aws:s3:::example-bucket/*"
                }
            ],
            version = "2012-10-17"
        }
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    eventBus:
        type: aws/events/eventBus
        metadata:
            displayName: AWS Events EventBus complete
        spec:
            deadLetterConfig:
                arn: example-arn
            description: example-description
            eventSourceName: example-event-source-name
            kmsKeyIdentifier: example-kms-key-identifier
            logConfig:
                includeDetail: FULL
                level: INFO
            name: example-name
            policy:
                statement:
                    - action:
                        - s3:GetObject
                      effect: Allow
                      resource: arn:aws:s3:::example-bucket/*
                version: "2012-10-17"
            tags:
                - key: example-key
                  value: example-value
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "eventBus": {
      "type": "aws/events/eventBus",
      "metadata": {
        "displayName": "AWS Events EventBus complete"
      },
      "spec": {
        "deadLetterConfig": {
          "arn": "example-arn"
        },
        "description": "example-description",
        "eventSourceName": "example-event-source-name",
        "kmsKeyIdentifier": "example-kms-key-identifier",
        "logConfig": {
          "includeDetail": "FULL",
          "level": "INFO"
        },
        "name": "example-name",
        "policy": {
          "statement": [
            {
              "action": [
                "s3:GetObject"
              ],
              "effect": "Allow",
              "resource": "arn:aws:s3:::example-bucket/*"
            }
          ],
          "version": "2012-10-17"
        },
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ]
      }
    }
  }
}
```
