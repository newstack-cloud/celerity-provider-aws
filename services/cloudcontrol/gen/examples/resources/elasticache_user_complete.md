A AWS ElastiCache User configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource user: aws/elasticache/user {
    metadata {
        displayName = "AWS ElastiCache User complete"
    }
    spec {
        accessString = "example-access-string"
        authenticationMode = {
            passwords = [
                "example-password"
            ],
            type = "password"
        }
        engine = "redis"
        noPasswordRequired = false
        passwords = [
            "example-password"
        ]
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
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
            displayName: AWS ElastiCache User complete
        spec:
            accessString: example-access-string
            authenticationMode:
                passwords:
                    - example-password
                type: password
            engine: redis
            noPasswordRequired: false
            passwords:
                - example-password
            tags:
                - key: example-key
                  value: example-value
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
        "displayName": "AWS ElastiCache User complete"
      },
      "spec": {
        "accessString": "example-access-string",
        "authenticationMode": {
          "passwords": [
            "example-password"
          ],
          "type": "password"
        },
        "engine": "redis",
        "noPasswordRequired": false,
        "passwords": [
          "example-password"
        ],
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "userId": "example-user-id",
        "userName": "example-user-name"
      }
    }
  }
}
```
