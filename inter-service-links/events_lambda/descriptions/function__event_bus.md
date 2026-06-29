## Lambda Function to EventBridge Event Bus Link

This link grants a Lambda function permission to publish events to an EventBridge event bus by:

1. **Environment Variables** (optional): Populates an environment variable in the Lambda function with the target event bus name, so application code can call `PutEvents` against it.

2. **IAM Permissions**: Adds an inline policy statement to the Lambda function's execution role granting `events:PutEvents` on the event bus ARN.

### Requirements

The Lambda function's execution role must be defined as a resource in the same blueprint. The link populates that role with the required permissions, you define the role with minimal permissions and let links activate the rest.

### Example

```blueprintlang
version "2025-11-02"

resource publishOrderFunction: aws/lambda/function {
    metadata {
        labels = {
            bus = "orders"
        }
        annotations = {
            "aws.lambda.events.orderEventBus.envVarName" = "ORDER_EVENT_BUS"
        }
    }

    select by label {
        bus = "orders"
    }

    spec {
        functionName = "publish-order"
        handler = "publish.handler"
        runtime = "nodejs22.x"
        role = publishOrderFunctionRole.spec.arn
    }
}

resource orderEventBus: aws/events/eventBus {
    metadata {
        labels = {
            bus = "orders"
        }
    }

    spec {
        name = "order-events"
    }
}

resource publishOrderFunctionRole: aws/iam/role {
    spec {
        name = "publish-order-role"
        assumeRolePolicyDocument = {
            Version = "2012-10-17",
            Statement = [
                {
                    Effect = "Allow",
                    Principal = {
                        Service = "lambda.amazonaws.com"
                    },
                    Action = "sts:AssumeRole"
                }
            ]
        }
    }
}
```
