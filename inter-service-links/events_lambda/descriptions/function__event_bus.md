## Lambda Function to EventBridge Event Bus Link

This link grants a Lambda function permission to publish events to an EventBridge event bus by:

1. **Environment Variables** (optional): Populates an environment variable in the Lambda function with the target event bus name, so application code can call `PutEvents` against it.

2. **IAM Permissions**: Adds an inline policy statement to the Lambda function's execution role granting `events:PutEvents` on the event bus ARN.

### Requirements

The Lambda function's execution role must be defined as a resource in the same blueprint. The link populates that role with the required permissions, you define the role with minimal permissions and let links activate the rest.

### Annotations

- `aws.lambda.events.populateEnvVars` — populate environment variables for all linked event buses (default `true`).
- `aws.lambda.events.<targetBus>.populateEnvVars` — override for a specific target event bus.
- `aws.lambda.events.<targetBus>.envVarName` — custom environment variable name (default `EVENT_BUS_<targetBus>`).

### Example

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "publishOrderFunction": {
      "type": "aws/lambda/function",
      "metadata": {
        "labels": { "bus": "orders" },
        "annotations": {
          "aws.lambda.events.orderEventBus.envVarName": "ORDER_EVENT_BUS"
        }
      },
      "linkSelector": { "byLabel": { "bus": "orders" } },
      "spec": {
        "functionName": "publish-order",
        "handler": "publish.handler",
        "runtime": "nodejs22.x",
        "role": "${publishOrderFunctionRole.spec.arn}"
      }
    },
    "orderEventBus": {
      "type": "aws/events/eventBus",
      "metadata": { "labels": { "bus": "orders" } },
      "spec": { "name": "order-events" }
    },
    "publishOrderFunctionRole": {
      "type": "aws/iam/role",
      "spec": {
        "name": "publish-order-role",
        "assumeRolePolicyDocument": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Principal": { "Service": "lambda.amazonaws.com" },
              "Action": "sts:AssumeRole"
            }
          ]
        }
      }
    }
  }
}
```
