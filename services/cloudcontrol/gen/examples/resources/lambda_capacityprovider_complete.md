A AWS Lambda CapacityProvider configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource capacityProvider: aws/lambda/capacityProvider {
    metadata {
        displayName = "AWS Lambda CapacityProvider complete"
    }
    spec {
        capacityProviderName = "example-capacity-provider-name"
        capacityProviderScalingConfig = {
            maxVCpuCount = 2,
            scalingMode = "Auto",
            scalingPolicies = [
                {
                    predefinedMetricType = "LambdaCapacityProviderAverageCPUUtilization",
                    targetValue = 0
                }
            ]
        }
        instanceRequirements = {
            allowedInstanceTypes = [
                "example-allowed-instance-type"
            ],
            architectures = [
                "x86_64"
            ],
            excludedInstanceTypes = [
                "example-excluded-instance-type"
            ]
        }
        kmsKeyArn = "example-kms-key-arn"
        permissionsConfig = {
            capacityProviderOperatorRoleArn = "example-capacity-provider-operator-role-arn"
        }
        propagateTags = {
            explicitTags = [
                {
                    key = "example-key",
                    value = "example-value"
                }
            ],
            mode = "None"
        }
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        telemetryConfig = {
            loggingConfig = {
                logGroup = "example-log-group",
                systemLogLevel = "DEBUG"
            }
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
            displayName: AWS Lambda CapacityProvider complete
        spec:
            capacityProviderName: example-capacity-provider-name
            capacityProviderScalingConfig:
                maxVCpuCount: 2
                scalingMode: Auto
                scalingPolicies:
                    - predefinedMetricType: LambdaCapacityProviderAverageCPUUtilization
                      targetValue: 0
            instanceRequirements:
                allowedInstanceTypes:
                    - example-allowed-instance-type
                architectures:
                    - x86_64
                excludedInstanceTypes:
                    - example-excluded-instance-type
            kmsKeyArn: example-kms-key-arn
            permissionsConfig:
                capacityProviderOperatorRoleArn: example-capacity-provider-operator-role-arn
            propagateTags:
                explicitTags:
                    - key: example-key
                      value: example-value
                mode: None
            tags:
                - key: example-key
                  value: example-value
            telemetryConfig:
                loggingConfig:
                    logGroup: example-log-group
                    systemLogLevel: DEBUG
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
        "displayName": "AWS Lambda CapacityProvider complete"
      },
      "spec": {
        "capacityProviderName": "example-capacity-provider-name",
        "capacityProviderScalingConfig": {
          "maxVCpuCount": 2,
          "scalingMode": "Auto",
          "scalingPolicies": [
            {
              "predefinedMetricType": "LambdaCapacityProviderAverageCPUUtilization",
              "targetValue": 0
            }
          ]
        },
        "instanceRequirements": {
          "allowedInstanceTypes": [
            "example-allowed-instance-type"
          ],
          "architectures": [
            "x86_64"
          ],
          "excludedInstanceTypes": [
            "example-excluded-instance-type"
          ]
        },
        "kmsKeyArn": "example-kms-key-arn",
        "permissionsConfig": {
          "capacityProviderOperatorRoleArn": "example-capacity-provider-operator-role-arn"
        },
        "propagateTags": {
          "explicitTags": [
            {
              "key": "example-key",
              "value": "example-value"
            }
          ],
          "mode": "None"
        },
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "telemetryConfig": {
          "loggingConfig": {
            "logGroup": "example-log-group",
            "systemLogLevel": "DEBUG"
          }
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
