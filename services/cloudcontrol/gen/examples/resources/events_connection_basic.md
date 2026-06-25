A basic AWS Events Connection using API key authentication, with the credentials authored as a native object.

```blueprintlang
version "2025-11-02"

resource connection: aws/events/connection {
    metadata {
        displayName = "Order events connection"
    }
    spec {
        name = "order-events-connection"
        authorizationType = "API_KEY"
        authParameters = {
            ApiKeyAuthParameters = {
                ApiKeyName = "x-api-key"
                ApiKeyValue = "example-secret-value"
            }
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    connection:
        type: aws/events/connection
        metadata:
            displayName: Order events connection
        spec:
            name: order-events-connection
            authorizationType: API_KEY
            authParameters:
                ApiKeyAuthParameters:
                    ApiKeyName: x-api-key
                    ApiKeyValue: example-secret-value
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "connection": {
      "type": "aws/events/connection",
      "metadata": {
        "displayName": "Order events connection"
      },
      "spec": {
        "name": "order-events-connection",
        "authorizationType": "API_KEY",
        "authParameters": {
          "ApiKeyAuthParameters": {
            "ApiKeyName": "x-api-key",
            "ApiKeyValue": "example-secret-value"
          }
        }
      }
    }
  }
}
```
