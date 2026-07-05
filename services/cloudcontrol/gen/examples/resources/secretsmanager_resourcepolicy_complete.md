A AWS SecretsManager ResourcePolicy configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource resourcePolicy: aws/secretsmanager/resourcePolicy {
    metadata {
        displayName = "AWS SecretsManager ResourcePolicy complete"
    }
    spec {
        blockPublicPolicy = false
        resourcePolicy = {
            exampleKey = "example-value"
        }
        secretId = "example-secret-id"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    resourcePolicy:
        type: aws/secretsmanager/resourcePolicy
        metadata:
            displayName: AWS SecretsManager ResourcePolicy complete
        spec:
            blockPublicPolicy: false
            resourcePolicy:
                exampleKey: example-value
            secretId: example-secret-id
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "resourcePolicy": {
      "type": "aws/secretsmanager/resourcePolicy",
      "metadata": {
        "displayName": "AWS SecretsManager ResourcePolicy complete"
      },
      "spec": {
        "blockPublicPolicy": false,
        "resourcePolicy": {
          "exampleKey": "example-value"
        },
        "secretId": "example-secret-id"
      }
    }
  }
}
```
