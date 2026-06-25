A basic AWS IAM AccessKey with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource accessKey: aws/iam/accessKey {
    metadata {
        displayName = "AWS IAM AccessKey basic"
    }
    spec {
        userName = "example-user-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    accessKey:
        type: aws/iam/accessKey
        metadata:
            displayName: AWS IAM AccessKey basic
        spec:
            userName: example-user-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "accessKey": {
      "type": "aws/iam/accessKey",
      "metadata": {
        "displayName": "AWS IAM AccessKey basic"
      },
      "spec": {
        "userName": "example-user-name"
      }
    }
  }
}
```
