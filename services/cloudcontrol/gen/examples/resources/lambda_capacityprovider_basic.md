A basic AWS Lambda CapacityProvider with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource capacityProvider: aws/lambda/capacityProvider {
    metadata {
        displayName = "AWS Lambda CapacityProvider basic"
    }
    spec {
        permissionsConfig = {
            capacityProviderOperatorRoleArn = "example-capacity-provider-operator-role-arn"
        }
        vpcConfig = {
            securityGroupIds = [
                "example-security-group-id"
            ],
            subnetIds = [
                "example-subnet-id"
            ]
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    capacityProvider:
        type: aws/lambda/capacityProvider
        metadata:
            displayName: AWS Lambda CapacityProvider basic
        spec:
            permissionsConfig:
                capacityProviderOperatorRoleArn: example-capacity-provider-operator-role-arn
            vpcConfig:
                securityGroupIds:
                    - example-security-group-id
                subnetIds:
                    - example-subnet-id
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "capacityProvider": {
      "type": "aws/lambda/capacityProvider",
      "metadata": {
        "displayName": "AWS Lambda CapacityProvider basic"
      },
      "spec": {
        "permissionsConfig": {
          "capacityProviderOperatorRoleArn": "example-capacity-provider-operator-role-arn"
        },
        "vpcConfig": {
          "securityGroupIds": [
            "example-security-group-id"
          ],
          "subnetIds": [
            "example-subnet-id"
          ]
        }
      }
    }
  }
}
```
