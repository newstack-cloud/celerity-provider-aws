A basic AWS Kinesis Stream with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource stream: aws/kinesis/stream {
    metadata {
        displayName = "AWS Kinesis Stream basic"
    }
    spec {
        maxRecordSizeInKiB = 1024
    }
}
```

```yaml
version: "2025-11-02"
resources:
    stream:
        type: aws/kinesis/stream
        metadata:
            displayName: AWS Kinesis Stream basic
        spec:
            maxRecordSizeInKiB: 1024
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "stream": {
      "type": "aws/kinesis/stream",
      "metadata": {
        "displayName": "AWS Kinesis Stream basic"
      },
      "spec": {
        "maxRecordSizeInKiB": 1024
      }
    }
  }
}
```
