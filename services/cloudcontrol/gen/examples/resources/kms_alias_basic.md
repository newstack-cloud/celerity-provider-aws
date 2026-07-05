A basic AWS KMS Alias with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource alias: aws/kms/alias {
    metadata {
        displayName = "AWS KMS Alias basic"
    }
    spec {
        aliasName = "example-alias-name"
        targetKeyId = "example-target-key-id"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    alias:
        type: aws/kms/alias
        metadata:
            displayName: AWS KMS Alias basic
        spec:
            aliasName: example-alias-name
            targetKeyId: example-target-key-id
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "alias": {
      "type": "aws/kms/alias",
      "metadata": {
        "displayName": "AWS KMS Alias basic"
      },
      "spec": {
        "aliasName": "example-alias-name",
        "targetKeyId": "example-target-key-id"
      }
    }
  }
}
```
