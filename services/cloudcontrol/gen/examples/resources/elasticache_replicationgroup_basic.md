A basic AWS ElastiCache ReplicationGroup with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource replicationGroup: aws/elasticache/replicationGroup {
    metadata {
        displayName = "AWS ElastiCache ReplicationGroup basic"
    }
    spec {
        replicationGroupDescription = "example-replication-group-description"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    replicationGroup:
        type: aws/elasticache/replicationGroup
        metadata:
            displayName: AWS ElastiCache ReplicationGroup basic
        spec:
            replicationGroupDescription: example-replication-group-description
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "replicationGroup": {
      "type": "aws/elasticache/replicationGroup",
      "metadata": {
        "displayName": "AWS ElastiCache ReplicationGroup basic"
      },
      "spec": {
        "replicationGroupDescription": "example-replication-group-description"
      }
    }
  }
}
```
