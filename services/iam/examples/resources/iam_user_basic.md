This example demonstrates creating a basic IAM user with minimal configuration.

```blueprintlang
version "2025-11-02"

resource myUser: aws/iam/user {
    metadata {
        displayName = "My Basic IAM User"
    }
    spec {
        userName = "my-basic-user"
        path = "/"
    }
}
```

```yaml
version: 2025-11-02

resources:
  myUser:
    type: aws/iam/user
    metadata:
      displayName: My Basic IAM User
    spec:
      userName: my-basic-user
      path: /
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "myUser": {
      "type": "aws/iam/user",
      "metadata": {
        "displayName": "My Basic IAM User"
      },
      "spec": {
        "userName": "my-basic-user",
        "path": "/"
      }
    }
  }
}
```
