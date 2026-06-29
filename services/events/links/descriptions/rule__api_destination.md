## EventBridge Rule to API Destination

Lets an EventBridge rule send matched events to an EventBridge API destination.

EventBridge invokes an API destination using an IAM role you provide. Define that role as an `aws/iam/role` resource in the blueprint and reference it from the rule's target entry via its `roleArn` field; the role is granted permission to invoke the destination, so events matched by the rule are sent to it.

Add the destination as a target on the rule by referencing the destination's `arn` in the rule's `targets`; the connection takes effect automatically, with no link selector required. Per-target options including the target id, input transformation and retry configuration are set on the rule's target entry.

### Example

```blueprintlang
version "2025-11-02"

resource orderCreatedRule: aws/events/rule {
    spec {
        name = "order-created-rule"
        eventPattern = {
            source = ["app.orders"]
        }
        targets = [
            {
                id = "order-webhook",
                arn = orderWebhookDestination.spec.arn,
                roleArn = orderWebhookRole.spec.arn
            }
        ]
    }
}

resource orderWebhookConnection: aws/events/connection {
    spec {
        name = "order-webhook-connection"
        authorizationType = "API_KEY"
        authParameters = {
            apiKeyAuthParameters = {
                apiKeyName = "x-api-key",
                apiKeyValue = variables.webhookApiKey
            }
        }
    }
}

resource orderWebhookDestination: aws/events/apiDestination {
    spec {
        name = "order-webhook-destination"
        connectionArn = orderWebhookConnection.spec.arn
        invocationEndpoint = "https://example.com/hooks/orders"
        httpMethod = "POST"
    }
}

resource orderWebhookRole: aws/iam/role {
    spec {
        roleName = "order-webhook-invoke-role"
        assumeRolePolicyDocument = {
            Version = "2012-10-17",
            Statement = [
                {
                    Effect = "Allow",
                    Principal = {
                        Service = "events.amazonaws.com"
                    },
                    Action = "sts:AssumeRole"
                }
            ]
        }
    }
}
```
