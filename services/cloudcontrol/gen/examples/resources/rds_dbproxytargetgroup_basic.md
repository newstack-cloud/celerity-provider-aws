A basic AWS RDS DBProxyTargetGroup with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource dBProxyTargetGroup: aws/rds/dbProxyTargetGroup {
    metadata {
        displayName = "AWS RDS DBProxyTargetGroup basic"
    }
    spec {
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
            displayName: AWS RDS DBProxyTargetGroup basic
        spec:
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
        "displayName": "AWS RDS DBProxyTargetGroup basic"
      },
      "spec": {
        "dbProxyName": "example-db-proxy-name",
        "targetGroupName": "default"
      }
    }
  }
}
```
