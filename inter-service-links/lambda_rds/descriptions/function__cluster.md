## Lambda Function to Aurora Cluster

Grants a Lambda function network and authentication access to an Aurora database cluster (including Aurora Serverless v2).

The function connects directly to the cluster's writer endpoint. For connection pooling in front of the cluster, place an `aws/rds/dbProxy` targeting it and use the `aws/lambda/function` → `aws/rds/dbProxy` link instead. When a function links to a cluster, the link:

1. **Opens a security-group rule** so the function can reach the cluster on the database port. The function's security group is authorised (egress) to reach the cluster's security group (ingress) on the port set by `aws.lambda.rds.<targetCluster>.port` (default `5432` for Aurora PostgreSQL; set `3306` for Aurora MySQL). When the function and cluster share the same VPC security group, this is a self-referencing rule. This is a no-op when the function is not attached to a VPC. When the function is VPC-attached, this network configuration is surfaced in staged changes as a best-effort known-on-deploy signal (`<function>NetworkAccess`).

2. **Grants IAM authentication** when `aws.lambda.rds.<targetCluster>.authMode` is `iam`: an inline policy on the function's execution role allows `rds-db:connect`, scoped to the cluster's resource id (`cluster-XXXX`) and the `dbUser` annotation (default `*`, any IAM-enabled user). The runtime generates short-lived tokens instead of a static password. When `authMode` is `password` (default), no IAM grant is added; link the function to a Secrets Manager secret separately for credentials.

3. **Populates connection environment variables** (unless disabled): `<PREFIX>_HOST` (the cluster writer endpoint), `<PREFIX>_PORT`, `<PREFIX>_DATABASE`, and when `aws.lambda.rds.<targetCluster>.readerEndpoint` is enabled, `<PREFIX>_READER_HOST` (the cluster reader endpoint, for read scaling across Aurora replicas). The prefix defaults to an auto-generated value based on the cluster name and can be set with `aws.lambda.rds.<targetCluster>.envVarPrefix`.

The function's execution role must be defined as a resource in the same blueprint (used only when `authMode` is `iam`).

Placing the database in the VPC (its subnet group and security-group membership) is configured on the cluster resource at creation time via references, RDS subnet placement is fixed at creation and cannot be changed by a link.

### Example

```blueprintlang
version "2025-11-02"

resource apiFunction: aws/lambda/function {
    metadata {
        labels = {
            database = "orders"
        }
        annotations = {
            "aws.lambda.rds.ordersCluster.envVarPrefix" = "ORDERS_DB"
            "aws.lambda.rds.ordersCluster.authMode" = "iam"
            "aws.lambda.rds.ordersCluster.dbUser" = "orders_app"
            "aws.lambda.rds.ordersCluster.databaseName" = "orders"
            "aws.lambda.rds.ordersCluster.readerEndpoint" = "true"
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

resource ordersCluster: aws/rds/dbCluster {
    metadata {
        labels = {
            database = "orders"
        }
    }

    spec {
        engine = "aurora-postgresql"
        serverlessV2ScalingConfiguration = {
            minCapacity = 0.5
            maxCapacity = 4
        }
        # ... dbSubnetGroupName, vpcSecurityGroupIds
    }
}

resource apiFunctionRole: aws/iam/role {
    spec {
        name = "api-role"
        # ... role configuration
    }
}
```
