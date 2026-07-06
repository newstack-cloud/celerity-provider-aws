A basic AWS ElastiCache User with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource user: aws/elasticache/user {
    metadata {
        displayName = "AWS ElastiCache User basic"
    }
    spec {
        engine = "redis"
        userId = "example-user-id"
        userName = "example-user-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    user:
        type: aws/elasticache/user
        metadata:
            displayName: AWS ElastiCache User basic
        spec:
            engine: redis
            userId: example-user-id
            userName: example-user-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "user": {
      "type": "aws/elasticache/user",
      "metadata": {
        "displayName": "AWS ElastiCache User basic"
      },
      "spec": {
        "engine": "redis",
        "userId": "example-user-id",
        "userName": "example-user-name"
      }
    }
  }
}
```
