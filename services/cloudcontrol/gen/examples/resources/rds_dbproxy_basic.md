A basic AWS RDS DBProxy with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource dBProxy: aws/rds/dbProxy {
    metadata {
        displayName = "AWS RDS DBProxy basic"
    }
    spec {
        dbProxyName = "example-db-proxy-name"
        engineFamily = "MYSQL"
        roleArn = "example-role-arn"
        vpcSubnetIds = [
            "example-vpc-subnet-id"
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    dBProxy:
        type: aws/rds/dbProxy
        metadata:
            displayName: AWS RDS DBProxy basic
        spec:
            dbProxyName: example-db-proxy-name
            engineFamily: MYSQL
            roleArn: example-role-arn
            vpcSubnetIds:
                - example-vpc-subnet-id
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "dBProxy": {
      "type": "aws/rds/dbProxy",
      "metadata": {
        "displayName": "AWS RDS DBProxy basic"
      },
      "spec": {
        "dbProxyName": "example-db-proxy-name",
        "engineFamily": "MYSQL",
        "roleArn": "example-role-arn",
        "vpcSubnetIds": [
          "example-vpc-subnet-id"
        ]
      }
    }
  }
}
```
