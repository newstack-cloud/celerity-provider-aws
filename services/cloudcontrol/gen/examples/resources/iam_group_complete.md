A AWS IAM Group configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource group: aws/iam/group {
    metadata {
        displayName = "AWS IAM Group complete"
    }
    spec {
        groupName = "example-group-name"
        managedPolicyArns = [
            "example-managed-policy-arn"
        ]
        path = "example-path"
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
    }
}
```

```yaml
version: "2025-11-02"
resources:
    group:
        type: aws/iam/group
        metadata:
            displayName: AWS IAM Group complete
        spec:
            groupName: example-group-name
            managedPolicyArns:
                - example-managed-policy-arn
            path: example-path
            policies:
                - policyDocument:
                    statement:
                        - action:
                            - s3:GetObject
                          effect: Allow
                          resource: arn:aws:s3:::example-bucket/*
                    version: "2012-10-17"
                  policyName: example-policy-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "group": {
      "type": "aws/iam/group",
      "metadata": {
        "displayName": "AWS IAM Group complete"
      },
      "spec": {
        "groupName": "example-group-name",
        "managedPolicyArns": [
          "example-managed-policy-arn"
        ],
        "path": "example-path",
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
        ]
      }
    }
  }
}
```
