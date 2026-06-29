A AWS SNS Subscription configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource subscription: aws/sns/subscription {
    metadata {
        displayName = "AWS SNS Subscription complete"
    }
    spec {
        deliveryPolicy = {
            exampleKey = "example-value"
        }
        endpoint = "example-endpoint"
        filterPolicy = {
            exampleKey = "example-value"
        }
        filterPolicyScope = "example-filter-policy-scope"
        protocol = "example-protocol"
        rawMessageDelivery = false
        redrivePolicy = {
            exampleKey = "example-value"
        }
        region = "example-region"
        replayPolicy = {
            exampleKey = "example-value"
        }
        subscriptionRoleArn = "example-subscription-role-arn"
        topicArn = "example-topic-arn"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    subscription:
        type: aws/sns/subscription
        metadata:
            displayName: AWS SNS Subscription complete
        spec:
            deliveryPolicy:
                exampleKey: example-value
            endpoint: example-endpoint
            filterPolicy:
                exampleKey: example-value
            filterPolicyScope: example-filter-policy-scope
            protocol: example-protocol
            rawMessageDelivery: false
            redrivePolicy:
                exampleKey: example-value
            region: example-region
            replayPolicy:
                exampleKey: example-value
            subscriptionRoleArn: example-subscription-role-arn
            topicArn: example-topic-arn
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "subscription": {
      "type": "aws/sns/subscription",
      "metadata": {
        "displayName": "AWS SNS Subscription complete"
      },
      "spec": {
        "deliveryPolicy": {
          "exampleKey": "example-value"
        },
        "endpoint": "example-endpoint",
        "filterPolicy": {
          "exampleKey": "example-value"
        },
        "filterPolicyScope": "example-filter-policy-scope",
        "protocol": "example-protocol",
        "rawMessageDelivery": false,
        "redrivePolicy": {
          "exampleKey": "example-value"
        },
        "region": "example-region",
        "replayPolicy": {
          "exampleKey": "example-value"
        },
        "subscriptionRoleArn": "example-subscription-role-arn",
        "topicArn": "example-topic-arn"
      }
    }
  }
}
```
