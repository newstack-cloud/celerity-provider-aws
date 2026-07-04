A basic AWS SNS TopicInlinePolicy with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource topicInlinePolicy: aws/sns/topicInlinePolicy {
    metadata {
        displayName = "AWS SNS TopicInlinePolicy basic"
    }
    spec {
        policyDocument = {
            exampleKey = "example-value"
        }
        topicArn = "example-topic-arn"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    topicInlinePolicy:
        type: aws/sns/topicInlinePolicy
        metadata:
            displayName: AWS SNS TopicInlinePolicy basic
        spec:
            policyDocument:
                exampleKey: example-value
            topicArn: example-topic-arn
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "topicInlinePolicy": {
      "type": "aws/sns/topicInlinePolicy",
      "metadata": {
        "displayName": "AWS SNS TopicInlinePolicy basic"
      },
      "spec": {
        "policyDocument": {
          "exampleKey": "example-value"
        },
        "topicArn": "example-topic-arn"
      }
    }
  }
}
```
