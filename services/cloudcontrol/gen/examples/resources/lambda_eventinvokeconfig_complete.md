A AWS Lambda EventInvokeConfig configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource eventInvokeConfig: aws/lambda/eventInvokeConfig {
    metadata {
        displayName = "AWS Lambda EventInvokeConfig complete"
    }
    spec {
        destinationConfig = {
            onFailure = {
                destination = "example-destination"
            },
            onSuccess = {
                destination = "example-destination"
            }
        }
        functionName = "example-function-name"
        maximumEventAgeInSeconds = 60
        maximumRetryAttempts = 0
        qualifier = "example-qualifier"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    eventInvokeConfig:
        type: aws/lambda/eventInvokeConfig
        metadata:
            displayName: AWS Lambda EventInvokeConfig complete
        spec:
            destinationConfig:
                onFailure:
                    destination: example-destination
                onSuccess:
                    destination: example-destination
            functionName: example-function-name
            maximumEventAgeInSeconds: 60
            maximumRetryAttempts: 0
            qualifier: example-qualifier
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "eventInvokeConfig": {
      "type": "aws/lambda/eventInvokeConfig",
      "metadata": {
        "displayName": "AWS Lambda EventInvokeConfig complete"
      },
      "spec": {
        "destinationConfig": {
          "onFailure": {
            "destination": "example-destination"
          },
          "onSuccess": {
            "destination": "example-destination"
          }
        },
        "functionName": "example-function-name",
        "maximumEventAgeInSeconds": 60,
        "maximumRetryAttempts": 0,
        "qualifier": "example-qualifier"
      }
    }
  }
}
```
