## EventBridge Rule to API Destination Link

This link grants an EventBridge rule permission to invoke an EventBridge API destination. API destination targets are invoked by EventBridge assuming the IAM role referenced by the rule's matching target entry (`targets[].roleArn`), so this link "activates" that role by attaching an inline policy statement granting `events:InvokeApiDestination` on the API destination's ARN.

The practitioner must define the IAM role as a top-level blueprint resource (an `aws/iam/role`) and reference it from the rule's target entry via its `roleArn` field.

The rule-to-destination wiring (the target entry, input transformation, retry configuration) is modelled inline in the rule's `targets[]` array and owned by the `aws/events/rule` resource, which references the API destination via the target entry's `arn` field. That reference is what activates this link, `targets[].arn` is a link wiring slot, so no `linkSelector` is required. This link carries no user input.

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
            "id": "order-webhook",
            "arn": "${orderWebhookDestination.spec.arn}",
            "roleArn": "${orderWebhookRole.spec.arn}"
          }
        ]
      }
    },
    "orderWebhookConnection": {
      "type": "aws/events/connection",
      "spec": {
        "name": "order-webhook-connection",
        "authorizationType": "API_KEY",
        "authParameters": {
          "apiKeyAuthParameters": {
            "apiKeyName": "x-api-key",
            "apiKeyValue": "${variables.webhookApiKey}"
          }
        }
      }
    },
    "orderWebhookDestination": {
      "type": "aws/events/apiDestination",
      "spec": {
        "name": "order-webhook-destination",
        "connectionArn": "${orderWebhookConnection.spec.arn}",
        "invocationEndpoint": "https://example.com/hooks/orders",
        "httpMethod": "POST"
      }
    },
    "orderWebhookRole": {
      "type": "aws/iam/role",
      "spec": {
        "roleName": "order-webhook-invoke-role",
        "assumeRolePolicyDocument": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Principal": { "Service": "events.amazonaws.com" },
              "Action": "sts:AssumeRole"
            }
          ]
        }
      }
    }
  }
}
```
