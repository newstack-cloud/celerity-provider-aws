**SQS Queue - JSONC**

This example shows how to create an SQS queue using JSONC format with comments.

```javascript
{
  "resources": {
    "myQueue": {
      "type": "aws/sqs/queue",
      "metadata": {
        "displayName": "My SQS Queue with JSONC"
      },
      "spec": {
        // Queue name - must be unique within the region
        "queueName": "my-jsonc-queue",
        
        // Delay all messages by 15 seconds
        "delaySeconds": 15,
        
        // Redrive policy for dead letter queue
        "redrivePolicy": {
          "deadLetterTargetArn": "arn:aws:sqs:us-west-2:123456789012:my-dlq",
          "maxReceiveCount": 3
        },
        
        // Keep messages for 2 days
        "messageRetentionPeriod": 172800,
        
        // Use long polling for better efficiency
        "receiveMessageWaitTimeSeconds": 20,
        
        // Hide messages from other consumers for 2 minutes
        "visibilityTimeout": 120,
        
        // Maximum message size (256 KB)
        "maximumMessageSize": 262144,
        
        // This is a standard queue (not FIFO)
        "fifoQueue": false,
        
        // Enable SQS-managed server-side encryption
        "sqsManagedSseEnabled": true,
        
        // KMS data key reuse period (5 minutes)
        "kmsDataKeyReusePeriodSeconds": 300,
        
        
        // Resource-based policy for access control
        "policy": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Principal": {
                "Service": "lambda.amazonaws.com"
              },
              "Action": [
                "sqs:SendMessage",
                "sqs:ReceiveMessage",
                "sqs:DeleteMessage"
              ],
              "Resource": "*"
            }
          ]
        },
        
        // Tags for resource management
        "tags": [
          {
            "key": "Environment",
            "value": "Development"
          },
          {
            "key": "Service",
            "value": "MessageProcessing"
          },
          {
            "key": "Team",
            "value": "Backend"
          }
        ]
      }
    }
  }
}
```
