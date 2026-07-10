## Lambda Function to SSM Parameter Path

Grants a Lambda function access to every Systems Manager (SSM) parameter beneath a path prefix.

An `aws/ssm/parameterPath` resource represents a parameter hierarchy (such as `/my-app/config`) as a whole. When a function links to it, the function's execution role is granted access to the entire namespace with a single statement, and (by default) an environment variable holding the path prefix is added to the function so your code can enumerate the namespace at runtime with `ssm:GetParametersByPath` (`Recursive: true`) without manual configuration or hardcoding.

This is the right link when a function reads a whole configuration namespace of many parameters. Linking each parameter individually with `aws/lambda/function::aws/ssm/parameter` produces one statement and one environment variable per parameter, and per-parameter grants never authorise the `ssm:GetParametersByPath` enumerate call over the containing path. With this link, one statement and one environment variable cover the namespace regardless of how many parameters live beneath it.

The access level is controlled with the `aws.lambda.ssm.<targetParameterPath>.accessLevel` annotation:

- `read` (default) — `ssm:GetParameter`, `ssm:GetParameters`, `ssm:GetParametersByPath`
- `write` — `ssm:PutParameter`
- `readwrite` — all of the above

Each statement grants access to two ARNs: the path itself (which authorises `ssm:GetParametersByPath` over the hierarchy) and everything beneath it (which authorises actions on the individual parameters, e.g. `ssm:GetParameter`).

Destroying the link revokes the grant and removes the environment variable; adding or removing individual parameters beneath the path requires no change to the link or the role.

The `aws.lambda.ssm.populateEnvVars` annotation is shared with the per-parameter link: disabling it turns off environment variable population for every linked SSM resource on the function. Use the per-target `aws.lambda.ssm.<targetParameterPath>.populateEnvVars` annotation to control a single target.

The function's execution role must be defined as a resource in the same blueprint; the link adds the access permission to it.

If the function runs inside a VPC without internet access, the link also activates an SSM interface VPC endpoint on the linked flex VPC so the function can reach Parameter Store over the private network.

`SecureString` parameters beneath the path encrypted with a customer managed KMS key additionally require the execution role to hold `kms:Decrypt` on that key; link the function to the `aws/kms/key` as well (or grant it via the key policy). Parameters using the default `alias/aws/ssm` key need no additional grant.

### Example

```blueprintlang
version "2025-11-02"

resource apiFunction: aws/lambda/function {
    metadata {
        labels = {
            configStore = "app-config"
        }
        annotations = {
            "aws.lambda.ssm.appConfig.envVarName" = "APP_CONFIG_STORE_PATH"
            "aws.lambda.ssm.appConfig.accessLevel" = "read"
        }
    }

    select by label {
        configStore = "app-config"
    }

    spec {
        functionName = "api"
        role = apiFunctionRole.spec.arn
        # ... other function configuration
    }
}

resource appConfig: aws/ssm/parameterPath {
    metadata {
        labels = {
            configStore = "app-config"
        }
    }

    spec {
        path = "/my-app/config"
    }
}

resource dbHost: aws/ssm/parameter {
    spec {
        name = "/my-app/config/db-host"
        type = "String"
        value = "db.internal.example.com"
    }
}

resource dbPassword: aws/ssm/parameter {
    spec {
        name = "/my-app/config/db-password"
        type = "SecureString"
        secureValue = "${variables.dbPassword}"
    }
}

resource apiFunctionRole: aws/iam/role {
    spec {
        name = "api-role"
        # ... role configuration
    }
}
```
