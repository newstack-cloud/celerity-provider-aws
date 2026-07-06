A basic AWS ElastiCache UserGroup with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource userGroup: aws/elasticache/userGroup {
    metadata {
        displayName = "AWS ElastiCache UserGroup basic"
    }
    spec {
        engine = "redis"
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
            displayName: AWS ElastiCache UserGroup basic
        spec:
            engine: redis
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
        "displayName": "AWS ElastiCache UserGroup basic"
      },
      "spec": {
        "engine": "redis",
        "userGroupId": "example-user-group-id",
        "userIds": [
          "example-user-id"
        ]
      }
    }
  }
}
```
