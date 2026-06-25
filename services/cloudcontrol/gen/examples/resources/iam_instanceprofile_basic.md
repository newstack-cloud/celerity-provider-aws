A basic AWS IAM InstanceProfile with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource instanceProfile: aws/iam/instanceProfile {
    metadata {
        displayName = "AWS IAM InstanceProfile basic"
    }
    spec {
        roles = [
            "example-role"
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    instanceProfile:
        type: aws/iam/instanceProfile
        metadata:
            displayName: AWS IAM InstanceProfile basic
        spec:
            roles:
                - example-role
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "instanceProfile": {
      "type": "aws/iam/instanceProfile",
      "metadata": {
        "displayName": "AWS IAM InstanceProfile basic"
      },
      "spec": {
        "roles": [
          "example-role"
        ]
      }
    }
  }
}
```
