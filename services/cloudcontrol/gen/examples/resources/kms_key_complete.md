A AWS KMS Key configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource "key": aws/kms/key {
    metadata {
        displayName = "AWS KMS Key complete"
    }
    spec {
        bypassPolicyLockoutSafetyCheck = false
        description = "example-description"
        enableKeyRotation = false
        enabled = false
        keyPolicy = {
            exampleKey = "example-value"
        }
        keySpec = "SYMMETRIC_DEFAULT"
        keyUsage = "ENCRYPT_DECRYPT"
        multiRegion = false
        origin = "AWS_KMS"
        pendingWindowInDays = 7
        rotationPeriodInDays = 90
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    key:
        type: aws/kms/key
        metadata:
            displayName: AWS KMS Key complete
        spec:
            bypassPolicyLockoutSafetyCheck: false
            description: example-description
            enableKeyRotation: false
            enabled: false
            keyPolicy:
                exampleKey: example-value
            keySpec: SYMMETRIC_DEFAULT
            keyUsage: ENCRYPT_DECRYPT
            multiRegion: false
            origin: AWS_KMS
            pendingWindowInDays: 7
            rotationPeriodInDays: 90
            tags:
                - key: example-key
                  value: example-value
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "key": {
      "type": "aws/kms/key",
      "metadata": {
        "displayName": "AWS KMS Key complete"
      },
      "spec": {
        "bypassPolicyLockoutSafetyCheck": false,
        "description": "example-description",
        "enableKeyRotation": false,
        "enabled": false,
        "keyPolicy": {
          "exampleKey": "example-value"
        },
        "keySpec": "SYMMETRIC_DEFAULT",
        "keyUsage": "ENCRYPT_DECRYPT",
        "multiRegion": false,
        "origin": "AWS_KMS",
        "pendingWindowInDays": 7,
        "rotationPeriodInDays": 90,
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ]
      }
    }
  }
}
```
