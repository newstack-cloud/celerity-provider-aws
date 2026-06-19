Create an SQS queue with all available configuration options, including redrive policies, a resource-based policy, encryption, and tags.

```blueprintlang
version "2025-11-02"

resource completeQueue: aws/sqs/queue {
    metadata {
        displayName = "Complete SQS Queue Configuration"
    }
    spec {
        queueName = "complete-queue-example"
        delaySeconds = 30
        messageRetentionPeriod = 86400
        receiveMessageWaitTimeSeconds = 20
        visibilityTimeout = 300
        maximumMessageSize = 262144
        fifoQueue = false
        contentBasedDeduplication = false
        fifoThroughputLimit = "perQueue"
        deduplicationScope = "queue"
        sqsManagedSseEnabled = true
        kmsDataKeyReusePeriodSeconds = 300
        redrivePolicy = {
            deadLetterTargetArn = "arn:aws:sqs:us-west-2:123456789012:my-dlq"
            maxReceiveCount = 5
        }
        redriveAllowPolicy = {
            redrivePermission = "denyAll"
            sourceQueueArns = [
                "arn:aws:sqs:us-west-2:123456789012:source-queue-1",
                "arn:aws:sqs:us-west-2:123456789012:source-queue-2"
            ]
        }
        policy = {
            Version = "2012-10-17"
            Statement = [
                {
                    Effect = "Allow"
                    Principal = {
                        AWS = "arn:aws:iam::123456789012:user/username"
                    }
                    Action = [
                        "sqs:SendMessage",
                        "sqs:ReceiveMessage"
                    ]
                    Resource = "*"
                }
            ]
        }
        tags = [
            {
                key = "Environment"
                value = "Production"
            },
            {
                key = "Project"
                value = "MyProject"
            },
            {
                key = "Owner"
                value = "admin@example.com"
            }
        ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  completeQueue:
    type: aws/sqs/queue
    metadata:
      displayName: Complete SQS Queue Configuration
    spec:
      queueName: complete-queue-example
      delaySeconds: 30
      messageRetentionPeriod: 86400
      receiveMessageWaitTimeSeconds: 20
      visibilityTimeout: 300
      maximumMessageSize: 262144
      fifoQueue: false
      contentBasedDeduplication: false
      fifoThroughputLimit: perQueue
      deduplicationScope: queue
      sqsManagedSseEnabled: true
      kmsDataKeyReusePeriodSeconds: 300
      redrivePolicy:
        deadLetterTargetArn: arn:aws:sqs:us-west-2:123456789012:my-dlq
        maxReceiveCount: 5
      redriveAllowPolicy:
        redrivePermission: denyAll
        sourceQueueArns:
          - arn:aws:sqs:us-west-2:123456789012:source-queue-1
          - arn:aws:sqs:us-west-2:123456789012:source-queue-2
      policy:
        Version: "2012-10-17"
        Statement:
          - Effect: Allow
            Principal:
              AWS: arn:aws:iam::123456789012:user/username
            Action:
              - sqs:SendMessage
              - sqs:ReceiveMessage
            Resource: "*"
      tags:
        - key: Environment
          value: Production
        - key: Project
          value: MyProject
        - key: Owner
          value: admin@example.com
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "completeQueue": {
      "type": "aws/sqs/queue",
      "metadata": {
        "displayName": "Complete SQS Queue Configuration"
      },
      "spec": {
        "queueName": "complete-queue-example",
        "delaySeconds": 30,
        "messageRetentionPeriod": 86400,
        "receiveMessageWaitTimeSeconds": 20,
        "visibilityTimeout": 300,
        "maximumMessageSize": 262144,
        "fifoQueue": false,
        "contentBasedDeduplication": false,
        "fifoThroughputLimit": "perQueue",
        "deduplicationScope": "queue",
        "sqsManagedSseEnabled": true,
        "kmsDataKeyReusePeriodSeconds": 300,
        "redrivePolicy": {
          "deadLetterTargetArn": "arn:aws:sqs:us-west-2:123456789012:my-dlq",
          "maxReceiveCount": 5
        },
        "redriveAllowPolicy": {
          "redrivePermission": "denyAll",
          "sourceQueueArns": [
            "arn:aws:sqs:us-west-2:123456789012:source-queue-1",
            "arn:aws:sqs:us-west-2:123456789012:source-queue-2"
          ]
        },
        "policy": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Principal": {
                "AWS": "arn:aws:iam::123456789012:user/username"
              },
              "Action": [
                "sqs:SendMessage",
                "sqs:ReceiveMessage"
              ],
              "Resource": "*"
            }
          ]
        },
        "tags": [
          {
            "key": "Environment",
            "value": "Production"
          },
          {
            "key": "Project",
            "value": "MyProject"
          },
          {
            "key": "Owner",
            "value": "admin@example.com"
          }
        ]
      }
    }
  }
}
```
