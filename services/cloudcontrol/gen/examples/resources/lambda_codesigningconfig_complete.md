A AWS Lambda CodeSigningConfig configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource codeSigningConfig: aws/lambda/codeSigningConfig {
    metadata {
        displayName = "AWS Lambda CodeSigningConfig complete"
    }
    spec {
        allowedPublishers = {
            signingProfileVersionArns = [
                "example-signing-profile-version-arn"
            ]
        }
        codeSigningPolicies = {
            untrustedArtifactOnDeployment = "Warn"
        }
        description = "example-description"
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    codeSigningConfig:
        type: aws/lambda/codeSigningConfig
        metadata:
            displayName: AWS Lambda CodeSigningConfig complete
        spec:
            allowedPublishers:
                signingProfileVersionArns:
                    - example-signing-profile-version-arn
            codeSigningPolicies:
                untrustedArtifactOnDeployment: Warn
            description: example-description
            tags:
                - key: example-key
                  value: example-value
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "codeSigningConfig": {
      "type": "aws/lambda/codeSigningConfig",
      "metadata": {
        "displayName": "AWS Lambda CodeSigningConfig complete"
      },
      "spec": {
        "allowedPublishers": {
          "signingProfileVersionArns": [
            "example-signing-profile-version-arn"
          ]
        },
        "codeSigningPolicies": {
          "untrustedArtifactOnDeployment": "Warn"
        },
        "description": "example-description",
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ]
      }
    }
  }
}
```
