Create a Lambda code signing configuration with a single allowed signing profile.

```blueprintlang
version "2025-11-02"

resource codeSigningConfig: aws/lambda/codeSigningConfig {
    metadata {
        displayName = "Code Signing Config"
    }
    spec {
        allowedPublishers = {
            signingProfileVersionArns = [
                "arn:aws:signer:us-east-1:123456789012:/signing-profiles/ExampleProfile/abcdef12"
            ]
        }
    }
}
```

```yaml
version: 2025-11-02

resources:
  codeSigningConfig:
    type: aws/lambda/codeSigningConfig
    metadata:
      displayName: Code Signing Config
    spec:
      allowedPublishers:
        signingProfileVersionArns:
          - arn:aws:signer:us-east-1:123456789012:/signing-profiles/ExampleProfile/abcdef12
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "codeSigningConfig": {
      "type": "aws/lambda/codeSigningConfig",
      "metadata": {
        "displayName": "Code Signing Config"
      },
      "spec": {
        "allowedPublishers": {
          "signingProfileVersionArns": [
            "arn:aws:signer:us-east-1:123456789012:/signing-profiles/ExampleProfile/abcdef12"
          ]
        }
      }
    }
  }
}
```
