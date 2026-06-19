Define a Lambda function using inline code. Ensure the execution role has the S3 permissions the handler needs.

```blueprintlang
version "2025-11-02"

resource getOrderFunction: aws/lambda/function {
    metadata {
        displayName = "Order Retrieval Function"
        description = "This function retrieves a customer order given an order ID."
        labels {
            app = "orders"
        }
    }
    spec {
        functionName = "orders-GetOrderFunction-v1"
        code = {
            zipFile = """const { S3Client, GetObjectCommand } = require("@aws-sdk/client-s3");
const s3 = new S3Client({ region: "us-west-2" });

async function handler(event) {
  const response = await s3.send(new GetObjectCommand({
    Bucket: process.env.ORDER_BUCKET,
    Key: `${event.orderId}.json`
  }));
  return response.Body;
}

export default handler;
"""
        }
        role = "arn:aws:iam::123456789012:role/lambda-execution-role"
        handler = "index.handler"
        runtime = "nodejs22.x"
        memorySize = 256
        timeout = 30
        environment = {
            variables = {
                ORDER_BUCKET = "orders-bucket"
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
  getOrderFunction:
    type: aws/lambda/function
    metadata:
      displayName: Order Retrieval Function
      description: This function retrieves a customer order given an order ID.
      labels:
        app: orders
    spec:
      functionName: orders-GetOrderFunction-v1
      code:
        zipFile: |
          const { S3Client, GetObjectCommand } = require("@aws-sdk/client-s3");
          const s3 = new S3Client({ region: "us-west-2" });

          async function handler(event) {
            const response = await s3.send(new GetObjectCommand({
              Bucket: process.env.ORDER_BUCKET,
              Key: `${event.orderId}.json`
            }));
            return response.Body;
          }

          export default handler;
      role: arn:aws:iam::123456789012:role/lambda-execution-role
      handler: index.handler
      runtime: nodejs22.x
      memorySize: 256
      timeout: 30
      environment:
        variables:
          ORDER_BUCKET: orders-bucket
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
    "getOrderFunction": {
      "type": "aws/lambda/function",
      "metadata": {
        "displayName": "Order Retrieval Function",
        "description": "This function retrieves a customer order given an order ID.",
        "labels": {
          "app": "orders"
        }
      },
      "spec": {
        "functionName": "orders-GetOrderFunction-v1",
        // The execution role must grant the Amazon S3 read permissions this handler needs.
        "code": {
          "zipFile": "const { S3Client, GetObjectCommand } = require(\"@aws-sdk/client-s3\");\nconst s3 = new S3Client({ region: \"us-west-2\" });\n\nasync function handler(event) {\n  const response = await s3.send(new GetObjectCommand({\n    Bucket: process.env.ORDER_BUCKET,\n    Key: `${event.orderId}.json`\n  }));\n  return response.Body;\n}\n\nexport default handler;\n"
        },
        "role": "arn:aws:iam::123456789012:role/lambda-execution-role",
        "handler": "index.handler",
        "runtime": "nodejs22.x",
        "memorySize": 256,
        "timeout": 30,
        "environment": {
          "variables": {
            "ORDER_BUCKET": "orders-bucket"
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
