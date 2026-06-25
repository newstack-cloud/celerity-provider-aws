A AWS IAM InstanceProfile configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource instanceProfile: aws/iam/instanceProfile {
    metadata {
        displayName = "AWS IAM InstanceProfile complete"
    }
    spec {
        instanceProfileName = "example-instance-profile-name"
        path = "example-path"
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
            displayName: AWS IAM InstanceProfile complete
        spec:
            instanceProfileName: example-instance-profile-name
            path: example-path
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
        "displayName": "AWS IAM InstanceProfile complete"
      },
      "spec": {
        "instanceProfileName": "example-instance-profile-name",
        "path": "example-path",
        "roles": [
          "example-role"
        ]
      }
    }
  }
}
```
