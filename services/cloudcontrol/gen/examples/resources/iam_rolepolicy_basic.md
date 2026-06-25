A basic AWS IAM RolePolicy with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource rolePolicy: aws/iam/rolePolicy {
    metadata {
        displayName = "AWS IAM RolePolicy basic"
    }
    spec {
        policyName = "example-policy-name"
        roleName = "example-role-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    rolePolicy:
        type: aws/iam/rolePolicy
        metadata:
            displayName: AWS IAM RolePolicy basic
        spec:
            policyName: example-policy-name
            roleName: example-role-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "rolePolicy": {
      "type": "aws/iam/rolePolicy",
      "metadata": {
        "displayName": "AWS IAM RolePolicy basic"
      },
      "spec": {
        "policyName": "example-policy-name",
        "roleName": "example-role-name"
      }
    }
  }
}
```
