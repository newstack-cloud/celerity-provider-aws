A AWS RDS DBProxyTargetGroup configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource dBProxyTargetGroup: aws/rds/dbProxyTargetGroup {
    metadata {
        displayName = "AWS RDS DBProxyTargetGroup complete"
    }
    spec {
        connectionPoolConfigurationInfo = {
            connectionBorrowTimeout = 1,
            initQuery = "example-init-query",
            maxConnectionsPercent = 0,
            maxIdleConnectionsPercent = 0,
            sessionPinningFilters = [
                "example-session-pinning-filter"
            ]
        }
        dbClusterIdentifiers = [
            "example-db-cluster-identifier"
        ]
        dbInstanceIdentifiers = [
            "example-db-instance-identifier"
        ]
        dbProxyName = "example-db-proxy-name"
        targetGroupName = "default"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    dBProxyTargetGroup:
        type: aws/rds/dbProxyTargetGroup
        metadata:
            displayName: AWS RDS DBProxyTargetGroup complete
        spec:
            connectionPoolConfigurationInfo:
                connectionBorrowTimeout: 1
                initQuery: example-init-query
                maxConnectionsPercent: 0
                maxIdleConnectionsPercent: 0
                sessionPinningFilters:
                    - example-session-pinning-filter
            dbClusterIdentifiers:
                - example-db-cluster-identifier
            dbInstanceIdentifiers:
                - example-db-instance-identifier
            dbProxyName: example-db-proxy-name
            targetGroupName: default
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "dBProxyTargetGroup": {
      "type": "aws/rds/dbProxyTargetGroup",
      "metadata": {
        "displayName": "AWS RDS DBProxyTargetGroup complete"
      },
      "spec": {
        "connectionPoolConfigurationInfo": {
          "connectionBorrowTimeout": 1,
          "initQuery": "example-init-query",
          "maxConnectionsPercent": 0,
          "maxIdleConnectionsPercent": 0,
          "sessionPinningFilters": [
            "example-session-pinning-filter"
          ]
        },
        "dbClusterIdentifiers": [
          "example-db-cluster-identifier"
        ],
        "dbInstanceIdentifiers": [
          "example-db-instance-identifier"
        ],
        "dbProxyName": "example-db-proxy-name",
        "targetGroupName": "default"
      }
    }
  }
}
```
