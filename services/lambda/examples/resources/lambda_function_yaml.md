Define a Lambda function whose deployment package is stored in an S3 bucket.

```blueprintlang
version "2025-11-02"

resource processOrdersFunction: aws/lambda/function {
    metadata {
        displayName = "Order Processing Function"
        description = "This function processes customer orders."
        labels {
            app = "orders"
        }
    }
    spec {
        functionName = "orders-ProcessOrdersFunction-v1"
        code = {
            s3Bucket = "my-bucket"
            s3Key = "order-processing.zip"
        }
        role = "arn:aws:iam::123456789012:role/lambda-execution-role"
        handler = "index.handler"
        runtime = "nodejs22.x"
        memorySize = 256
        timeout = 30
        environment = {
            variables = {
                ORDER_QUEUE_URL = "https://sqs.us-east-1.amazonaws.com/123456789012/order-queue"
            }
        }
        tracingConfig = {
            mode = "Active"
        }
        vpcConfig = {
            securityGroupIds = ["sg-0123456789abcdef0", "sg-0fedcba9876543210"]
            subnetIds = ["subnet-0123456789abcdef0", "subnet-0fedcba9876543210"]
        }
        tags = {
            Environment = "Production"
            Application = "OrderProcessing"
        }
    }
}
```

```yaml
version: 2025-11-02

resources:
  processOrdersFunction:
    type: aws/lambda/function
    metadata:
      displayName: Order Processing Function
      description: This function processes customer orders.
      labels:
        app: orders
    spec:
      functionName: orders-ProcessOrdersFunction-v1
      code:
        s3Bucket: my-bucket
        s3Key: order-processing.zip
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      handler: index.handler
      runtime: nodejs22.x
      memorySize: 256
      timeout: 30
      environment:
        variables:
          ORDER_QUEUE_URL: https://sqs.us-east-1.amazonaws.com/123456789012/order-queue
      tracingConfig:
        mode: Active
      vpcConfig:
        securityGroupIds:
          - sg-0123456789abcdef0
          - sg-0fedcba9876543210
        subnetIds:
          - subnet-0123456789abcdef0
          - subnet-0fedcba9876543210
      tags:
        Environment: Production
        Application: OrderProcessing
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "processOrdersFunction": {
      "type": "aws/lambda/function",
      "metadata": {
        "displayName": "Order Processing Function",
        "description": "This function processes customer orders.",
        "labels": {
          "app": "orders"
        }
      },
      "spec": {
        "functionName": "orders-ProcessOrdersFunction-v1",
        // The deployment package for this function is stored in an S3 bucket.
        "code": {
          "s3Bucket": "my-bucket",
          "s3Key": "order-processing.zip"
        },
        "role": "arn:aws:iam::123456789012:role/lambda-execution-role",
        "handler": "index.handler",
        "runtime": "nodejs22.x",
        "memorySize": 256,
        "timeout": 30,
        "environment": {
          "variables": {
            "ORDER_QUEUE_URL": "https://sqs.us-east-1.amazonaws.com/123456789012/order-queue"
          }
        },
        "tracingConfig": {
          "mode": "Active"
        },
        "vpcConfig": {
          "securityGroupIds": [
            "sg-0123456789abcdef0",
            "sg-0fedcba9876543210"
          ],
          "subnetIds": [
            "subnet-0123456789abcdef0",
            "subnet-0fedcba9876543210"
          ]
        },
        "tags": {
          "Environment": "Production",
          "Application": "OrderProcessing"
        }
      }
    }
  }
}
```
