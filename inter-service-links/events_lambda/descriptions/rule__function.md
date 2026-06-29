## EventBridge Rule to Lambda Function

Lets an EventBridge rule invoke a Lambda function.

When a rule targets a function, the function is configured to allow the rule to invoke it, so events matched by the rule trigger the function. No IAM role is required.

Add the function as a target on the rule by referencing the function's `arn` in the rule's `targets`; the connection takes effect automatically, with no link selector required. Per-target options such as the target id, input transformation and retry configuration are set on the rule's target entry.

### Example

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "orderCreatedRule": {
      "type": "aws/events/rule",
      "spec": {
        "name": "order-created-rule",
        "eventPattern": { "source": ["app.orders"] },
        "targets": [
          {
            "id": "process-order",
            "arn": "${processOrderFunction.spec.arn}"
          }
        ]
      }
    },
    "processOrderFunction": {
      "type": "aws/lambda/function",
      "spec": {
        "functionName": "process-order",
        "handler": "process.handler",
        "runtime": "nodejs22.x",
        "role": "${processOrderFunctionRole.spec.arn}"
      }
    }
  }
}
```
