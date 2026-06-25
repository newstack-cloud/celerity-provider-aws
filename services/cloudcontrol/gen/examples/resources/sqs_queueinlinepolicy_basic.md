A basic AWS SQS QueueInlinePolicy with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource queueInlinePolicy: aws/sqs/queueInlinePolicy {
    metadata {
        displayName = "AWS SQS QueueInlinePolicy basic"
    }
    spec {
        policyDocument = {
            statement = [
                {
                    action = [
                        "s3:GetObject"
                    ],
                    effect = "Allow",
                    resource = "arn:aws:s3:::example-bucket/*"
                }
            ],
            version = "2012-10-17"
        }
        queue = "example-queue"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    queueInlinePolicy:
        type: aws/sqs/queueInlinePolicy
        metadata:
            displayName: AWS SQS QueueInlinePolicy basic
        spec:
            policyDocument:
                statement:
                    - action:
                        - s3:GetObject
                      effect: Allow
                      resource: arn:aws:s3:::example-bucket/*
                version: "2012-10-17"
            queue: example-queue
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "queueInlinePolicy": {
      "type": "aws/sqs/queueInlinePolicy",
      "metadata": {
        "displayName": "AWS SQS QueueInlinePolicy basic"
      },
      "spec": {
        "policyDocument": {
          "statement": [
            {
              "action": [
                "s3:GetObject"
              ],
              "effect": "Allow",
              "resource": "arn:aws:s3:::example-bucket/*"
            }
          ],
          "version": "2012-10-17"
        },
        "queue": "example-queue"
      }
    }
  }
}
```
