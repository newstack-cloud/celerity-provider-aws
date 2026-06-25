A basic AWS IAM SAMLProvider with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource sAMLProvider: aws/iam/samlProvider {
    metadata {
        displayName = "AWS IAM SAMLProvider basic"
    }
    spec {
        addPrivateKey = "example-add-private-key"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    sAMLProvider:
        type: aws/iam/samlProvider
        metadata:
            displayName: AWS IAM SAMLProvider basic
        spec:
            addPrivateKey: example-add-private-key
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "sAMLProvider": {
      "type": "aws/iam/samlProvider",
      "metadata": {
        "displayName": "AWS IAM SAMLProvider basic"
      },
      "spec": {
        "addPrivateKey": "example-add-private-key"
      }
    }
  }
}
```
