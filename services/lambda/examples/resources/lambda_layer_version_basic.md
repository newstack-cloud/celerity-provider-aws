Create a basic Lambda layer version from an S3 object with Python runtime compatibility.

```blueprintlang
version "2025-11-02"

resource myLayerVersion: aws/lambda/layerVersion {
    metadata {
        displayName = "My Python Layer Version"
    }
    spec {
        layerName = "my-python-layer"
        description = "Basic Python utilities layer"
        content = {
            s3Bucket = "my-lambda-layers-bucket"
            s3Key = "python-utils-layer.zip"
        }
        compatibleRuntimes = [ "python3.9", "python3.10", "python3.11" ]
        compatibleArchitectures = [ "x86_64" ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  myLayerVersion:
    type: aws/lambda/layerVersion
    metadata:
      displayName: My Python Layer Version
    spec:
      layerName: my-python-layer
      description: "Basic Python utilities layer"
      content:
        s3Bucket: my-lambda-layers-bucket
        s3Key: python-utils-layer.zip
      compatibleRuntimes:
        - python3.9
        - python3.10
        - python3.11
      compatibleArchitectures:
        - x86_64
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "myLayerVersion": {
      "type": "aws/lambda/layerVersion",
      "metadata": {
        "displayName": "My Python Layer Version"
      },
      "spec": {
        "layerName": "my-python-layer",
        "description": "Basic Python utilities layer",
        "content": {
          "s3Bucket": "my-lambda-layers-bucket",
          "s3Key": "python-utils-layer.zip"
        },
        "compatibleRuntimes": ["python3.9", "python3.10", "python3.11"],
        "compatibleArchitectures": ["x86_64"]
      }
    }
  }
}
```
