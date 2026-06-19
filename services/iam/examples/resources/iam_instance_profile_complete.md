This example demonstrates creating an IAM instance profile with all available configuration options.

```blueprintlang
version "2025-11-02"

resource myInstanceProfile: aws/iam/instanceProfile {
    metadata {
        displayName = "My Instance Profile"
    }
    spec {
        instanceProfileName = "MyInstanceProfile"
        path = "/"
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
      instanceProfileName: MyInstanceProfile
      path: /
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
        "instanceProfileName": "MyInstanceProfile",
        "path": "/",
        "role": "MyRole"
      }
    }
  }
}
```
