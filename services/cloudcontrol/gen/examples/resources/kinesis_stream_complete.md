A AWS Kinesis Stream configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource stream: aws/kinesis/stream {
    metadata {
        displayName = "AWS Kinesis Stream complete"
    }
    spec {
        desiredShardLevelMetrics = [
            "IncomingBytes"
        ]
        maxRecordSizeInKiB = 1024
        name = "example-name"
        retentionPeriodHours = 24
        shardCount = 1
        streamEncryption = {
            encryptionType = "KMS",
            keyId = "example-key-id"
        }
        streamModeDetails = {
            streamMode = "ON_DEMAND"
        }
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        warmThroughputMiBps = 1
    }
}
```

```yaml
version: "2025-11-02"
resources:
    stream:
        type: aws/kinesis/stream
        metadata:
            displayName: AWS Kinesis Stream complete
        spec:
            desiredShardLevelMetrics:
                - IncomingBytes
            maxRecordSizeInKiB: 1024
            name: example-name
            retentionPeriodHours: 24
            shardCount: 1
            streamEncryption:
                encryptionType: KMS
                keyId: example-key-id
            streamModeDetails:
                streamMode: ON_DEMAND
            tags:
                - key: example-key
                  value: example-value
            warmThroughputMiBps: 1
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "stream": {
      "type": "aws/kinesis/stream",
      "metadata": {
        "displayName": "AWS Kinesis Stream complete"
      },
      "spec": {
        "desiredShardLevelMetrics": [
          "IncomingBytes"
        ],
        "maxRecordSizeInKiB": 1024,
        "name": "example-name",
        "retentionPeriodHours": 24,
        "shardCount": 1,
        "streamEncryption": {
          "encryptionType": "KMS",
          "keyId": "example-key-id"
        },
        "streamModeDetails": {
          "streamMode": "ON_DEMAND"
        },
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "warmThroughputMiBps": 1
      }
    }
  }
}
```
