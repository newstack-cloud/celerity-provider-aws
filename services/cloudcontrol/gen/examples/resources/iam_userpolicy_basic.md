A basic AWS IAM UserPolicy with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource userPolicy: aws/iam/userPolicy {
    metadata {
        displayName = "AWS IAM UserPolicy basic"
    }
    spec {
        policyName = "example-policy-name"
        userName = "example-user-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    userPolicy:
        type: aws/iam/userPolicy
        metadata:
            displayName: AWS IAM UserPolicy basic
        spec:
            policyName: example-policy-name
            userName: example-user-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "userPolicy": {
      "type": "aws/iam/userPolicy",
      "metadata": {
        "displayName": "AWS IAM UserPolicy basic"
      },
      "spec": {
        "policyName": "example-policy-name",
        "userName": "example-user-name"
      }
    }
  }
}
```
