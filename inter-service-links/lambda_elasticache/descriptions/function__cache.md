## Lambda Function to ElastiCache Replication Group

Grants a Lambda function network access to an ElastiCache (Redis/Valkey) replication group.

When a function links to a cache, the link:

1. **Opens a security-group rule** so the function can reach the cache on the cache port. The function's security group is authorised (egress) to reach the cache's security group (ingress) on the port set by `aws.lambda.elasticache.<targetCache>.port` (default `6379`). When the function and cache share the same VPC security group, this is a self-referencing rule. This is a no-op when the function is not attached to a VPC. When the function is VPC-attached, this network configuration is surfaced in staged changes as a best-effort known-on-deploy signal (`<function>NetworkAccess`).

2. **Grants IAM authentication** when `aws.lambda.elasticache.<targetCache>.authMode` is `iam`: an inline policy on the function's execution role allows `elasticache:Connect`, scoped to **both** the replication group and the IAM-enabled ElastiCache user identified by `aws.lambda.elasticache.<targetCache>.userId` (default `*`, any user in the cache's user group). The runtime SDK generates short-lived SigV4 auth tokens instead of a static AUTH token. When `authMode` is `password` (default), no IAM grant is added; the function obtains the AUTH token from a Secrets Manager secret it is linked to separately (via the `aws/lambda/function` → `aws/secretsmanager/secret` link).

3. **Populates connection environment variables** (unless disabled): `<PREFIX>_HOST` which holds the configuration endpoint in cluster mode, otherwise the primary endpoint and `<PREFIX>_PORT`. The prefix defaults to an auto-generated value based on the cache name and can be set with `aws.lambda.elasticache.<targetCache>.envVarPrefix`.

The replication group exposes no ARN attribute, so the `elasticache:Connect` resource ARNs (`arn:aws:elasticache:<region>:<account>:replicationgroup:<id>` and `...:user:<userId>`) are constructed from the region, account (from the function ARN) and the replication group / user ids. The function's execution role must be defined as a resource in the same blueprint (used only when `authMode` is `iam`).

Placing the cache in the VPC (its subnet group and security-group membership) is configured on the replication group resource at creation time via references, ElastiCache subnet placement is fixed at creation and cannot be changed by a link.

### Example

```blueprintlang
version "2025-11-02"

resource apiFunction: aws/lambda/function {
    metadata {
        labels = {
            cache = "sessions"
        }
        annotations = {
            "aws.lambda.elasticache.sessionCache.envVarPrefix" = "SESSION_CACHE"
        }
    }

    select by label {
        cache = "sessions"
    }

    spec {
        functionName = "api"
        # ... other function configuration (VPC config placed by the flex VPC link)
    }
}

resource sessionCache: aws/elasticache/replicationGroup {
    metadata {
        labels = {
            cache = "sessions"
        }
    }

    spec {
        replicationGroupId = "sessions"
        replicationGroupDescription = "Session cache"
        engine = "valkey"
        cacheNodeType = "cache.t4g.micro"
        # ... cacheSubnetGroupName, securityGroupIds, port
    }
}
```
