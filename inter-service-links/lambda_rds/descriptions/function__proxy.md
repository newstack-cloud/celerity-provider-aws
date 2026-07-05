## Lambda Function to RDS Proxy

Grants a Lambda function network and authentication access to an RDS database through an RDS Proxy.

RDS Proxy pools database connections, which is important for Lambda functions to avoid exhausting database connections across concurrent invocations. When a function links to a proxy, the link:

1. **Opens a security-group rule** so the function can reach the proxy on the database port. The function's security group is authorised (egress) to reach the proxy's security group (ingress) on the port set by `aws.lambda.rds.<targetProxy>.port` (default `5432`). When the function and proxy share the same VPC security group, this is a self-referencing rule. This is a no-op when the function is not attached to a VPC. When the function is VPC-attached, this network configuration is surfaced in staged changes as a best-effort known-on-deploy signal (`<function>NetworkAccess`).

2. **Grants IAM authentication** when `aws.lambda.rds.<targetProxy>.authMode` is `iam`: an inline policy on the function's execution role allows `rds-db:connect`, scoped to the proxy's resource id and the `dbUser` annotation (default `*`, any IAM-enabled user). The runtime generates short-lived tokens instead of a static password. When `authMode` is `password` (default), no IAM grant is added, this will link the function to a Secrets Manager secret separately for credentials.

3. **Populates connection environment variables** (unless disabled): `<PREFIX>_HOST` (the proxy endpoint), `<PREFIX>_PORT` and `<PREFIX>_DATABASE`. The prefix defaults to an auto-generated value based on the proxy name and can be set with `aws.lambda.rds.<targetProxy>.envVarPrefix`.

The function's execution role must be defined as a resource in the same blueprint (used only when `authMode` is `iam`).

Placing the database in the VPC (its subnet group and security-group membership) is configured on the database and proxy resources at creation time via references, RDS subnet placement is fixed at creation and cannot be changed by a link.

### Example

```blueprintlang
version "2025-11-02"

resource apiFunction: aws/lambda/function {
    metadata {
        labels = {
            database = "orders"
        }
        annotations = {
            "aws.lambda.rds.ordersProxy.envVarPrefix" = "ORDERS_DB"
            "aws.lambda.rds.ordersProxy.authMode" = "iam"
            "aws.lambda.rds.ordersProxy.dbUser" = "orders_app"
            "aws.lambda.rds.ordersProxy.databaseName" = "orders"
        }
    }

    select by label {
        database = "orders"
    }

    spec {
        functionName = "api"
        role = apiFunctionRole.spec.arn
        # ... other function configuration (VPC config placed by the flex VPC link)
    }
}

resource ordersProxy: aws/rds/dbProxy {
    metadata {
        labels = {
            database = "orders"
        }
    }

    spec {
        dbProxyName = "orders-proxy"
        engineFamily = "POSTGRESQL"
        # ... auth, roleArn, vpcSubnetIds, vpcSecurityGroupIds
    }
}

resource apiFunctionRole: aws/iam/role {
    spec {
        name = "api-role"
        # ... role configuration
    }
}
```
