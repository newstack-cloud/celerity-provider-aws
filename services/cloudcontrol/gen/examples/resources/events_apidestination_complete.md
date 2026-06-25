A AWS Events ApiDestination configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource apiDestination: aws/events/apiDestination {
    metadata {
        displayName = "AWS Events ApiDestination complete"
    }
    spec {
        connectionArn = "example-connection-arn"
        description = "example-description"
        httpMethod = "GET"
        invocationEndpoint = "example-invocation-endpoint"
        invocationRateLimitPerSecond = 1
        name = "example-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    apiDestination:
        type: aws/events/apiDestination
        metadata:
            displayName: AWS Events ApiDestination complete
        spec:
            connectionArn: example-connection-arn
            description: example-description
            httpMethod: GET
            invocationEndpoint: example-invocation-endpoint
            invocationRateLimitPerSecond: 1
            name: example-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "apiDestination": {
      "type": "aws/events/apiDestination",
      "metadata": {
        "displayName": "AWS Events ApiDestination complete"
      },
      "spec": {
        "connectionArn": "example-connection-arn",
        "description": "example-description",
        "httpMethod": "GET",
        "invocationEndpoint": "example-invocation-endpoint",
        "invocationRateLimitPerSecond": 1,
        "name": "example-name"
      }
    }
  }
}
```
