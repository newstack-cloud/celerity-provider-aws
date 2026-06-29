A basic AWS SNS Subscription with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource subscription: aws/sns/subscription {
    metadata {
        displayName = "AWS SNS Subscription basic"
    }
    spec {
        protocol = "example-protocol"
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
            displayName: AWS SNS Subscription basic
        spec:
            protocol: example-protocol
            topicArn: example-topic-arn
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "subscription": {
      "type": "aws/sns/subscription",
      "metadata": {
        "displayName": "AWS SNS Subscription basic"
      },
      "spec": {
        "protocol": "example-protocol",
        "topicArn": "example-topic-arn"
      }
    }
  }
}
```
