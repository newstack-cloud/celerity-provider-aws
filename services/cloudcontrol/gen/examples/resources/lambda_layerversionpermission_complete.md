A AWS Lambda LayerVersionPermission configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource layerVersionPermission: aws/lambda/layerVersionPermission {
    metadata {
        displayName = "AWS Lambda LayerVersionPermission complete"
    }
    spec {
        action = "example-action"
        layerVersionArn = "example-layer-version-arn"
        organizationId = "example-organization-id"
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
            displayName: AWS Lambda LayerVersionPermission complete
        spec:
            action: example-action
            layerVersionArn: example-layer-version-arn
            organizationId: example-organization-id
            principal: example-principal
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "layerVersionPermission": {
      "type": "aws/lambda/layerVersionPermission",
      "metadata": {
        "displayName": "AWS Lambda LayerVersionPermission complete"
      },
      "spec": {
        "action": "example-action",
        "layerVersionArn": "example-layer-version-arn",
        "organizationId": "example-organization-id",
        "principal": "example-principal"
      }
    }
  }
}
```
