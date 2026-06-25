A AWS Lambda Version configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource "version": aws/lambda/functionVersion {
    metadata {
        displayName = "AWS Lambda Version complete"
    }
    spec {
        codeSha256 = "example-code-sha256"
        description = "example-description"
        functionName = "example-function-name"
        functionScalingConfig = {
            maxExecutionEnvironments = 0,
            minExecutionEnvironments = 0
        }
        provisionedConcurrencyConfig = {
            provisionedConcurrentExecutions = 1
        }
        runtimePolicy = {
            runtimeVersionArn = "example-runtime-version-arn",
            updateRuntimeOn = "example-update-runtime-on"
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    version:
        type: aws/lambda/functionVersion
        metadata:
            displayName: AWS Lambda Version complete
        spec:
            codeSha256: example-code-sha256
            description: example-description
            functionName: example-function-name
            functionScalingConfig:
                maxExecutionEnvironments: 0
                minExecutionEnvironments: 0
            provisionedConcurrencyConfig:
                provisionedConcurrentExecutions: 1
            runtimePolicy:
                runtimeVersionArn: example-runtime-version-arn
                updateRuntimeOn: example-update-runtime-on
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "version": {
      "type": "aws/lambda/functionVersion",
      "metadata": {
        "displayName": "AWS Lambda Version complete"
      },
      "spec": {
        "codeSha256": "example-code-sha256",
        "description": "example-description",
        "functionName": "example-function-name",
        "functionScalingConfig": {
          "maxExecutionEnvironments": 0,
          "minExecutionEnvironments": 0
        },
        "provisionedConcurrencyConfig": {
          "provisionedConcurrentExecutions": 1
        },
        "runtimePolicy": {
          "runtimeVersionArn": "example-runtime-version-arn",
          "updateRuntimeOn": "example-update-runtime-on"
        }
      }
    }
  }
}
```
