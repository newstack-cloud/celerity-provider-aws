Create a complete Lambda event invoke config with retry settings plus success and failure destinations.

```blueprintlang
version "2025-11-02"

resource myEventInvokeConfig: aws/lambda/eventInvokeConfig {
    metadata {
        displayName = "My Event Invoke Config"
    }
    spec {
        functionName = "my-lambda-function"
        qualifier = "$LATEST"
        maximumRetryAttempts = 2
        maximumEventAgeInSeconds = 1800
        destinationConfig = {
            onSuccess = {
                destination = "arn:aws:sqs:us-east-1:123456789012:success-queue"
            }
            onFailure = {
                destination = "arn:aws:sqs:us-east-1:123456789012:failure-queue"
            }
        }
    }
}
```

```yaml
version: 2025-11-02

resources:
  myEventInvokeConfig:
    type: aws/lambda/eventInvokeConfig
    metadata:
      displayName: My Event Invoke Config
    spec:
      functionName: my-lambda-function
      qualifier: $LATEST
      maximumRetryAttempts: 2
      maximumEventAgeInSeconds: 1800
      destinationConfig:
        onSuccess:
          destination: arn:aws:sqs:us-east-1:123456789012:success-queue
        onFailure:
          destination: arn:aws:sqs:us-east-1:123456789012:failure-queue
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "myEventInvokeConfig": {
      "type": "aws/lambda/eventInvokeConfig",
      "metadata": {
        "displayName": "My Event Invoke Config"
      },
      "spec": {
        "functionName": "my-lambda-function",
        "qualifier": "$LATEST",
        "maximumRetryAttempts": 2,
        "maximumEventAgeInSeconds": 1800,
        "destinationConfig": {
          "onSuccess": {
            "destination": "arn:aws:sqs:us-east-1:123456789012:success-queue"
          },
          "onFailure": {
            "destination": "arn:aws:sqs:us-east-1:123456789012:failure-queue"
          }
        }
      }
    }
  }
}
```
