A AWS Events Endpoint configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource endpoint: aws/events/endpoint {
    metadata {
        displayName = "AWS Events Endpoint complete"
    }
    spec {
        description = "example-description"
        eventBuses = [
            {
                eventBusArn = "example-event-bus-arn"
            }
        ]
        name = "example-name"
        replicationConfig = {
            state = "ENABLED"
        }
        roleArn = "example-role-arn"
        routingConfig = {
            failoverConfig = {
                primary = {
                    healthCheck = "example-health-check"
                },
                secondary = {
                    route = "example-route"
                }
            }
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    endpoint:
        type: aws/events/endpoint
        metadata:
            displayName: AWS Events Endpoint complete
        spec:
            description: example-description
            eventBuses:
                - eventBusArn: example-event-bus-arn
            name: example-name
            replicationConfig:
                state: ENABLED
            roleArn: example-role-arn
            routingConfig:
                failoverConfig:
                    primary:
                        healthCheck: example-health-check
                    secondary:
                        route: example-route
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "endpoint": {
      "type": "aws/events/endpoint",
      "metadata": {
        "displayName": "AWS Events Endpoint complete"
      },
      "spec": {
        "description": "example-description",
        "eventBuses": [
          {
            "eventBusArn": "example-event-bus-arn"
          }
        ],
        "name": "example-name",
        "replicationConfig": {
          "state": "ENABLED"
        },
        "roleArn": "example-role-arn",
        "routingConfig": {
          "failoverConfig": {
            "primary": {
              "healthCheck": "example-health-check"
            },
            "secondary": {
              "route": "example-route"
            }
          }
        }
      }
    }
  }
}
```
