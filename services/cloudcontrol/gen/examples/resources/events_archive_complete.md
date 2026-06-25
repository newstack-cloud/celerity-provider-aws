A AWS Events Archive configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource archive: aws/events/archive {
    metadata {
        displayName = "AWS Events Archive complete"
    }
    spec {
        archiveName = "example-archive-name"
        description = "example-description"
        eventPattern = {
            source = [
                "com.example.orders"
            ]
        }
        kmsKeyIdentifier = "example-kms-key-identifier"
        retentionDays = 1
        sourceArn = "example-source-arn"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    archive:
        type: aws/events/archive
        metadata:
            displayName: AWS Events Archive complete
        spec:
            archiveName: example-archive-name
            description: example-description
            eventPattern:
                source:
                    - com.example.orders
            kmsKeyIdentifier: example-kms-key-identifier
            retentionDays: 1
            sourceArn: example-source-arn
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "archive": {
      "type": "aws/events/archive",
      "metadata": {
        "displayName": "AWS Events Archive complete"
      },
      "spec": {
        "archiveName": "example-archive-name",
        "description": "example-description",
        "eventPattern": {
          "source": [
            "com.example.orders"
          ]
        },
        "kmsKeyIdentifier": "example-kms-key-identifier",
        "retentionDays": 1,
        "sourceArn": "example-source-arn"
      }
    }
  }
}
```
