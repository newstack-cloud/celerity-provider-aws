A basic AWS IAM OIDCProvider with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource oIDCProvider: aws/iam/oidcProvider {
    metadata {
        displayName = "AWS IAM OIDCProvider basic"
    }
    spec {
        url = "example-url"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    oIDCProvider:
        type: aws/iam/oidcProvider
        metadata:
            displayName: AWS IAM OIDCProvider basic
        spec:
            url: example-url
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "oIDCProvider": {
      "type": "aws/iam/oidcProvider",
      "metadata": {
        "displayName": "AWS IAM OIDCProvider basic"
      },
      "spec": {
        "url": "example-url"
      }
    }
  }
}
```
