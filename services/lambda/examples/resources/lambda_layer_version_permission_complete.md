Create a complete Lambda setup with a layer version and an organization-wide permission referencing it, using all available options.

```blueprintlang
version "2025-11-02"

resource myLayer: aws/lambda/layerVersion {
    metadata {
        displayName = "My Shared Utilities Layer"
    }
    spec {
        layerName = "my-shared-utilities"
        description = "Shared utilities layer for organization"
        content = {
            s3Bucket = "my-lambda-layers-bucket"
            s3Key = "utilities-layer.zip"
        }
        compatibleRuntimes = [ "python3.9", "python3.10", "nodejs18.x" ]
        compatibleArchitectures = [ "x86_64", "arm64" ]
    }
}

resource organizationPermission: aws/lambda/layerVersionPermission {
    metadata {
        displayName = "Organization Permission"
    }
    spec {
        layerVersionArn = resources.myLayer.spec.layerVersionArn
        statementId = "organization-access"
        action = "lambda:GetLayerVersion"
        principal = "*"
        organizationId = "o-abc123defg"
    }
}
```

```yaml
version: 2025-11-02

resources:
  myLayer:
    type: aws/lambda/layerVersion
    metadata:
      displayName: My Shared Utilities Layer
    spec:
      layerName: my-shared-utilities
      description: "Shared utilities layer for organization"
      content:
        s3Bucket: my-lambda-layers-bucket
        s3Key: utilities-layer.zip
      compatibleRuntimes:
        - python3.9
        - python3.10
        - nodejs18.x
      compatibleArchitectures:
        - x86_64
        - arm64
  organizationPermission:
    type: aws/lambda/layerVersionPermission
    metadata:
      displayName: Organization Permission
    spec:
      layerVersionArn: ${resources.myLayer.spec.layerVersionArn}
      statementId: organization-access
      action: lambda:GetLayerVersion
      principal: "*"
      organizationId: o-abc123defg
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "myLayer": {
      "type": "aws/lambda/layerVersion",
      "metadata": {
        "displayName": "My Shared Utilities Layer"
      },
      "spec": {
        "layerName": "my-shared-utilities",
        "description": "Shared utilities layer for organization",
        "content": {
          "s3Bucket": "my-lambda-layers-bucket",
          "s3Key": "utilities-layer.zip"
        },
        "compatibleRuntimes": ["python3.9", "python3.10", "nodejs18.x"],
        "compatibleArchitectures": ["x86_64", "arm64"]
      }
    },
    "organizationPermission": {
      "type": "aws/lambda/layerVersionPermission",
      "metadata": {
        "displayName": "Organization Permission"
      },
      "spec": {
        "layerVersionArn": "${resources.myLayer.spec.layerVersionArn}",
        "statementId": "organization-access",
        "action": "lambda:GetLayerVersion",
        "principal": "*",
        "organizationId": "o-abc123defg"
      }
    }
  }
}
```
