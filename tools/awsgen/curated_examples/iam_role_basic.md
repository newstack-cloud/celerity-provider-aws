A basic AWS IAM Role assumable by the Lambda service, with a structured trust policy.

```blueprintlang
version "2025-11-02"

resource role: aws/iam/role {
    metadata {
        displayName = "Order processor role"
    }
    spec {
        roleName = "order-processor-role"
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
            assumeRolePolicyDocument:
                version: "2012-10-17"
                statement:
                    - effect: Allow
                      principal:
                          service: lambda.amazonaws.com
                      action: sts:AssumeRole
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
        }
      }
    }
  }
}
```
