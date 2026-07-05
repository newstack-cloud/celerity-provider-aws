A plaintext String parameter holding a non-secret configuration value.

```blueprintlang
version "2025-11-02"

resource dbHost: aws/ssm/parameter {
    metadata {
        displayName = "Database host parameter"
    }
    spec {
        name = "/my-app/prod/db-host"
        type = "String"
        value = "db.internal.example.com"
        description = "The database host for the production environment"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    dbHost:
        type: aws/ssm/parameter
        metadata:
            displayName: Database host parameter
        spec:
            name: /my-app/prod/db-host
            type: String
            value: db.internal.example.com
            description: The database host for the production environment
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "dbHost": {
      "type": "aws/ssm/parameter",
      "metadata": {
        "displayName": "Database host parameter"
      },
      "spec": {
        "name": "/my-app/prod/db-host",
        "type": "String",
        "value": "db.internal.example.com",
        "description": "The database host for the production environment"
      }
    }
  }
}
```
