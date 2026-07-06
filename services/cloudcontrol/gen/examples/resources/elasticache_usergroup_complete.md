A AWS ElastiCache UserGroup configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource userGroup: aws/elasticache/userGroup {
    metadata {
        displayName = "AWS ElastiCache UserGroup complete"
    }
    spec {
        engine = "redis"
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        userGroupId = "example-user-group-id"
        userIds = [
            "example-user-id"
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    userGroup:
        type: aws/elasticache/userGroup
        metadata:
            displayName: AWS ElastiCache UserGroup complete
        spec:
            engine: redis
            tags:
                - key: example-key
                  value: example-value
            userGroupId: example-user-group-id
            userIds:
                - example-user-id
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "userGroup": {
      "type": "aws/elasticache/userGroup",
      "metadata": {
        "displayName": "AWS ElastiCache UserGroup complete"
      },
      "spec": {
        "engine": "redis",
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "userGroupId": "example-user-group-id",
        "userIds": [
          "example-user-id"
        ]
      }
    }
  }
}
```
