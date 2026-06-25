A basic AWS IAM GroupPolicy with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource groupPolicy: aws/iam/groupPolicy {
    metadata {
        displayName = "AWS IAM GroupPolicy basic"
    }
    spec {
        groupName = "example-group-name"
        policyName = "example-policy-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    groupPolicy:
        type: aws/iam/groupPolicy
        metadata:
            displayName: AWS IAM GroupPolicy basic
        spec:
            groupName: example-group-name
            policyName: example-policy-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "groupPolicy": {
      "type": "aws/iam/groupPolicy",
      "metadata": {
        "displayName": "AWS IAM GroupPolicy basic"
      },
      "spec": {
        "groupName": "example-group-name",
        "policyName": "example-policy-name"
      }
    }
  }
}
```
