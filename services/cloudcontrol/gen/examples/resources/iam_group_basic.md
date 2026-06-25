A basic AWS IAM Group with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource group: aws/iam/group {
    metadata {
        displayName = "AWS IAM Group basic"
    }
    spec {
        groupName = "example-group-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    group:
        type: aws/iam/group
        metadata:
            displayName: AWS IAM Group basic
        spec:
            groupName: example-group-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "group": {
      "type": "aws/iam/group",
      "metadata": {
        "displayName": "AWS IAM Group basic"
      },
      "spec": {
        "groupName": "example-group-name"
      }
    }
  }
}
```
