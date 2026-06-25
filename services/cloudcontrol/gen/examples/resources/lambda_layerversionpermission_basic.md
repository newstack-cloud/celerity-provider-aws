A basic AWS Lambda LayerVersionPermission with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource layerVersionPermission: aws/lambda/layerVersionPermission {
    metadata {
        displayName = "AWS Lambda LayerVersionPermission basic"
    }
    spec {
        action = "example-action"
        layerVersionArn = "example-layer-version-arn"
        principal = "example-principal"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    layerVersionPermission:
        type: aws/lambda/layerVersionPermission
        metadata:
            displayName: AWS Lambda LayerVersionPermission basic
        spec:
            action: example-action
            layerVersionArn: example-layer-version-arn
            principal: example-principal
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "layerVersionPermission": {
      "type": "aws/lambda/layerVersionPermission",
      "metadata": {
        "displayName": "AWS Lambda LayerVersionPermission basic"
      },
      "spec": {
        "action": "example-action",
        "layerVersionArn": "example-layer-version-arn",
        "principal": "example-principal"
      }
    }
  }
}
```
