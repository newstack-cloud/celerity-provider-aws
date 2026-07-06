Look up an existing AWS ElastiCache ReplicationGroup by replicationGroupId and export its atRestEncryptionEnabled.

```blueprintlang
version "2025-11-02"

data exampleReplicationGroup: aws/elasticache/replicationGroup {
    filter "replicationGroupId" == "example-replicationgroupid"

    export atRestEncryptionEnabled: boolean
    export autoMinorVersionUpgrade: boolean
    export automaticFailoverEnabled: boolean
}

export exampleReplicationGroupAtRestEncryptionEnabled: boolean {
    field = datasources.exampleReplicationGroup.atRestEncryptionEnabled
}
```

```yaml
version: 2025-11-02

datasources:
  exampleReplicationGroup:
    type: aws/elasticache/replicationGroup
    filter:
      field: replicationGroupId
      operator: "="
      search: example-replicationgroupid
    exports:
      atRestEncryptionEnabled:
        type: boolean
      autoMinorVersionUpgrade:
        type: boolean
      automaticFailoverEnabled:
        type: boolean

exports:
  exampleReplicationGroupAtRestEncryptionEnabled:
    type: boolean
    field: datasources.exampleReplicationGroup.atRestEncryptionEnabled
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleReplicationGroup": {
      "type": "aws/elasticache/replicationGroup",
      "filter": { "field": "replicationGroupId", "operator": "=", "search": "example-replicationgroupid" },
      "exports": {
        "atRestEncryptionEnabled": { "type": "boolean" },
        "autoMinorVersionUpgrade": { "type": "boolean" },
        "automaticFailoverEnabled": { "type": "boolean" }
      }
    }
  }
}
```
