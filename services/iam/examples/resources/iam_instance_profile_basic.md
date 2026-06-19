This example demonstrates creating a basic IAM instance profile with minimal configuration.

```blueprintlang
version "2025-11-02"

resource myInstanceProfile: aws/iam/instanceProfile {
    metadata {
        displayName = "My Instance Profile"
    }
    spec {
        role = "MyRole"
    }
}
```

```yaml
version: 2025-11-02

resources:
  myInstanceProfile:
    type: aws/iam/instanceProfile
    metadata:
      displayName: My Instance Profile
    spec:
      role: MyRole
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "myInstanceProfile": {
      "type": "aws/iam/instanceProfile",
      "metadata": {
        "displayName": "My Instance Profile"
      },
      "spec": {
        "role": "MyRole"
      }
    }
  }
}
```
