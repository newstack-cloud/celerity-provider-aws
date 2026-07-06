A basic AWS RDS DBCluster with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource dBCluster: aws/rds/dbCluster {
    metadata {
        displayName = "AWS RDS DBCluster basic"
    }
    spec {
        dbClusterParameterGroupName = "example-db-cluster-parameter-group-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    dBCluster:
        type: aws/rds/dbCluster
        metadata:
            displayName: AWS RDS DBCluster basic
        spec:
            dbClusterParameterGroupName: example-db-cluster-parameter-group-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "dBCluster": {
      "type": "aws/rds/dbCluster",
      "metadata": {
        "displayName": "AWS RDS DBCluster basic"
      },
      "spec": {
        "dbClusterParameterGroupName": "example-db-cluster-parameter-group-name"
      }
    }
  }
}
```
