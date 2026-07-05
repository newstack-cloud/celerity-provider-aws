## Lambda Function to Secrets Manager Secret

Grants a Lambda function access to a Secrets Manager secret.

When a function links to a secret, the function's execution role is granted access to the secret and (by default) an environment variable holding the secret's ARN is added to the function so your code can retrieve it at runtime without manual configuration or hardcoding.

The access level is controlled with the `aws.lambda.secretsmanager.<targetSecret>.accessLevel` annotation:

- `read` (default) — `secretsmanager:GetSecretValue`, `secretsmanager:DescribeSecret`
- `write` — `secretsmanager:PutSecretValue`, `secretsmanager:UpdateSecret`
- `readwrite` — all of the above

Each statement grants access to the specific secret ARN.

The function's execution role must be defined as a resource in the same blueprint; the link adds the access permission to it.

If the function runs inside a VPC without internet access, the link also activates a Secrets Manager interface VPC endpoint on the linked flex VPC so the function can reach Secrets Manager over the private network.

Secrets encrypted with a customer managed KMS key additionally require the execution role to hold `kms:Decrypt` on that key; link the function to the `aws/kms/key` as well (or grant it via the key policy).

### Example

```blueprintlang
version "2025-11-02"

resource apiFunction: aws/lambda/function {
    metadata {
        labels = {
            secret = "db-credentials"
        }
        annotations = {
            "aws.lambda.secretsmanager.dbCredentials.envVarName" = "DB_CREDENTIALS_SECRET"
            "aws.lambda.secretsmanager.dbCredentials.accessLevel" = "read"
        }
    }

    select by label {
        secret = "db-credentials"
    }

    spec {
        functionName = "api"
        role = apiFunctionRole.spec.arn
        # ... other function configuration
    }
}

resource dbCredentials: aws/secretsmanager/secret {
    metadata {
        labels = {
            secret = "db-credentials"
        }
    }

    spec {
        name = "prod/db-credentials"
    }
}

resource apiFunctionRole: aws/iam/role {
    spec {
        name = "api-role"
        # ... role configuration
    }
}
```
