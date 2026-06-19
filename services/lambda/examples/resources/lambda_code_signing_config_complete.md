Create a Lambda code signing configuration with multiple signing profiles, an enforced deployment policy, a description, and tags.

```blueprintlang
version "2025-11-02"

resource codeSigningConfig: aws/lambda/codeSigningConfig {
    metadata {
        displayName = "Code Signing Config"
    }
    spec {
        allowedPublishers = {
            signingProfileVersionArns = [
                "arn:aws:signer:us-east-1:123456789012:/signing-profiles/ExampleProfile/abcdef12",
                "arn:aws:signer:us-east-1:123456789012:/signing-profiles/ExampleProfile2/ghijkl34"
            ]
        }
        codeSigningPolicies = {
            untrustedArtifactOnDeployment = "Enforce"
        }
        description = "Production code signing configuration"
        tags = [
            {
                key = "Environment"
                value = "Production"
            },
            {
                key = "Team"
                value = "Backend"
            },
            {
                key = "Project"
                value = "MyApplication"
            }
        ]
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
          - arn:aws:signer:us-east-1:123456789012:/signing-profiles/ExampleProfile2/ghijkl34
      codeSigningPolicies:
        untrustedArtifactOnDeployment: Enforce
      description: "Production code signing configuration"
      tags:
        - key: Environment
          value: Production
        - key: Team
          value: Backend
        - key: Project
          value: MyApplication
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
            "arn:aws:signer:us-east-1:123456789012:/signing-profiles/ExampleProfile/abcdef12",
            "arn:aws:signer:us-east-1:123456789012:/signing-profiles/ExampleProfile2/ghijkl34"
          ]
        },
        "codeSigningPolicies": {
          "untrustedArtifactOnDeployment": "Enforce"
        },
        "description": "Production code signing configuration",
        "tags": [
          {
            "key": "Environment",
            "value": "Production"
          },
          {
            "key": "Team",
            "value": "Backend"
          },
          {
            "key": "Project",
            "value": "MyApplication"
          }
        ]
      }
    }
  }
}
```
