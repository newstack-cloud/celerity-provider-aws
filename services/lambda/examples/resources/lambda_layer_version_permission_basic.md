Grant a specific AWS account permission to use a Lambda layer version.

```blueprintlang
version "2025-11-02"

resource layerPermission: aws/lambda/layerVersionPermission {
    metadata {
        displayName = "Layer Permission"
    }
    spec {
        layerVersionArn = "arn:aws:lambda:us-east-1:123456789012:layer:my-layer:1"
        statementId = "my-permission"
        action = "lambda:GetLayerVersion"
        principal = "987654321098"
    }
}
```

```yaml
version: 2025-11-02

resources:
  layerPermission:
    type: aws/lambda/layerVersionPermission
    metadata:
      displayName: Layer Permission
    spec:
      layerVersionArn: arn:aws:lambda:us-east-1:123456789012:layer:my-layer:1
      statementId: my-permission
      action: lambda:GetLayerVersion
      principal: "987654321098"
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "layerPermission": {
      "type": "aws/lambda/layerVersionPermission",
      "metadata": {
        "displayName": "Layer Permission"
      },
      "spec": {
        "layerVersionArn": "arn:aws:lambda:us-east-1:123456789012:layer:my-layer:1",
        "statementId": "my-permission",
        "action": "lambda:GetLayerVersion",
        "principal": "987654321098"
      }
    }
  }
}
```
