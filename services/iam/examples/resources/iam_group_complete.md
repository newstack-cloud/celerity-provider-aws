This example demonstrates creating an IAM group with inline policies and managed policies.

```blueprintlang
version "2025-11-02"

resource developers: aws/iam/group {
    metadata {
        displayName = "Developers Group"
    }
    spec {
        groupName = "developers"
        path = "/"
        policies = [
            {
                policyName = "EC2ReadOnly"
                policyDocument = {
                    Version = "2012-10-17"
                    Statement = [
                        {
                            Effect = "Allow"
                            Action = [
                                "ec2:Describe*",
                                "ec2:Get*"
                            ]
                            Resource = "*"
                        }
                    ]
                }
            }
        ]
        managedPolicyArns = [
            "arn:aws:iam::aws:policy/ReadOnlyAccess"
        ]
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
      policies:
        - policyName: EC2ReadOnly
          policyDocument:
            Version: "2012-10-17"
            Statement:
              - Effect: Allow
                Action:
                  - "ec2:Describe*"
                  - "ec2:Get*"
                Resource: "*"
      managedPolicyArns:
        - "arn:aws:iam::aws:policy/ReadOnlyAccess"
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
        "path": "/",
        "policies": [
          {
            "policyName": "EC2ReadOnly",
            "policyDocument": {
              "Version": "2012-10-17",
              "Statement": [
                {
                  "Effect": "Allow",
                  "Action": [
                    "ec2:Describe*",
                    "ec2:Get*"
                  ],
                  "Resource": "*"
                }
              ]
            }
          }
        ],
        "managedPolicyArns": [
          "arn:aws:iam::aws:policy/ReadOnlyAccess"
        ]
      }
    }
  }
}
```
