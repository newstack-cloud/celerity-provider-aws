A complete IAM access key with all available options.

```blueprintlang
version "2025-11-02"

resource adminAccessKey: aws/iam/accessKey {
    metadata {
        displayName = "Admin Access Key"
    }
    spec {
        userName = "admin.user"
        status = "Active"
    }
}
```

```yaml
version: 2025-11-02

resources:
  adminAccessKey:
    type: aws/iam/accessKey
    metadata:
      displayName: Admin Access Key
    spec:
      userName: admin.user
      status: Active
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "adminAccessKey": {
      "type": "aws/iam/accessKey",
      "metadata": {
        "displayName": "Admin Access Key"
      },
      "spec": {
        "userName": "admin.user",
        "status": "Active"
      }
    }
  }
}
```
