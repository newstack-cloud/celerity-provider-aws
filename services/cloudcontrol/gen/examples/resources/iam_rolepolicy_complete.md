A AWS IAM RolePolicy configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource rolePolicy: aws/iam/rolePolicy {
    metadata {
        displayName = "AWS IAM RolePolicy complete"
    }
    spec {
        policyDocument = {
            statement = [
                {
                    action = [
                        "s3:GetObject"
                    ],
                    effect = "Allow",
                    resource = "arn:aws:s3:::example-bucket/*"
                }
            ],
            version = "2012-10-17"
        }
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
            displayName: AWS IAM RolePolicy complete
        spec:
            policyDocument:
                statement:
                    - action:
                        - s3:GetObject
                      effect: Allow
                      resource: arn:aws:s3:::example-bucket/*
                version: "2012-10-17"
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
        "displayName": "AWS IAM RolePolicy complete"
      },
      "spec": {
        "policyDocument": {
          "statement": [
            {
              "action": [
                "s3:GetObject"
              ],
              "effect": "Allow",
              "resource": "arn:aws:s3:::example-bucket/*"
            }
          ],
          "version": "2012-10-17"
        },
        "policyName": "example-policy-name",
        "roleName": "example-role-name"
      }
    }
  }
}
```
