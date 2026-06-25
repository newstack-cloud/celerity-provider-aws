A AWS Lambda Function configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource function: aws/lambda/function {
    metadata {
        displayName = "AWS Lambda Function complete"
    }
    spec {
        architectures = [
            "x86_64"
        ]
        capacityProviderConfig = {
            lambdaManagedInstancesCapacityProviderConfig = {
                capacityProviderArn = "example-capacity-provider-arn",
                executionEnvironmentMemoryGiBPerVCpu = 2,
                perExecutionEnvironmentMaxConcurrency = 1
            }
        }
        code = {
            imageUri = "example-image-uri",
            s3Bucket = "example-s3-bucket",
            s3Key = "example-s3-key",
            s3ObjectVersion = "example-s3-object-version",
            sourceKMSKeyArn = "example-source-k-m-s-key-arn",
            zipFile = "example-zip-file"
        }
        codeSigningConfigArn = "example-code-signing-config-arn"
        deadLetterConfig = {
            targetArn = "example-target-arn"
        }
        description = "example-description"
        durableConfig = {
            executionTimeout = 1,
            retentionPeriodInDays = 1
        }
        environment = {
            variables = {
                example = "example-variables"
            }
        }
        ephemeralStorage = {
            size = 512
        }
        fileSystemConfigs = [
            {
                arn = "example-arn",
                localMountPath = "example-local-mount-path"
            }
        ]
        functionName = "example-function-name"
        functionScalingConfig = {
            maxExecutionEnvironments = 0,
            minExecutionEnvironments = 0
        }
        handler = "example-handler"
        imageConfig = {
            command = [
                "example-command"
            ],
            entryPoint = [
                "example-entry-point"
            ],
            workingDirectory = "example-working-directory"
        }
        kmsKeyArn = "example-kms-key-arn"
        layers = [
            "example-layer"
        ]
        loggingConfig = {
            applicationLogLevel = "TRACE",
            logFormat = "Text",
            logGroup = "example-log-group",
            systemLogLevel = "DEBUG"
        }
        memorySize = 1
        packageType = "Image"
        publishToLatestPublished = false
        recursiveLoop = "Allow"
        reservedConcurrentExecutions = 0
        role = "example-role"
        runtime = "example-runtime"
        runtimeManagementConfig = {
            runtimeVersionArn = "example-runtime-version-arn",
            updateRuntimeOn = "Auto"
        }
        snapStart = {
            applyOn = "PublishedVersions"
        }
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        tenancyConfig = {
            tenantIsolationMode = "PER_TENANT"
        }
        timeout = 1
        tracingConfig = {
            mode = "Active"
        }
        vpcConfig = {
            ipv6AllowedForDualStack = false,
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
    function:
        type: aws/lambda/function
        metadata:
            displayName: AWS Lambda Function complete
        spec:
            architectures:
                - x86_64
            capacityProviderConfig:
                lambdaManagedInstancesCapacityProviderConfig:
                    capacityProviderArn: example-capacity-provider-arn
                    executionEnvironmentMemoryGiBPerVCpu: 2
                    perExecutionEnvironmentMaxConcurrency: 1
            code:
                imageUri: example-image-uri
                s3Bucket: example-s3-bucket
                s3Key: example-s3-key
                s3ObjectVersion: example-s3-object-version
                sourceKMSKeyArn: example-source-k-m-s-key-arn
                zipFile: example-zip-file
            codeSigningConfigArn: example-code-signing-config-arn
            deadLetterConfig:
                targetArn: example-target-arn
            description: example-description
            durableConfig:
                executionTimeout: 1
                retentionPeriodInDays: 1
            environment:
                variables:
                    example: example-variables
            ephemeralStorage:
                size: 512
            fileSystemConfigs:
                - arn: example-arn
                  localMountPath: example-local-mount-path
            functionName: example-function-name
            functionScalingConfig:
                maxExecutionEnvironments: 0
                minExecutionEnvironments: 0
            handler: example-handler
            imageConfig:
                command:
                    - example-command
                entryPoint:
                    - example-entry-point
                workingDirectory: example-working-directory
            kmsKeyArn: example-kms-key-arn
            layers:
                - example-layer
            loggingConfig:
                applicationLogLevel: TRACE
                logFormat: Text
                logGroup: example-log-group
                systemLogLevel: DEBUG
            memorySize: 1
            packageType: Image
            publishToLatestPublished: false
            recursiveLoop: Allow
            reservedConcurrentExecutions: 0
            role: example-role
            runtime: example-runtime
            runtimeManagementConfig:
                runtimeVersionArn: example-runtime-version-arn
                updateRuntimeOn: Auto
            snapStart:
                applyOn: PublishedVersions
            tags:
                - key: example-key
                  value: example-value
            tenancyConfig:
                tenantIsolationMode: PER_TENANT
            timeout: 1
            tracingConfig:
                mode: Active
            vpcConfig:
                ipv6AllowedForDualStack: false
                securityGroupIds:
                    - example-security-group-id
                subnetIds:
                    - example-subnet-id
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "function": {
      "type": "aws/lambda/function",
      "metadata": {
        "displayName": "AWS Lambda Function complete"
      },
      "spec": {
        "architectures": [
          "x86_64"
        ],
        "capacityProviderConfig": {
          "lambdaManagedInstancesCapacityProviderConfig": {
            "capacityProviderArn": "example-capacity-provider-arn",
            "executionEnvironmentMemoryGiBPerVCpu": 2,
            "perExecutionEnvironmentMaxConcurrency": 1
          }
        },
        "code": {
          "imageUri": "example-image-uri",
          "s3Bucket": "example-s3-bucket",
          "s3Key": "example-s3-key",
          "s3ObjectVersion": "example-s3-object-version",
          "sourceKMSKeyArn": "example-source-k-m-s-key-arn",
          "zipFile": "example-zip-file"
        },
        "codeSigningConfigArn": "example-code-signing-config-arn",
        "deadLetterConfig": {
          "targetArn": "example-target-arn"
        },
        "description": "example-description",
        "durableConfig": {
          "executionTimeout": 1,
          "retentionPeriodInDays": 1
        },
        "environment": {
          "variables": {
            "example": "example-variables"
          }
        },
        "ephemeralStorage": {
          "size": 512
        },
        "fileSystemConfigs": [
          {
            "arn": "example-arn",
            "localMountPath": "example-local-mount-path"
          }
        ],
        "functionName": "example-function-name",
        "functionScalingConfig": {
          "maxExecutionEnvironments": 0,
          "minExecutionEnvironments": 0
        },
        "handler": "example-handler",
        "imageConfig": {
          "command": [
            "example-command"
          ],
          "entryPoint": [
            "example-entry-point"
          ],
          "workingDirectory": "example-working-directory"
        },
        "kmsKeyArn": "example-kms-key-arn",
        "layers": [
          "example-layer"
        ],
        "loggingConfig": {
          "applicationLogLevel": "TRACE",
          "logFormat": "Text",
          "logGroup": "example-log-group",
          "systemLogLevel": "DEBUG"
        },
        "memorySize": 1,
        "packageType": "Image",
        "publishToLatestPublished": false,
        "recursiveLoop": "Allow",
        "reservedConcurrentExecutions": 0,
        "role": "example-role",
        "runtime": "example-runtime",
        "runtimeManagementConfig": {
          "runtimeVersionArn": "example-runtime-version-arn",
          "updateRuntimeOn": "Auto"
        },
        "snapStart": {
          "applyOn": "PublishedVersions"
        },
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "tenancyConfig": {
          "tenantIsolationMode": "PER_TENANT"
        },
        "timeout": 1,
        "tracingConfig": {
          "mode": "Active"
        },
        "vpcConfig": {
          "ipv6AllowedForDualStack": false,
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
