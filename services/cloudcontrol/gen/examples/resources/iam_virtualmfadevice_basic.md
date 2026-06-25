A basic AWS IAM VirtualMFADevice with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource virtualMFADevice: aws/iam/virtualMFADevice {
    metadata {
        displayName = "AWS IAM VirtualMFADevice basic"
    }
    spec {
        users = [
            "example-user"
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    virtualMFADevice:
        type: aws/iam/virtualMFADevice
        metadata:
            displayName: AWS IAM VirtualMFADevice basic
        spec:
            users:
                - example-user
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "virtualMFADevice": {
      "type": "aws/iam/virtualMFADevice",
      "metadata": {
        "displayName": "AWS IAM VirtualMFADevice basic"
      },
      "spec": {
        "users": [
          "example-user"
        ]
      }
    }
  }
}
```
