This example creates a comprehensive IAM managed policy with multiple statements and all optional fields.

```blueprintlang
version "2025-11-02"

resource ec2FullAccessPolicy: aws/iam/managedPolicy {
    metadata {
        displayName = "EC2 Full Access Policy"
    }
    spec {
        policyName = "EC2FullAccessPolicy"
        policyDocument = {
            Version = "2012-10-17"
            Statement = [
                {
                    Effect = "Allow"
                    Action = [
                        "ec2:DescribeInstances",
                        "ec2:DescribeSecurityGroups",
                        "ec2:DescribeVpcs",
                        "ec2:DescribeSubnets",
                        "ec2:DescribeRouteTables"
                    ]
                    Resource = "*"
                },
                {
                    Effect = "Allow"
                    Action = [
                        "ec2:RunInstances",
                        "ec2:StartInstances",
                        "ec2:StopInstances",
                        "ec2:TerminateInstances",
                        "ec2:RebootInstances"
                    ]
                    Resource = "*"
                    Condition = {
                        StringEquals = {
                            "aws:RequestTag/Environment" = "Production"
                        }
                    }
                }
            ]
        }
        description = "Policy that provides full access to EC2 resources with conditions"
        path = "/managed-policies/"
        tags = [
            {
                key = "Environment"
                value = "Production"
            },
            {
                key = "Project"
                value = "EC2Management"
            },
            {
                key = "Owner"
                value = "DevOps Team"
            }
        ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  ec2FullAccessPolicy:
    type: aws/iam/managedPolicy
    metadata:
      displayName: EC2 Full Access Policy
    spec:
      policyName: EC2FullAccessPolicy
      policyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: Allow
            Action:
              - ec2:DescribeInstances
              - ec2:DescribeSecurityGroups
              - ec2:DescribeVpcs
              - ec2:DescribeSubnets
              - ec2:DescribeRouteTables
            Resource: "*"
          - Effect: Allow
            Action:
              - ec2:RunInstances
              - ec2:StartInstances
              - ec2:StopInstances
              - ec2:TerminateInstances
              - ec2:RebootInstances
            Resource: "*"
            Condition:
              StringEquals:
                aws:RequestTag/Environment: Production
      description: Policy that provides full access to EC2 resources with conditions
      path: /managed-policies/
      tags:
        - key: Environment
          value: Production
        - key: Project
          value: EC2Management
        - key: Owner
          value: DevOps Team
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "ec2FullAccessPolicy": {
      "type": "aws/iam/managedPolicy",
      "metadata": {
        "displayName": "EC2 Full Access Policy"
      },
      "spec": {
        "policyName": "EC2FullAccessPolicy",
        "policyDocument": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Action": [
                "ec2:DescribeInstances",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeVpcs",
                "ec2:DescribeSubnets",
                "ec2:DescribeRouteTables"
              ],
              "Resource": "*"
            },
            {
              "Effect": "Allow",
              "Action": [
                "ec2:RunInstances",
                "ec2:StartInstances",
                "ec2:StopInstances",
                "ec2:TerminateInstances",
                "ec2:RebootInstances"
              ],
              "Resource": "*",
              "Condition": {
                "StringEquals": {
                  "aws:RequestTag/Environment": "Production"
                }
              }
            }
          ]
        },
        "description": "Policy that provides full access to EC2 resources with conditions",
        "path": "/managed-policies/",
        "tags": [
          {
            "key": "Environment",
            "value": "Production"
          },
          {
            "key": "Project",
            "value": "EC2Management"
          },
          {
            "key": "Owner",
            "value": "DevOps Team"
          }
        ]
      }
    }
  }
}
```
