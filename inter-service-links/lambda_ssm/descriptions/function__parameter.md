## Lambda Function to SSM Parameter

Grants a Lambda function access to a Systems Manager (SSM) parameter.

When a function links to a parameter, the function's execution role is granted access to the parameter and (by default) an environment variable holding the parameter's name is added to the function so your code can fetch it at runtime with `ssm:GetParameter` without manual configuration or hardcoding.

The access level is controlled with the `aws.lambda.ssm.<targetParameter>.accessLevel` annotation:

- `read` (default) — `ssm:GetParameter`, `ssm:GetParameters`, `ssm:GetParametersByPath`
- `write` — `ssm:PutParameter`
- `readwrite` — all of the above

Each statement grants access to the specific parameter ARN.

The function's execution role must be defined as a resource in the same blueprint; the link adds the access permission to it.

If the function runs inside a VPC without internet access, the link also activates an SSM interface VPC endpoint on the linked flex VPC so the function can reach Parameter Store over the private network.

`SecureString` parameters encrypted with a customer managed KMS key additionally require the execution role to hold `kms:Decrypt` on that key; link the function to the `aws/kms/key` as well (or grant it via the key policy). Parameters using the default `alias/aws/ssm` key need no additional grant.

### Example

```blueprintlang
version "2025-11-02"

resource apiFunction: aws/lambda/function {
    metadata {
        labels = {
            parameter = "db-host"
        }
        annotations = {
            "aws.lambda.ssm.dbHost.envVarName" = "DB_HOST_PARAM"
            "aws.lambda.ssm.dbHost.accessLevel" = "read"
        }
    }

    select by label {
        parameter = "db-host"
    }

    spec {
        functionName = "api"
        role = apiFunctionRole.spec.arn
        # ... other function configuration
    }
}

resource dbHost: aws/ssm/parameter {
    metadata {
        labels = {
            parameter = "db-host"
        }
    }

    spec {
        name = "/my-app/prod/db-host"
        type = "String"
        value = "db.internal.example.com"
    }
}

resource apiFunctionRole: aws/iam/role {
    spec {
        name = "api-role"
        # ... role configuration
    }
}
```
