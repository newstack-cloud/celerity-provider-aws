A AWS Lambda NetworkConnector configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource networkConnector: aws/lambda/networkConnector {
    metadata {
        displayName = "AWS Lambda NetworkConnector complete"
    }
    spec {
        configuration = {
            vpcEgressConfiguration = {
                associatedComputeResourceTypes = [
                    "MicroVm"
                ],
                networkProtocol = "IPv4",
                securityGroupIds = [
                    "example-security-group-id"
                ],
                subnetIds = [
                    "example-subnet-id"
                ]
            }
        }
        name = "example-name"
        operatorRole = "example-operator-role"
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
    networkConnector:
        type: aws/lambda/networkConnector
        metadata:
            displayName: AWS Lambda NetworkConnector complete
        spec:
            configuration:
                vpcEgressConfiguration:
                    associatedComputeResourceTypes:
                        - MicroVm
                    networkProtocol: IPv4
                    securityGroupIds:
                        - example-security-group-id
                    subnetIds:
                        - example-subnet-id
            name: example-name
            operatorRole: example-operator-role
            tags:
                - key: example-key
                  value: example-value
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "networkConnector": {
      "type": "aws/lambda/networkConnector",
      "metadata": {
        "displayName": "AWS Lambda NetworkConnector complete"
      },
      "spec": {
        "configuration": {
          "vpcEgressConfiguration": {
            "associatedComputeResourceTypes": [
              "MicroVm"
            ],
            "networkProtocol": "IPv4",
            "securityGroupIds": [
              "example-security-group-id"
            ],
            "subnetIds": [
              "example-subnet-id"
            ]
          }
        },
        "name": "example-name",
        "operatorRole": "example-operator-role",
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
