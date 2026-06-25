A AWS IAM User configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource user: aws/iam/user {
    metadata {
        displayName = "AWS IAM User complete"
    }
    spec {
        groups = [
            "example-group"
        ]
        loginProfile = {
            password = "example-password",
            passwordResetRequired = false
        }
        managedPolicyArns = [
            "example-managed-policy-arn"
        ]
        path = "example-path"
        permissionsBoundary = "example-permissions-boundary"
        policies = [
            {
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
                },
                policyName = "example-policy-name"
            }
        ]
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        userName = "example-user-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    user:
        type: aws/iam/user
        metadata:
            displayName: AWS IAM User complete
        spec:
            groups:
                - example-group
            loginProfile:
                password: example-password
                passwordResetRequired: false
            managedPolicyArns:
                - example-managed-policy-arn
            path: example-path
            permissionsBoundary: example-permissions-boundary
            policies:
                - policyDocument:
                    statement:
                        - action:
                            - s3:GetObject
                          effect: Allow
                          resource: arn:aws:s3:::example-bucket/*
                    version: "2012-10-17"
                  policyName: example-policy-name
            tags:
                - key: example-key
                  value: example-value
            userName: example-user-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "user": {
      "type": "aws/iam/user",
      "metadata": {
        "displayName": "AWS IAM User complete"
      },
      "spec": {
        "groups": [
          "example-group"
        ],
        "loginProfile": {
          "password": "example-password",
          "passwordResetRequired": false
        },
        "managedPolicyArns": [
          "example-managed-policy-arn"
        ],
        "path": "example-path",
        "permissionsBoundary": "example-permissions-boundary",
        "policies": [
          {
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
            "policyName": "example-policy-name"
          }
        ],
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "userName": "example-user-name"
      }
    }
  }
}
```
