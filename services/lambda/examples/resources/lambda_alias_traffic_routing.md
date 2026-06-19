Create a Lambda alias with traffic routing configuration for canary deployments.

```blueprintlang
version "2025-11-02"

resource canaryAlias: aws/lambda/alias {
    metadata {
        displayName = "Canary Alias"
    }
    spec {
        functionName = "my-lambda-function"
        name = "CANARY"
        functionVersion = "2"
        description = "Canary deployment with traffic splitting"
        routingConfig = {
            additionalVersionWeights = {
                "1" = 0.1
            }
        }
    }
}
```

```yaml
version: 2025-11-02

resources:
  canaryAlias:
    type: aws/lambda/alias
    metadata:
      displayName: Canary Alias
    spec:
      functionName: my-lambda-function
      name: CANARY
      functionVersion: "2"
      description: Canary deployment with traffic splitting
      routingConfig:
        additionalVersionWeights:
          "1": 0.1
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "canaryAlias": {
      "type": "aws/lambda/alias",
      "metadata": {
        "displayName": "Canary Alias"
      },
      "spec": {
        "functionName": "my-lambda-function",
        "name": "CANARY",
        "functionVersion": "2",
        "description": "Canary deployment with traffic splitting",
        "routingConfig": {
          "additionalVersionWeights": {
            "1": 0.1
          }
        }
      }
    }
  }
}
```
