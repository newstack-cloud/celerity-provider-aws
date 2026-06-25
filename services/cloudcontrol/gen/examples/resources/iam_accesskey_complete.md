A AWS IAM AccessKey configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource accessKey: aws/iam/accessKey {
    metadata {
        displayName = "AWS IAM AccessKey complete"
    }
    spec {
        serial = 1
        status = "example-status"
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
            displayName: AWS IAM AccessKey complete
        spec:
            serial: 1
            status: example-status
            userName: example-user-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "accessKey": {
      "type": "aws/iam/accessKey",
      "metadata": {
        "displayName": "AWS IAM AccessKey complete"
      },
      "spec": {
        "serial": 1,
        "status": "example-status",
        "userName": "example-user-name"
      }
    }
  }
}
```
