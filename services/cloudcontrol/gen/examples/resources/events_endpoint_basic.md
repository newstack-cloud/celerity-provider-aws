A basic AWS Events Endpoint with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource endpoint: aws/events/endpoint {
    metadata {
        displayName = "AWS Events Endpoint basic"
    }
    spec {
        eventBuses = [
            {
                eventBusArn = "example-event-bus-arn"
            }
        ]
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
            displayName: AWS Events Endpoint basic
        spec:
            eventBuses:
                - eventBusArn: example-event-bus-arn
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
        "displayName": "AWS Events Endpoint basic"
      },
      "spec": {
        "eventBuses": [
          {
            "eventBusArn": "example-event-bus-arn"
          }
        ],
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
