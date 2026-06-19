This example demonstrates how to create a comprehensive IAM role with all available configuration options.

```blueprintlang
version "2025-11-02"

value customDynamoDBAccessPolicy: object {
    value = {
        policyName = "CustomDynamoDBAccess"
        policyDocument = {
            Version = "2012-10-17"
            Statement = [
                {
                    Effect = "Allow"
                    Action = [
                        "dynamodb:GetItem",
                        "dynamodb:PutItem",
                        "dynamodb:UpdateItem",
                        "dynamodb:DeleteItem",
                        "dynamodb:Query",
                        "dynamodb:Scan"
                    ]
                    Resource = "arn:aws:dynamodb:*:*:table/MyTable"
                }
            ]
        }
    }
}

resource comprehensiveRole: aws/iam/role {
    metadata {
        displayName = "Comprehensive IAM Role"
        description = "IAM role demonstrating all available configuration options"
        labels = {
            app = "example"
            tier = "comprehensive"
        }
    }
    spec {
        roleName = "comprehensive-service-role"
        description = "A comprehensive example role with all configuration options"
        assumeRolePolicyDocument = {
            Version = "2012-10-17"
            Statement = [
                {
                    Effect = "Allow"
                    Principal = {
                        Service = [
                            "lambda.amazonaws.com",
                            "ec2.amazonaws.com"
                        ]
                    }
                    Action = [
                        "sts:AssumeRole"
                    ]
                    Condition = {
                        StringEquals = {
                            "sts:ExternalId" = "unique-external-id"
                        }
                    }
                }
            ]
        }
        path = "/service-roles/"
        maxSessionDuration = 7200
        managedPolicyArns = [
            "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
            "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess",
            "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
        ]
        policies = [
            values.customDynamoDBAccessPolicy,
            {
                policyName = "CustomSQSAccess"
                policyDocument = {
                    Version = "2012-10-17"
                    Statement = [
                        {
                            Effect = "Allow"
                            Action = [
                                "sqs:SendMessage",
                                "sqs:ReceiveMessage",
                                "sqs:DeleteMessage"
                            ]
                            Resource = "arn:aws:sqs:*:*:my-queue"
                        }
                    ]
                }
            }
        ]
        permissionsBoundary = "arn:aws:iam::123456789012:policy/DeveloperBoundary"
        tags = {
            Environment = "Production"
            Application = "MyApplication"
            Team = "DevOps"
            CostCenter = "Engineering"
            Owner = "admin@example.com"
        }
    }
}
```

```yaml
version: 2025-11-02

values:
  customDynamoDBAccessPolicy:
    type: object
    value:
      policyName: CustomDynamoDBAccess
      policyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: Allow
            Action:
              - dynamodb:GetItem
              - dynamodb:PutItem
              - dynamodb:UpdateItem
              - dynamodb:DeleteItem
              - dynamodb:Query
              - dynamodb:Scan
            Resource: arn:aws:dynamodb:*:*:table/MyTable

resources:
  comprehensiveRole:
    type: aws/iam/role
    metadata:
      displayName: Comprehensive IAM Role
      description: IAM role demonstrating all available configuration options
      labels:
        app: example
        tier: comprehensive
    spec:
      roleName: comprehensive-service-role
      description: A comprehensive example role with all configuration options
      assumeRolePolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Effect: Allow
            Principal:
              Service:
                - lambda.amazonaws.com
                - ec2.amazonaws.com
            Action:
              - sts:AssumeRole
            Condition:
              StringEquals:
                sts:ExternalId: unique-external-id
      path: /service-roles/
      maxSessionDuration: 7200
      managedPolicyArns:
        - arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
        - arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess
        - arn:aws:iam::aws:policy/CloudWatchLogsFullAccess
      policies:
        - ${values.customDynamoDBAccessPolicy}
        - policyName: CustomSQSAccess
          policyDocument:
            Version: "2012-10-17"
            Statement:
              - Effect: Allow
                Action:
                  - sqs:SendMessage
                  - sqs:ReceiveMessage
                  - sqs:DeleteMessage
                Resource: arn:aws:sqs:*:*:my-queue
      permissionsBoundary: arn:aws:iam::123456789012:policy/DeveloperBoundary
      tags:
        Environment: Production
        Application: MyApplication
        Team: DevOps
        CostCenter: Engineering
        Owner: admin@example.com
```

```javascript
{
  "version": "2025-11-02",
  "values": {
    "customDynamoDBAccessPolicy": {
      "type": "object",
      "value": {
        "policyName": "CustomDynamoDBAccess",
        "policyDocument": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Action": [
                "dynamodb:GetItem",
                "dynamodb:PutItem",
                "dynamodb:UpdateItem",
                "dynamodb:DeleteItem",
                "dynamodb:Query",
                "dynamodb:Scan"
              ],
              "Resource": "arn:aws:dynamodb:*:*:table/MyTable"
            }
          ]
        }
      }
    }
  },
  "resources": {
    "comprehensiveRole": {
      "type": "aws/iam/role",
      "metadata": {
        "displayName": "Comprehensive IAM Role",
        "description": "IAM role demonstrating all available configuration options",
        "labels": {
          "app": "example",
          "tier": "comprehensive"
        }
      },
      "spec": {
        "roleName": "comprehensive-service-role",
        "description": "A comprehensive example role with all configuration options",
        "assumeRolePolicyDocument": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Principal": {
                "Service": [
                  "lambda.amazonaws.com",
                  "ec2.amazonaws.com"
                ]
              },
              "Action": [
                "sts:AssumeRole"
              ],
              "Condition": {
                "StringEquals": {
                  "sts:ExternalId": "unique-external-id"
                }
              }
            }
          ]
        },
        "path": "/service-roles/",
        "maxSessionDuration": 7200,
        "managedPolicyArns": [
          "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
          "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess",
          "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess"
        ],
        "policies": [
          "${values.customDynamoDBAccessPolicy}",
          {
            "policyName": "CustomSQSAccess",
            "policyDocument": {
              "Version": "2012-10-17",
              "Statement": [
                {
                  "Effect": "Allow",
                  "Action": [
                    "sqs:SendMessage",
                    "sqs:ReceiveMessage",
                    "sqs:DeleteMessage"
                  ],
                  "Resource": "arn:aws:sqs:*:*:my-queue"
                }
              ]
            }
          }
        ],
        "permissionsBoundary": "arn:aws:iam::123456789012:policy/DeveloperBoundary",
        "tags": {
          "Environment": "Production",
          "Application": "MyApplication",
          "Team": "DevOps",
          "CostCenter": "Engineering",
          "Owner": "admin@example.com"
        }
      }
    }
  }
}
```
