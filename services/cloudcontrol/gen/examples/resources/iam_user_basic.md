A basic AWS IAM User with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource user: aws/iam/user {
    metadata {
        displayName = "AWS IAM User basic"
    }
    spec {
        userName = "example-user-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    user:
        type: aws/iam/user
        metadata:
            displayName: AWS IAM User basic
        spec:
            userName: example-user-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "user": {
      "type": "aws/iam/user",
      "metadata": {
        "displayName": "AWS IAM User basic"
      },
      "spec": {
        "userName": "example-user-name"
      }
    }
  }
}
```
