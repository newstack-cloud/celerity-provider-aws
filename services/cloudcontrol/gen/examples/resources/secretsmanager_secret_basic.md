A basic AWS SecretsManager Secret with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource secret: aws/secretsmanager/secret {
    metadata {
        displayName = "AWS SecretsManager Secret basic"
    }
    spec {
        description = "example-description"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    secret:
        type: aws/secretsmanager/secret
        metadata:
            displayName: AWS SecretsManager Secret basic
        spec:
            description: example-description
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "secret": {
      "type": "aws/secretsmanager/secret",
      "metadata": {
        "displayName": "AWS SecretsManager Secret basic"
      },
      "spec": {
        "description": "example-description"
      }
    }
  }
}
```
