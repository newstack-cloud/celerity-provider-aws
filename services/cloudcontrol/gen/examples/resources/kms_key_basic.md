A basic AWS KMS Key with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource "key": aws/kms/key {
    metadata {
        displayName = "AWS KMS Key basic"
    }
    spec {
        bypassPolicyLockoutSafetyCheck = false
    }
}
```

```yaml
version: "2025-11-02"
resources:
    key:
        type: aws/kms/key
        metadata:
            displayName: AWS KMS Key basic
        spec:
            bypassPolicyLockoutSafetyCheck: false
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "key": {
      "type": "aws/kms/key",
      "metadata": {
        "displayName": "AWS KMS Key basic"
      },
      "spec": {
        "bypassPolicyLockoutSafetyCheck": false
      }
    }
  }
}
```
