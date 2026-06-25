A basic AWS Lambda LayerVersion with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource layerVersion: aws/lambda/layerVersion {
    metadata {
        displayName = "AWS Lambda LayerVersion basic"
    }
    spec {
        content = {
            s3Bucket = "example-s3-bucket",
            s3Key = "example-s3-key",
            s3ObjectStorageMode = "COPY",
            s3ObjectVersion = "example-s3-object-version"
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    layerVersion:
        type: aws/lambda/layerVersion
        metadata:
            displayName: AWS Lambda LayerVersion basic
        spec:
            content:
                s3Bucket: example-s3-bucket
                s3Key: example-s3-key
                s3ObjectStorageMode: COPY
                s3ObjectVersion: example-s3-object-version
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "layerVersion": {
      "type": "aws/lambda/layerVersion",
      "metadata": {
        "displayName": "AWS Lambda LayerVersion basic"
      },
      "spec": {
        "content": {
          "s3Bucket": "example-s3-bucket",
          "s3Key": "example-s3-key",
          "s3ObjectStorageMode": "COPY",
          "s3ObjectVersion": "example-s3-object-version"
        }
      }
    }
  }
}
```
