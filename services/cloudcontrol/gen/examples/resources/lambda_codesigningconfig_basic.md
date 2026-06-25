A basic AWS Lambda CodeSigningConfig with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource codeSigningConfig: aws/lambda/codeSigningConfig {
    metadata {
        displayName = "AWS Lambda CodeSigningConfig basic"
    }
    spec {
        allowedPublishers = {
            signingProfileVersionArns = [
                "example-signing-profile-version-arn"
            ]
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    codeSigningConfig:
        type: aws/lambda/codeSigningConfig
        metadata:
            displayName: AWS Lambda CodeSigningConfig basic
        spec:
            allowedPublishers:
                signingProfileVersionArns:
                    - example-signing-profile-version-arn
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "codeSigningConfig": {
      "type": "aws/lambda/codeSigningConfig",
      "metadata": {
        "displayName": "AWS Lambda CodeSigningConfig basic"
      },
      "spec": {
        "allowedPublishers": {
          "signingProfileVersionArns": [
            "example-signing-profile-version-arn"
          ]
        }
      }
    }
  }
}
```
