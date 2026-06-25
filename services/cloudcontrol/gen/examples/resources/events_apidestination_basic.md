A basic AWS Events ApiDestination with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource apiDestination: aws/events/apiDestination {
    metadata {
        displayName = "AWS Events ApiDestination basic"
    }
    spec {
        connectionArn = "example-connection-arn"
        httpMethod = "GET"
        invocationEndpoint = "example-invocation-endpoint"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    apiDestination:
        type: aws/events/apiDestination
        metadata:
            displayName: AWS Events ApiDestination basic
        spec:
            connectionArn: example-connection-arn
            httpMethod: GET
            invocationEndpoint: example-invocation-endpoint
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "apiDestination": {
      "type": "aws/events/apiDestination",
      "metadata": {
        "displayName": "AWS Events ApiDestination basic"
      },
      "spec": {
        "connectionArn": "example-connection-arn",
        "httpMethod": "GET",
        "invocationEndpoint": "example-invocation-endpoint"
      }
    }
  }
}
```
