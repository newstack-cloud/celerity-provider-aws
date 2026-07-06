Look up an existing AWS RDS DBCluster by dbClusterIdentifier and export its allocatedStorage.

```blueprintlang
version "2025-11-02"

data exampleDbCluster: aws/rds/dbCluster {
    filter "dbClusterIdentifier" == "example-dbclusteridentifier"

    export allocatedStorage: integer
    export autoMinorVersionUpgrade: boolean
    export backtrackWindow: integer
}

export exampleDbClusterAllocatedStorage: integer {
    field = datasources.exampleDbCluster.allocatedStorage
}
```

```yaml
version: 2025-11-02

datasources:
  exampleDbCluster:
    type: aws/rds/dbCluster
    filter:
      field: dbClusterIdentifier
      operator: "="
      search: example-dbclusteridentifier
    exports:
      allocatedStorage:
        type: integer
      autoMinorVersionUpgrade:
        type: boolean
      backtrackWindow:
        type: integer

exports:
  exampleDbClusterAllocatedStorage:
    type: integer
    field: datasources.exampleDbCluster.allocatedStorage
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleDbCluster": {
      "type": "aws/rds/dbCluster",
      "filter": { "field": "dbClusterIdentifier", "operator": "=", "search": "example-dbclusteridentifier" },
      "exports": {
        "allocatedStorage": { "type": "integer" },
        "autoMinorVersionUpgrade": { "type": "boolean" },
        "backtrackWindow": { "type": "integer" }
      }
    }
  }
}
```
