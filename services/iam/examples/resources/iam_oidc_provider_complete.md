This example demonstrates how to create a comprehensive IAM OIDC provider with all available configuration options.

```blueprintlang
version "2025-11-02"

resource googleOidc: aws/iam/oidcProvider {
    metadata {
        displayName = "Google OIDC Provider"
        description = "OIDC provider for Google authentication"
        labels = {
            provider = "google"
            environment = "production"
        }
    }
    spec {
        url = "https://accounts.google.com"
        clientIdList = [
            "123456789012-abcdef.apps.googleusercontent.com",
            "123456789012-ghijkl.apps.googleusercontent.com"
        ]
        thumbprintList = [
            "cf23df2207d99a74fbe169e3eba035e633b65d94",
            "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"
        ]
        tags = [
            {
                key = "Environment"
                value = "Production"
            },
            {
                key = "Service"
                value = "Authentication"
            },
            {
                key = "Provider"
                value = "Google"
            },
            {
                key = "ManagedBy"
                value = "DevOps"
            }
        ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  googleOidc:
    type: aws/iam/oidcProvider
    metadata:
      displayName: Google OIDC Provider
      description: OIDC provider for Google authentication
      labels:
        provider: google
        environment: production
    spec:
      url: https://accounts.google.com
      clientIdList:
        - 123456789012-abcdef.apps.googleusercontent.com
        - 123456789012-ghijkl.apps.googleusercontent.com
      thumbprintList:
        - cf23df2207d99a74fbe169e3eba035e633b65d94
        - 9e99a48a9960b14926bb7f3b02e22da2b0ab7280
      tags:
        - key: Environment
          value: Production
        - key: Service
          value: Authentication
        - key: Provider
          value: Google
        - key: ManagedBy
          value: DevOps
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "googleOidc": {
      "type": "aws/iam/oidcProvider",
      "metadata": {
        "displayName": "Google OIDC Provider",
        "description": "OIDC provider for Google authentication",
        "labels": {
          "provider": "google",
          "environment": "production"
        }
      },
      "spec": {
        "url": "https://accounts.google.com",
        "clientIdList": [
          "123456789012-abcdef.apps.googleusercontent.com",
          "123456789012-ghijkl.apps.googleusercontent.com"
        ],
        "thumbprintList": [
          "cf23df2207d99a74fbe169e3eba035e633b65d94",
          "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"
        ],
        "tags": [
          {
            "key": "Environment",
            "value": "Production"
          },
          {
            "key": "Service",
            "value": "Authentication"
          },
          {
            "key": "Provider",
            "value": "Google"
          },
          {
            "key": "ManagedBy",
            "value": "DevOps"
          }
        ]
      }
    }
  }
}
```
