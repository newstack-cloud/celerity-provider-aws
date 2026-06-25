A AWS IAM UserPolicy configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource userPolicy: aws/iam/userPolicy {
    metadata {
        displayName = "AWS IAM UserPolicy complete"
    }
    spec {
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
        userName = "example-user-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    userPolicy:
        type: aws/iam/userPolicy
        metadata:
            displayName: AWS IAM UserPolicy complete
        spec:
            policyDocument:
                statement:
                    - action:
                        - s3:GetObject
                      effect: Allow
                      resource: arn:aws:s3:::example-bucket/*
                version: "2012-10-17"
            policyName: example-policy-name
            userName: example-user-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "userPolicy": {
      "type": "aws/iam/userPolicy",
      "metadata": {
        "displayName": "AWS IAM UserPolicy complete"
      },
      "spec": {
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
        "policyName": "example-policy-name",
        "userName": "example-user-name"
      }
    }
  }
}
```
