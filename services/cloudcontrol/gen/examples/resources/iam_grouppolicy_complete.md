A AWS IAM GroupPolicy configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource groupPolicy: aws/iam/groupPolicy {
    metadata {
        displayName = "AWS IAM GroupPolicy complete"
    }
    spec {
        groupName = "example-group-name"
        policyDocument = {
            statement = [
                {
                    action = [
                        "s3:GetObject"
                    ],
                    effect = "Allow",
                    resource = "arn:aws:s3:::example-bucket/*"
                }
            ],
            version = "2012-10-17"
        }
        policyName = "example-policy-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    groupPolicy:
        type: aws/iam/groupPolicy
        metadata:
            displayName: AWS IAM GroupPolicy complete
        spec:
            groupName: example-group-name
            policyDocument:
                statement:
                    - action:
                        - s3:GetObject
                      effect: Allow
                      resource: arn:aws:s3:::example-bucket/*
                version: "2012-10-17"
            policyName: example-policy-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "groupPolicy": {
      "type": "aws/iam/groupPolicy",
      "metadata": {
        "displayName": "AWS IAM GroupPolicy complete"
      },
      "spec": {
        "groupName": "example-group-name",
        "policyDocument": {
          "statement": [
            {
              "action": [
                "s3:GetObject"
              ],
              "effect": "Allow",
              "resource": "arn:aws:s3:::example-bucket/*"
            }
          ],
          "version": "2012-10-17"
        },
        "policyName": "example-policy-name"
      }
    }
  }
}
```
