A basic AWS Lambda MicrovmImage with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource microvmImage: aws/lambda/microvmImage {
    metadata {
        displayName = "AWS Lambda MicrovmImage basic"
    }
    spec {
        additionalOsCapabilities = [
            "ALL"
        ]
        baseImageArn = "example-base-image-arn"
        baseImageVersion = "example-base-image-version"
        buildRoleArn = "example-build-role-arn"
        codeArtifact = {
            uri = "example-uri"
        }
        cpuConfigurations = [
            {
                architecture = "ARM_64"
            }
        ]
        description = "example-description"
        egressNetworkConnectors = [
            "example-egress-network-connector"
        ]
        environmentVariables = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        hooks = {
            microvmHooks = {
                resume = "DISABLED",
                resumeTimeoutInSeconds = 1,
                run = "DISABLED",
                runTimeoutInSeconds = 1,
                suspend = "DISABLED",
                suspendTimeoutInSeconds = 1,
                terminate = "DISABLED",
                terminateTimeoutInSeconds = 1
            },
            microvmImageHooks = {
                ready = "DISABLED",
                readyTimeoutInSeconds = 1,
                validate = "DISABLED",
                validateTimeoutInSeconds = 1
            },
            port = 1
        }
        logging = {
            cloudWatch = {
                logGroup = "example-log-group",
                logStream = "example-log-stream"
            },
            disabled = false
        }
        name = "example-name"
        resources = [
            {
                minimumMemoryInMiB = 1
            }
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    microvmImage:
        type: aws/lambda/microvmImage
        metadata:
            displayName: AWS Lambda MicrovmImage basic
        spec:
            additionalOsCapabilities:
                - ALL
            baseImageArn: example-base-image-arn
            baseImageVersion: example-base-image-version
            buildRoleArn: example-build-role-arn
            codeArtifact:
                uri: example-uri
            cpuConfigurations:
                - architecture: ARM_64
            description: example-description
            egressNetworkConnectors:
                - example-egress-network-connector
            environmentVariables:
                - key: example-key
                  value: example-value
            hooks:
                microvmHooks:
                    resume: DISABLED
                    resumeTimeoutInSeconds: 1
                    run: DISABLED
                    runTimeoutInSeconds: 1
                    suspend: DISABLED
                    suspendTimeoutInSeconds: 1
                    terminate: DISABLED
                    terminateTimeoutInSeconds: 1
                microvmImageHooks:
                    ready: DISABLED
                    readyTimeoutInSeconds: 1
                    validate: DISABLED
                    validateTimeoutInSeconds: 1
                port: 1
            logging:
                cloudWatch:
                    logGroup: example-log-group
                    logStream: example-log-stream
                disabled: false
            name: example-name
            resources:
                - minimumMemoryInMiB: 1
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "microvmImage": {
      "type": "aws/lambda/microvmImage",
      "metadata": {
        "displayName": "AWS Lambda MicrovmImage basic"
      },
      "spec": {
        "additionalOsCapabilities": [
          "ALL"
        ],
        "baseImageArn": "example-base-image-arn",
        "baseImageVersion": "example-base-image-version",
        "buildRoleArn": "example-build-role-arn",
        "codeArtifact": {
          "uri": "example-uri"
        },
        "cpuConfigurations": [
          {
            "architecture": "ARM_64"
          }
        ],
        "description": "example-description",
        "egressNetworkConnectors": [
          "example-egress-network-connector"
        ],
        "environmentVariables": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "hooks": {
          "microvmHooks": {
            "resume": "DISABLED",
            "resumeTimeoutInSeconds": 1,
            "run": "DISABLED",
            "runTimeoutInSeconds": 1,
            "suspend": "DISABLED",
            "suspendTimeoutInSeconds": 1,
            "terminate": "DISABLED",
            "terminateTimeoutInSeconds": 1
          },
          "microvmImageHooks": {
            "ready": "DISABLED",
            "readyTimeoutInSeconds": 1,
            "validate": "DISABLED",
            "validateTimeoutInSeconds": 1
          },
          "port": 1
        },
        "logging": {
          "cloudWatch": {
            "logGroup": "example-log-group",
            "logStream": "example-log-stream"
          },
          "disabled": false
        },
        "name": "example-name",
        "resources": [
          {
            "minimumMemoryInMiB": 1
          }
        ]
      }
    }
  }
}
```
