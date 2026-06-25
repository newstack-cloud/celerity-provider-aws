A AWS Lambda Alias configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource alias: aws/lambda/alias {
    metadata {
        displayName = "AWS Lambda Alias complete"
    }
    spec {
        description = "example-description"
        functionName = "example-function-name"
        functionVersion = "example-function-version"
        name = "example-name"
        provisionedConcurrencyConfig = {
            provisionedConcurrentExecutions = 1
        }
        routingConfig = {
            additionalVersionWeights = [
                {
                    functionVersion = "example-function-version",
                    functionWeight = 1
                }
            ]
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    alias:
        type: aws/lambda/alias
        metadata:
            displayName: AWS Lambda Alias complete
        spec:
            description: example-description
            functionName: example-function-name
            functionVersion: example-function-version
            name: example-name
            provisionedConcurrencyConfig:
                provisionedConcurrentExecutions: 1
            routingConfig:
                additionalVersionWeights:
                    - functionVersion: example-function-version
                      functionWeight: 1
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "alias": {
      "type": "aws/lambda/alias",
      "metadata": {
        "displayName": "AWS Lambda Alias complete"
      },
      "spec": {
        "description": "example-description",
        "functionName": "example-function-name",
        "functionVersion": "example-function-version",
        "name": "example-name",
        "provisionedConcurrencyConfig": {
          "provisionedConcurrentExecutions": 1
        },
        "routingConfig": {
          "additionalVersionWeights": [
            {
              "functionVersion": "example-function-version",
              "functionWeight": 1
            }
          ]
        }
      }
    }
  }
}
```
