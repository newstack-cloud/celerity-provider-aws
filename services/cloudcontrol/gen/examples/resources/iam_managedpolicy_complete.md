A AWS IAM ManagedPolicy configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource managedPolicy: aws/iam/managedPolicy {
    metadata {
        displayName = "AWS IAM ManagedPolicy complete"
    }
    spec {
        description = "example-description"
        groups = [
            "example-group"
        ]
        managedPolicyName = "example-managed-policy-name"
        path = "example-path"
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
        roles = [
            "example-role"
        ]
        users = [
            "example-user"
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    managedPolicy:
        type: aws/iam/managedPolicy
        metadata:
            displayName: AWS IAM ManagedPolicy complete
        spec:
            description: example-description
            groups:
                - example-group
            managedPolicyName: example-managed-policy-name
            path: example-path
            policyDocument:
                statement:
                    - action:
                        - s3:GetObject
                      effect: Allow
                      resource: arn:aws:s3:::example-bucket/*
                version: "2012-10-17"
            roles:
                - example-role
            users:
                - example-user
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "managedPolicy": {
      "type": "aws/iam/managedPolicy",
      "metadata": {
        "displayName": "AWS IAM ManagedPolicy complete"
      },
      "spec": {
        "description": "example-description",
        "groups": [
          "example-group"
        ],
        "managedPolicyName": "example-managed-policy-name",
        "path": "example-path",
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
        "roles": [
          "example-role"
        ],
        "users": [
          "example-user"
        ]
      }
    }
  }
}
```
