A AWS RDS DBProxy configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource dBProxy: aws/rds/dbProxy {
    metadata {
        displayName = "AWS RDS DBProxy complete"
    }
    spec {
        auth = [
            {
                authScheme = "SECRETS",
                clientPasswordAuthType = "MYSQL_NATIVE_PASSWORD",
                description = "example-description",
                iamAuth = "DISABLED",
                secretArn = "example-secret-arn"
            }
        ]
        dbProxyName = "example-db-proxy-name"
        debugLogging = false
        defaultAuthScheme = "IAM_AUTH"
        endpointNetworkType = "IPV4"
        engineFamily = "MYSQL"
        idleClientTimeout = 1
        requireTLS = false
        roleArn = "example-role-arn"
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        targetConnectionNetworkType = "IPV4"
        vpcSecurityGroupIds = [
            "example-vpc-security-group-id"
        ]
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
            displayName: AWS RDS DBProxy complete
        spec:
            auth:
                - authScheme: SECRETS
                  clientPasswordAuthType: MYSQL_NATIVE_PASSWORD
                  description: example-description
                  iamAuth: DISABLED
                  secretArn: example-secret-arn
            dbProxyName: example-db-proxy-name
            debugLogging: false
            defaultAuthScheme: IAM_AUTH
            endpointNetworkType: IPV4
            engineFamily: MYSQL
            idleClientTimeout: 1
            requireTLS: false
            roleArn: example-role-arn
            tags:
                - key: example-key
                  value: example-value
            targetConnectionNetworkType: IPV4
            vpcSecurityGroupIds:
                - example-vpc-security-group-id
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
        "displayName": "AWS RDS DBProxy complete"
      },
      "spec": {
        "auth": [
          {
            "authScheme": "SECRETS",
            "clientPasswordAuthType": "MYSQL_NATIVE_PASSWORD",
            "description": "example-description",
            "iamAuth": "DISABLED",
            "secretArn": "example-secret-arn"
          }
        ],
        "dbProxyName": "example-db-proxy-name",
        "debugLogging": false,
        "defaultAuthScheme": "IAM_AUTH",
        "endpointNetworkType": "IPV4",
        "engineFamily": "MYSQL",
        "idleClientTimeout": 1,
        "requireTLS": false,
        "roleArn": "example-role-arn",
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "targetConnectionNetworkType": "IPV4",
        "vpcSecurityGroupIds": [
          "example-vpc-security-group-id"
        ],
        "vpcSubnetIds": [
          "example-vpc-subnet-id"
        ]
      }
    }
  }
}
```
