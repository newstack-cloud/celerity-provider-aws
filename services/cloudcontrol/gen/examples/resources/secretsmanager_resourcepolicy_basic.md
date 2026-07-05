A basic AWS SecretsManager ResourcePolicy with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource resourcePolicy: aws/secretsmanager/resourcePolicy {
    metadata {
        displayName = "AWS SecretsManager ResourcePolicy basic"
    }
    spec {
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
            displayName: AWS SecretsManager ResourcePolicy basic
        spec:
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
        "displayName": "AWS SecretsManager ResourcePolicy basic"
      },
      "spec": {
        "resourcePolicy": {
          "exampleKey": "example-value"
        },
        "secretId": "example-secret-id"
      }
    }
  }
}
```
