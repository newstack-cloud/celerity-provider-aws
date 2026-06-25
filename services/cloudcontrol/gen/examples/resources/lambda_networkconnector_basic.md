A basic AWS Lambda NetworkConnector with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource networkConnector: aws/lambda/networkConnector {
    metadata {
        displayName = "AWS Lambda NetworkConnector basic"
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
    }
}
```

```yaml
version: "2025-11-02"
resources:
    networkConnector:
        type: aws/lambda/networkConnector
        metadata:
            displayName: AWS Lambda NetworkConnector basic
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
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "networkConnector": {
      "type": "aws/lambda/networkConnector",
      "metadata": {
        "displayName": "AWS Lambda NetworkConnector basic"
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
        }
      }
    }
  }
}
```
