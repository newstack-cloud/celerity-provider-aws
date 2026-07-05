A basic AWS RDS DBSubnetGroup with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource dBSubnetGroup: aws/rds/dbSubnetGroup {
    metadata {
        displayName = "AWS RDS DBSubnetGroup basic"
    }
    spec {
        dbSubnetGroupDescription = "example-db-subnet-group-description"
        subnetIds = [
            "example-subnet-id"
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    dBSubnetGroup:
        type: aws/rds/dbSubnetGroup
        metadata:
            displayName: AWS RDS DBSubnetGroup basic
        spec:
            dbSubnetGroupDescription: example-db-subnet-group-description
            subnetIds:
                - example-subnet-id
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "dBSubnetGroup": {
      "type": "aws/rds/dbSubnetGroup",
      "metadata": {
        "displayName": "AWS RDS DBSubnetGroup basic"
      },
      "spec": {
        "dbSubnetGroupDescription": "example-db-subnet-group-description",
        "subnetIds": [
          "example-subnet-id"
        ]
      }
    }
  }
}
```
