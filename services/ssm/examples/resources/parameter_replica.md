A parameter replicated to another region by declaring a copy with a target region, since Parameter Store has no built-in cross-region replication.

```blueprintlang
version "2025-11-02"

resource dbHost: aws/ssm/parameter {
    spec {
        name = "/my-app/prod/db-host"
        type = "String"
        value = "db.internal.example.com"
    }
}

resource dbHostReplicaEuWest1: aws/ssm/parameter {
    spec {
        name = "/my-app/prod/db-host"
        type = "String"
        value = "db.internal.example.com"
        region = "eu-west-1"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    dbHost:
        type: aws/ssm/parameter
        spec:
            name: /my-app/prod/db-host
            type: String
            value: db.internal.example.com
    dbHostReplicaEuWest1:
        type: aws/ssm/parameter
        spec:
            name: /my-app/prod/db-host
            type: String
            value: db.internal.example.com
            region: eu-west-1
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "dbHost": {
      "type": "aws/ssm/parameter",
      "spec": {
        "name": "/my-app/prod/db-host",
        "type": "String",
        "value": "db.internal.example.com"
      }
    },
    "dbHostReplicaEuWest1": {
      "type": "aws/ssm/parameter",
      "spec": {
        "name": "/my-app/prod/db-host",
        "type": "String",
        "value": "db.internal.example.com",
        "region": "eu-west-1"
      }
    }
  }
}
```
