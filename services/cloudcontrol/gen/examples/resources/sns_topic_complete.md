A AWS SNS Topic configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource topic: aws/sns/topic {
    metadata {
        displayName = "AWS SNS Topic complete"
    }
    spec {
        archivePolicy = {
            exampleKey = "example-value"
        }
        contentBasedDeduplication = false
        dataProtectionPolicy = {
            exampleKey = "example-value"
        }
        deliveryStatusLogging = [
            {
                failureFeedbackRoleArn = "example-failure-feedback-role-arn",
                protocol = "http/s",
                successFeedbackRoleArn = "example-success-feedback-role-arn",
                successFeedbackSampleRate = "example-success-feedback-sample-rate"
            }
        ]
        displayName = "example-display-name"
        fifoThroughputScope = "example-fifo-throughput-scope"
        fifoTopic = false
        kmsMasterKeyId = "example-kms-master-key-id"
        signatureVersion = "example-signature-version"
        subscription = [
            {
                endpoint = "example-endpoint",
                protocol = "example-protocol"
            }
        ]
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        topicName = "example-topic-name"
        tracingConfig = "example-tracing-config"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    topic:
        type: aws/sns/topic
        metadata:
            displayName: AWS SNS Topic complete
        spec:
            archivePolicy:
                exampleKey: example-value
            contentBasedDeduplication: false
            dataProtectionPolicy:
                exampleKey: example-value
            deliveryStatusLogging:
                - failureFeedbackRoleArn: example-failure-feedback-role-arn
                  protocol: http/s
                  successFeedbackRoleArn: example-success-feedback-role-arn
                  successFeedbackSampleRate: example-success-feedback-sample-rate
            displayName: example-display-name
            fifoThroughputScope: example-fifo-throughput-scope
            fifoTopic: false
            kmsMasterKeyId: example-kms-master-key-id
            signatureVersion: example-signature-version
            subscription:
                - endpoint: example-endpoint
                  protocol: example-protocol
            tags:
                - key: example-key
                  value: example-value
            topicName: example-topic-name
            tracingConfig: example-tracing-config
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "topic": {
      "type": "aws/sns/topic",
      "metadata": {
        "displayName": "AWS SNS Topic complete"
      },
      "spec": {
        "archivePolicy": {
          "exampleKey": "example-value"
        },
        "contentBasedDeduplication": false,
        "dataProtectionPolicy": {
          "exampleKey": "example-value"
        },
        "deliveryStatusLogging": [
          {
            "failureFeedbackRoleArn": "example-failure-feedback-role-arn",
            "protocol": "http/s",
            "successFeedbackRoleArn": "example-success-feedback-role-arn",
            "successFeedbackSampleRate": "example-success-feedback-sample-rate"
          }
        ],
        "displayName": "example-display-name",
        "fifoThroughputScope": "example-fifo-throughput-scope",
        "fifoTopic": false,
        "kmsMasterKeyId": "example-kms-master-key-id",
        "signatureVersion": "example-signature-version",
        "subscription": [
          {
            "endpoint": "example-endpoint",
            "protocol": "example-protocol"
          }
        ],
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "topicName": "example-topic-name",
        "tracingConfig": "example-tracing-config"
      }
    }
  }
}
```
