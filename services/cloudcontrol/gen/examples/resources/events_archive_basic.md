A basic AWS Events Archive with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource archive: aws/events/archive {
    metadata {
        displayName = "AWS Events Archive basic"
    }
    spec {
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
            displayName: AWS Events Archive basic
        spec:
            sourceArn: example-source-arn
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "archive": {
      "type": "aws/events/archive",
      "metadata": {
        "displayName": "AWS Events Archive basic"
      },
      "spec": {
        "sourceArn": "example-source-arn"
      }
    }
  }
}
```
