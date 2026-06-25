A complete AWS IAM Role with a structured trust policy, an inline access policy, an attached managed policy, a permissions boundary and tags.

```blueprintlang
version "2025-11-02"

resource role: aws/iam/role {
    metadata {
        displayName = "Order processor role"
    }
    spec {
        roleName = "order-processor-role"
        path = "/service-roles/"
        description = "Execution role for the order processor service"
        maxSessionDuration = 3600
        permissionsBoundary = "arn:aws:iam::123456789012:policy/service-boundary"
        managedPolicyArns = [
            "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
        ]
        assumeRolePolicyDocument = {
            version = "2012-10-17",
            statement = [
                {
                    effect = "Allow",
                    principal = {
                        service = "lambda.amazonaws.com"
                    },
                    action = "sts:AssumeRole"
                }
            ]
        }
        policies = [
            {
                policyName = "orders-table-access",
                policyDocument = {
                    version = "2012-10-17",
                    statement = [
                        {
                            sid = "ReadWriteOrders",
                            effect = "Allow",
                            action = [
                                "dynamodb:GetItem",
                                "dynamodb:PutItem",
                                "dynamodb:Query"
                            ],
                            resource = "arn:aws:dynamodb:us-east-1:123456789012:table/orders",
                            condition = {
                                StringEquals = {
                                    "aws:RequestedRegion" = "us-east-1"
                                }
                            }
                        }
                    ]
                }
            }
        ]
        tags = [
            {
                key = "service",
                value = "orders"
            }
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    role:
        type: aws/iam/role
        metadata:
            displayName: Order processor role
        spec:
            roleName: order-processor-role
            path: /service-roles/
            description: Execution role for the order processor service
            maxSessionDuration: 3600
            permissionsBoundary: arn:aws:iam::123456789012:policy/service-boundary
            managedPolicyArns:
                - arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
            assumeRolePolicyDocument:
                version: "2012-10-17"
                statement:
                    - effect: Allow
                      principal:
                          service: lambda.amazonaws.com
                      action: sts:AssumeRole
            policies:
                - policyName: orders-table-access
                  policyDocument:
                      version: "2012-10-17"
                      statement:
                          - sid: ReadWriteOrders
                            effect: Allow
                            action:
                                - dynamodb:GetItem
                                - dynamodb:PutItem
                                - dynamodb:Query
                            resource: arn:aws:dynamodb:us-east-1:123456789012:table/orders
                            condition:
                                StringEquals:
                                    aws:RequestedRegion: us-east-1
            tags:
                - key: service
                  value: orders
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "role": {
      "type": "aws/iam/role",
      "metadata": {
        "displayName": "Order processor role"
      },
      "spec": {
        "roleName": "order-processor-role",
        "path": "/service-roles/",
        "description": "Execution role for the order processor service",
        "maxSessionDuration": 3600,
        "permissionsBoundary": "arn:aws:iam::123456789012:policy/service-boundary",
        "managedPolicyArns": [
          "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
        ],
        "assumeRolePolicyDocument": {
          "version": "2012-10-17",
          "statement": [
            {
              "effect": "Allow",
              "principal": {
                "service": "lambda.amazonaws.com"
              },
              "action": "sts:AssumeRole"
            }
          ]
        },
        "policies": [
          {
            "policyName": "orders-table-access",
            "policyDocument": {
              "version": "2012-10-17",
              "statement": [
                {
                  "sid": "ReadWriteOrders",
                  "effect": "Allow",
                  "action": [
                    "dynamodb:GetItem",
                    "dynamodb:PutItem",
                    "dynamodb:Query"
                  ],
                  "resource": "arn:aws:dynamodb:us-east-1:123456789012:table/orders",
                  "condition": {
                    "StringEquals": {
                      "aws:RequestedRegion": "us-east-1"
                    }
                  }
                }
              ]
            }
          }
        ],
        "tags": [
          {
            "key": "service",
            "value": "orders"
          }
        ]
      }
    }
  }
}
```
