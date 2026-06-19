This example demonstrates creating a basic IAM group with minimal configuration.

```blueprintlang
version "2025-11-02"

resource developers: aws/iam/group {
    metadata {
        displayName = "Developers Group"
    }
    spec {
        groupName = "developers"
        path = "/"
    }
}
```

```yaml
version: 2025-11-02

resources:
  developers:
    type: aws/iam/group
    metadata:
      displayName: Developers Group
    spec:
      groupName: developers
      path: /
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "developers": {
      "type": "aws/iam/group",
      "metadata": {
        "displayName": "Developers Group"
      },
      "spec": {
        "groupName": "developers",
        "path": "/"
      }
    }
  }
}
```
