A AWS IAM VirtualMFADevice configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource virtualMFADevice: aws/iam/virtualMFADevice {
    metadata {
        displayName = "AWS IAM VirtualMFADevice complete"
    }
    spec {
        path = "example-path"
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        users = [
            "example-user"
        ]
        virtualMfaDeviceName = "example-virtual-mfa-device-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    virtualMFADevice:
        type: aws/iam/virtualMFADevice
        metadata:
            displayName: AWS IAM VirtualMFADevice complete
        spec:
            path: example-path
            tags:
                - key: example-key
                  value: example-value
            users:
                - example-user
            virtualMfaDeviceName: example-virtual-mfa-device-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "virtualMFADevice": {
      "type": "aws/iam/virtualMFADevice",
      "metadata": {
        "displayName": "AWS IAM VirtualMFADevice complete"
      },
      "spec": {
        "path": "example-path",
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "users": [
          "example-user"
        ],
        "virtualMfaDeviceName": "example-virtual-mfa-device-name"
      }
    }
  }
}
```
