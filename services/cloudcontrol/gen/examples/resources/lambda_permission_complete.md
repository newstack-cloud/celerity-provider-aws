A AWS Lambda Permission configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource permission: aws/lambda/permission {
    metadata {
        displayName = "AWS Lambda Permission complete"
    }
    spec {
        action = "example-action"
        eventSourceToken = "example-event-source-token"
        functionName = "example-function-name"
        functionUrlAuthType = "AWS_IAM"
        invokedViaFunctionUrl = false
        principal = "example-principal"
        principalOrgID = "example-principal-org-i-d"
        sourceAccount = "example-source-account"
        sourceArn = "example-source-arn"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    permission:
        type: aws/lambda/permission
        metadata:
            displayName: AWS Lambda Permission complete
        spec:
            action: example-action
            eventSourceToken: example-event-source-token
            functionName: example-function-name
            functionUrlAuthType: AWS_IAM
            invokedViaFunctionUrl: false
            principal: example-principal
            principalOrgID: example-principal-org-i-d
            sourceAccount: example-source-account
            sourceArn: example-source-arn
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "permission": {
      "type": "aws/lambda/permission",
      "metadata": {
        "displayName": "AWS Lambda Permission complete"
      },
      "spec": {
        "action": "example-action",
        "eventSourceToken": "example-event-source-token",
        "functionName": "example-function-name",
        "functionUrlAuthType": "AWS_IAM",
        "invokedViaFunctionUrl": false,
        "principal": "example-principal",
        "principalOrgID": "example-principal-org-i-d",
        "sourceAccount": "example-source-account",
        "sourceArn": "example-source-arn"
      }
    }
  }
}
```
