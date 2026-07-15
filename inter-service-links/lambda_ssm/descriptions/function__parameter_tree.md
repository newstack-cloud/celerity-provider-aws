## Lambda Function to SSM Parameter Tree

Grants a Lambda function access to every Systems Manager (SSM) parameter in a parameter tree.

An `aws/ssm/parameterTree` resource manages every parameter beneath a path prefix (such as `/my-app/config`) as a single store. When a function links to it, the function's execution role is granted access to the entire tree with a single statement, and (by default) an environment variable holding the path prefix is added to the function so your code can enumerate the store at runtime with `ssm:GetParametersByPath` (`Recursive: true`) without manual configuration or hardcoding.

The grant and environment variable are identical in shape to the `aws/lambda/function::aws/ssm/parameterPath` link's; the default environment variable name uses the same `SSM_PARAMETER_PATH_<name>` convention so runtime code consuming a path prefix works unchanged with either resource type. The difference is that the tree also owns and provisions the parameters beneath the prefix, whereas a parameter path is a pure namespace handle for parameters managed elsewhere.

The access level is controlled with the `aws.lambda.ssm.<targetParameterTree>.accessLevel` annotation:

- `read` (default) — `ssm:GetParameter`, `ssm:GetParameters`, `ssm:GetParametersByPath`
- `write` — `ssm:PutParameter`
- `readwrite` — all of the above

Grant `write` or `readwrite` when the function (or tooling running with its role) updates stored values at runtime; the tree's values are treated as an opaque blob for drift purposes, so runtime writes are never reverted by a redeploy.

Each statement grants access to two ARNs: the path itself (which authorises `ssm:GetParametersByPath` over the hierarchy) and everything beneath it (which authorises actions on the individual parameters, e.g. `ssm:GetParameter`).

Destroying the link revokes the grant and removes the environment variable; adding or removing entries in the tree requires no change to the link or the role.

The `aws.lambda.ssm.populateEnvVars` annotation is shared with the other SSM links: disabling it turns off environment variable population for every linked SSM resource on the function. Use the per-target `aws.lambda.ssm.<targetParameterTree>.populateEnvVars` annotation to control a single target.

The function's execution role must be defined as a resource in the same blueprint; the link adds the access permission to it.

If the function runs inside a VPC without internet access, the link also activates an SSM interface VPC endpoint on the linked flex VPC so the function can reach Parameter Store over the private network.

Entries in the tree's `secureValues` encrypted with a customer managed KMS key additionally require the execution role to hold `kms:Decrypt` on that key; link the function to the `aws/kms/key` as well (or grant it via the key policy). Trees using the default `alias/aws/ssm` key need no additional grant.

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

resource appConfig: aws/ssm/parameterTree {
    metadata {
        labels = {
            configStore = "app-config"
        }
    }

    spec {
        path = "/my-app/config"
        values = {
            "db/host" = "db.internal.example.com"
            logLevel = "info"
        }
        secureValues = {
            "db/password" = "${variables.dbPassword}"
        }
    }
}

resource apiFunctionRole: aws/iam/role {
    spec {
        name = "api-role"
        # ... role configuration
    }
}
```
