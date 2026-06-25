A basic AWS RDS DBInstance with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource dBInstance: aws/rds/dbInstance {
    metadata {
        displayName = "AWS RDS DBInstance basic"
    }
    spec {
        characterSetName = "example-character-set-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    dBInstance:
        type: aws/rds/dbInstance
        metadata:
            displayName: AWS RDS DBInstance basic
        spec:
            characterSetName: example-character-set-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "dBInstance": {
      "type": "aws/rds/dbInstance",
      "metadata": {
        "displayName": "AWS RDS DBInstance basic"
      },
      "spec": {
        "characterSetName": "example-character-set-name"
      }
    }
  }
}
```
