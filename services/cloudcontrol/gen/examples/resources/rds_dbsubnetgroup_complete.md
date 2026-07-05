A AWS RDS DBSubnetGroup configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource dBSubnetGroup: aws/rds/dbSubnetGroup {
    metadata {
        displayName = "AWS RDS DBSubnetGroup complete"
    }
    spec {
        dbSubnetGroupDescription = "example-db-subnet-group-description"
        dbSubnetGroupName = "example-db-subnet-group-name"
        subnetIds = [
            "example-subnet-id"
        ]
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
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
            displayName: AWS RDS DBSubnetGroup complete
        spec:
            dbSubnetGroupDescription: example-db-subnet-group-description
            dbSubnetGroupName: example-db-subnet-group-name
            subnetIds:
                - example-subnet-id
            tags:
                - key: example-key
                  value: example-value
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "dBSubnetGroup": {
      "type": "aws/rds/dbSubnetGroup",
      "metadata": {
        "displayName": "AWS RDS DBSubnetGroup complete"
      },
      "spec": {
        "dbSubnetGroupDescription": "example-db-subnet-group-description",
        "dbSubnetGroupName": "example-db-subnet-group-name",
        "subnetIds": [
          "example-subnet-id"
        ],
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ]
      }
    }
  }
}
```
