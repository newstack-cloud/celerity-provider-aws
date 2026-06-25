A AWS Lambda LayerVersion configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource layerVersion: aws/lambda/layerVersion {
    metadata {
        displayName = "AWS Lambda LayerVersion complete"
    }
    spec {
        compatibleArchitectures = [
            "example-compatible-architecture"
        ]
        compatibleRuntimes = [
            "example-compatible-runtime"
        ]
        content = {
            s3Bucket = "example-s3-bucket",
            s3Key = "example-s3-key",
            s3ObjectStorageMode = "COPY",
            s3ObjectVersion = "example-s3-object-version"
        }
        description = "example-description"
        layerName = "example-layer-name"
        licenseInfo = "example-license-info"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    layerVersion:
        type: aws/lambda/layerVersion
        metadata:
            displayName: AWS Lambda LayerVersion complete
        spec:
            compatibleArchitectures:
                - example-compatible-architecture
            compatibleRuntimes:
                - example-compatible-runtime
            content:
                s3Bucket: example-s3-bucket
                s3Key: example-s3-key
                s3ObjectStorageMode: COPY
                s3ObjectVersion: example-s3-object-version
            description: example-description
            layerName: example-layer-name
            licenseInfo: example-license-info
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "layerVersion": {
      "type": "aws/lambda/layerVersion",
      "metadata": {
        "displayName": "AWS Lambda LayerVersion complete"
      },
      "spec": {
        "compatibleArchitectures": [
          "example-compatible-architecture"
        ],
        "compatibleRuntimes": [
          "example-compatible-runtime"
        ],
        "content": {
          "s3Bucket": "example-s3-bucket",
          "s3Key": "example-s3-key",
          "s3ObjectStorageMode": "COPY",
          "s3ObjectVersion": "example-s3-object-version"
        },
        "description": "example-description",
        "layerName": "example-layer-name",
        "licenseInfo": "example-license-info"
      }
    }
  }
}
```
