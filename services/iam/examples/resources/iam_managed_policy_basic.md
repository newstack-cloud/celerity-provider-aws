This example creates a simple IAM managed policy that allows read-only access to S3 buckets.

```blueprintlang
version "2025-11-02"

resource s3ReadOnlyPolicy: aws/iam/managedPolicy {
    metadata {
        displayName = "S3 Read Only Policy"
    }
    spec {
        policyName = "S3ReadOnlyPolicy"
        policyDocument = {
            Version = "2012-10-17"
            Statement = [
                {
                    Effect = "Allow"
                    Action = [
                        "s3:GetObject",
                        "s3:ListBucket"
                    ]
                    Resource = [
                        "arn:aws:s3:::my-bucket",
                        "arn:aws:s3:::my-bucket/*"
                    ]
                }
            ]
        }
        description = "Policy that allows read-only access to S3 bucket"
        path = "/"
        tags = [
            {
                key = "Environment"
                value = "Production"
            },
            {
                key = "Project"
                value = "MyProject"
            }
        ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  s3ReadOnlyPolicy:
    type: aws/iam/managedPolicy
    metadata:
      displayName: S3 Read Only Policy
    spec:
      policyName: S3ReadOnlyPolicy
      policyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: Allow
            Action:
              - s3:GetObject
              - s3:ListBucket
            Resource:
              - arn:aws:s3:::my-bucket
              - arn:aws:s3:::my-bucket/*
      description: Policy that allows read-only access to S3 bucket
      path: /
      tags:
        - key: Environment
          value: Production
        - key: Project
          value: MyProject
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "s3ReadOnlyPolicy": {
      "type": "aws/iam/managedPolicy",
      "metadata": {
        "displayName": "S3 Read Only Policy"
      },
      "spec": {
        "policyName": "S3ReadOnlyPolicy",
        "policyDocument": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Action": [
                "s3:GetObject",
                "s3:ListBucket"
              ],
              "Resource": [
                "arn:aws:s3:::my-bucket",
                "arn:aws:s3:::my-bucket/*"
              ]
            }
          ]
        },
        "description": "Policy that allows read-only access to S3 bucket",
        "path": "/",
        "tags": [
          {
            "key": "Environment",
            "value": "Production"
          },
          {
            "key": "Project",
            "value": "MyProject"
          }
        ]
      }
    }
  }
}
```
