A KMS-encrypted SecureString parameter with tags, using the default AWS-managed key.

```blueprintlang
version "2025-11-02"

resource apiKey: aws/ssm/parameter {
    metadata {
        displayName = "Third-party API key"
    }
    spec {
        name = "/my-app/prod/api-key"
        type = "SecureString"
        secureValue = "${variables.apiKey}"
        tier = "Standard"
        tags {
            Environment = "production"
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    apiKey:
        type: aws/ssm/parameter
        metadata:
            displayName: Third-party API key
        spec:
            name: /my-app/prod/api-key
            type: SecureString
            secureValue: "${variables.apiKey}"
            tier: Standard
            tags:
                Environment: production
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "apiKey": {
      "type": "aws/ssm/parameter",
      "metadata": {
        "displayName": "Third-party API key"
      },
      "spec": {
        "name": "/my-app/prod/api-key",
        "type": "SecureString",
        "secureValue": "${variables.apiKey}",
        "tier": "Standard",
        "tags": {
          "Environment": "production"
        }
      }
    }
  }
}
```
