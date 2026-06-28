A AWS Logs LogGroup configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource logGroup: aws/logs/logGroup {
    metadata {
        displayName = "AWS Logs LogGroup complete"
    }
    spec {
        bearerTokenAuthenticationEnabled = false
        dataProtectionPolicy = {
            exampleKey = "example-value"
        }
        deletionProtectionEnabled = false
        fieldIndexPolicies = [
            {
                exampleKey = "example-value"
            }
        ]
        kmsKeyId = "example-kms-key-id"
        logGroupClass = "STANDARD"
        logGroupName = "example-log-group-name"
        resourcePolicyDocument = {
            exampleKey = "example-value"
        }
        retentionInDays = 1
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    logGroup:
        type: aws/logs/logGroup
        metadata:
            displayName: AWS Logs LogGroup complete
        spec:
            bearerTokenAuthenticationEnabled: false
            dataProtectionPolicy:
                exampleKey: example-value
            deletionProtectionEnabled: false
            fieldIndexPolicies:
                - exampleKey: example-value
            kmsKeyId: example-kms-key-id
            logGroupClass: STANDARD
            logGroupName: example-log-group-name
            resourcePolicyDocument:
                exampleKey: example-value
            retentionInDays: 1
            tags:
                - key: example-key
                  value: example-value
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "logGroup": {
      "type": "aws/logs/logGroup",
      "metadata": {
        "displayName": "AWS Logs LogGroup complete"
      },
      "spec": {
        "bearerTokenAuthenticationEnabled": false,
        "dataProtectionPolicy": {
          "exampleKey": "example-value"
        },
        "deletionProtectionEnabled": false,
        "fieldIndexPolicies": [
          {
            "exampleKey": "example-value"
          }
        ],
        "kmsKeyId": "example-kms-key-id",
        "logGroupClass": "STANDARD",
        "logGroupName": "example-log-group-name",
        "resourcePolicyDocument": {
          "exampleKey": "example-value"
        },
        "retentionInDays": 1,
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ]
      }
    }
  }
}
```
