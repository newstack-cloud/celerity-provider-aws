Create Lambda layer versions with all optional fields, including object versions, multiple compatible runtimes and architectures, and license information.

```blueprintlang
version "2025-11-02"

resource fullS3LayerVersion: aws/lambda/layerVersion {
    metadata {
        displayName = "Full S3 Layer Version"
    }
    spec {
        layerName = "comprehensive-python-layer"
        description = "Comprehensive Python layer with data science libraries"
        content = {
            s3Bucket = "my-company-lambda-layers"
            s3Key = "data-science-layer-v2.3.1.zip"
            s3ObjectVersion = "version-12345abcdef"
        }
        compatibleRuntimes = [ "python3.9", "python3.10", "python3.11", "python3.12" ]
        compatibleArchitectures = [ "x86_64", "arm64" ]
        licenseInfo = "Apache-2.0"
    }
}

resource universalLayerVersion: aws/lambda/layerVersion {
    metadata {
        displayName = "Universal Layer Version"
    }
    spec {
        layerName = "universal-utilities"
        description = "Universal utilities compatible with multiple runtimes"
        content = {
            s3Bucket = "shared-lambda-layers"
            s3Key = "universal-utils/latest.zip"
        }
        compatibleRuntimes = [ "python3.9", "python3.10", "python3.11", "nodejs18.x", "nodejs20.x", "java11", "java17", "java21", "dotnet6", "dotnet8", "go1.x", "provided.al2", "provided.al2023" ]
        compatibleArchitectures = [ "x86_64", "arm64" ]
        licenseInfo = "BSD-3-Clause"
    }
}
```

```yaml
version: 2025-11-02

resources:
  fullS3LayerVersion:
    type: aws/lambda/layerVersion
    metadata:
      displayName: Full S3 Layer Version
    spec:
      layerName: comprehensive-python-layer
      description: "Comprehensive Python layer with data science libraries"
      content:
        s3Bucket: my-company-lambda-layers
        s3Key: data-science-layer-v2.3.1.zip
        s3ObjectVersion: "version-12345abcdef"
      compatibleRuntimes:
        - python3.9
        - python3.10
        - python3.11
        - python3.12
      compatibleArchitectures:
        - x86_64
        - arm64
      licenseInfo: "Apache-2.0"
  universalLayerVersion:
    type: aws/lambda/layerVersion
    metadata:
      displayName: Universal Layer Version
    spec:
      layerName: universal-utilities
      description: "Universal utilities compatible with multiple runtimes"
      content:
        s3Bucket: shared-lambda-layers
        s3Key: universal-utils/latest.zip
      compatibleRuntimes:
        - python3.9
        - python3.10
        - python3.11
        - nodejs18.x
        - nodejs20.x
        - java11
        - java17
        - java21
        - dotnet6
        - dotnet8
        - go1.x
        - provided.al2
        - provided.al2023
      compatibleArchitectures:
        - x86_64
        - arm64
      licenseInfo: "BSD-3-Clause"
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "fullS3LayerVersion": {
      "type": "aws/lambda/layerVersion",
      "metadata": {
        "displayName": "Full S3 Layer Version"
      },
      "spec": {
        "layerName": "comprehensive-python-layer",
        "description": "Comprehensive Python layer with data science libraries",
        "content": {
          "s3Bucket": "my-company-lambda-layers",
          "s3Key": "data-science-layer-v2.3.1.zip",
          "s3ObjectVersion": "version-12345abcdef"
        },
        "compatibleRuntimes": ["python3.9", "python3.10", "python3.11", "python3.12"],
        "compatibleArchitectures": ["x86_64", "arm64"],
        "licenseInfo": "Apache-2.0"
      }
    },
    "universalLayerVersion": {
      "type": "aws/lambda/layerVersion",
      "metadata": {
        "displayName": "Universal Layer Version"
      },
      "spec": {
        "layerName": "universal-utilities",
        "description": "Universal utilities compatible with multiple runtimes",
        "content": {
          "s3Bucket": "shared-lambda-layers",
          "s3Key": "universal-utils/latest.zip"
        },
        "compatibleRuntimes": ["python3.9", "python3.10", "python3.11", "nodejs18.x", "nodejs20.x", "java11", "java17", "java21", "dotnet6", "dotnet8", "go1.x", "provided.al2", "provided.al2023"],
        "compatibleArchitectures": ["x86_64", "arm64"],
        "licenseInfo": "BSD-3-Clause"
      }
    }
  }
}
```
