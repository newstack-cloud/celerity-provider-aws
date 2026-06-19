A basic IAM OIDC provider for GitHub Actions.

```blueprintlang
version "2025-11-02"

resource githubActionsOidc: aws/iam/oidcProvider {
    metadata {
        displayName = "GitHub Actions OIDC Provider"
    }
    spec {
        url = "https://token.actions.githubusercontent.com"
        clientIdList = [
            "sts.amazonaws.com"
        ]
        thumbprintList = [
            "cf23df2207d99a74fbe169e3eba035e633b65d94"
        ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  githubActionsOidc:
    type: aws/iam/oidcProvider
    metadata:
      displayName: GitHub Actions OIDC Provider
    spec:
      url: https://token.actions.githubusercontent.com
      clientIdList:
        - sts.amazonaws.com
      thumbprintList:
        - cf23df2207d99a74fbe169e3eba035e633b65d94
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "githubActionsOidc": {
      "type": "aws/iam/oidcProvider",
      "metadata": {
        "displayName": "GitHub Actions OIDC Provider"
      },
      "spec": {
        "url": "https://token.actions.githubusercontent.com",
        "clientIdList": [
          "sts.amazonaws.com"
        ],
        "thumbprintList": [
          "cf23df2207d99a74fbe169e3eba035e633b65d94"
        ]
      }
    }
  }
}
```
