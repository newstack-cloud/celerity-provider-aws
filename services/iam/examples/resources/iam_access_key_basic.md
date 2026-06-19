A basic IAM access key for a user.

```blueprintlang
version "2025-11-02"

resource johnAccessKey: aws/iam/accessKey {
    metadata {
        displayName = "John's Access Key"
    }
    spec {
        userName = "john.doe"
    }
}
```

```yaml
version: 2025-11-02

resources:
  johnAccessKey:
    type: aws/iam/accessKey
    metadata:
      displayName: John's Access Key
    spec:
      userName: john.doe
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "johnAccessKey": {
      "type": "aws/iam/accessKey",
      "metadata": {
        "displayName": "John's Access Key"
      },
      "spec": {
        "userName": "john.doe"
      }
    }
  }
}
```
