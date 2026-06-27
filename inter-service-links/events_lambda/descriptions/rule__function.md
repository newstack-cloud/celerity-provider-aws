## EventBridge Rule to Lambda Function Link

This link grants an EventBridge rule permission to invoke a Lambda function. Lambda targets are invoked by EventBridge using the function's resource-based policy (no IAM role is required), so this link deploys a single, link-owned `aws/lambda/permission` resource allowing the principal `events.amazonaws.com` to call `lambda:InvokeFunction`, scoped by `SourceArn` to the rule's ARN.

The rule-to-function wiring itself (the target entry, plus any input transformation or retry configuration) is modelled inline in the rule's `targets[]` array and owned by the `aws/events/rule` resource, which references the function via the target entry's `arn` field. That reference is what activates this link, `targets[].arn` is a link wiring slot, so no `linkSelector` is required. This link carries no user input.

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
