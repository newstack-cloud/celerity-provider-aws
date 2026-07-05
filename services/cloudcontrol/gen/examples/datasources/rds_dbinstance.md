Look up an existing AWS RDS DBInstance by dbInstanceIdentifier and export its allocatedStorage.

```blueprintlang
version "2025-11-02"

data exampleDbInstance: aws/rds/dbInstance {
    filter "dbInstanceIdentifier" == "example-dbinstanceidentifier"

    export allocatedStorage: string
    export autoMinorVersionUpgrade: boolean
    export automaticBackupReplicationRegion: string
}

export exampleDbInstanceAllocatedStorage: string {
    field = datasources.exampleDbInstance.allocatedStorage
}
```

```yaml
version: 2025-11-02

datasources:
  exampleDbInstance:
    type: aws/rds/dbInstance
    filter:
      field: dbInstanceIdentifier
      operator: "="
      search: example-dbinstanceidentifier
    exports:
      allocatedStorage:
        type: string
      autoMinorVersionUpgrade:
        type: boolean
      automaticBackupReplicationRegion:
        type: string

exports:
  exampleDbInstanceAllocatedStorage:
    type: string
    field: datasources.exampleDbInstance.allocatedStorage
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleDbInstance": {
      "type": "aws/rds/dbInstance",
      "filter": { "field": "dbInstanceIdentifier", "operator": "=", "search": "example-dbinstanceidentifier" },
      "exports": {
        "allocatedStorage": { "type": "string" },
        "autoMinorVersionUpgrade": { "type": "boolean" },
        "automaticBackupReplicationRegion": { "type": "string" }
      }
    }
  }
}
```
