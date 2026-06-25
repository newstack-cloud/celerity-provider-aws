A basic AWS IAM ManagedPolicy with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource managedPolicy: aws/iam/managedPolicy {
    metadata {
        displayName = "AWS IAM ManagedPolicy basic"
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
    }
}
```

```yaml
version: "2025-11-02"
resources:
    managedPolicy:
        type: aws/iam/managedPolicy
        metadata:
            displayName: AWS IAM ManagedPolicy basic
        spec:
            policyDocument:
                statement:
                    - action:
                        - s3:GetObject
                      effect: Allow
                      resource: arn:aws:s3:::example-bucket/*
                version: "2012-10-17"
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "managedPolicy": {
      "type": "aws/iam/managedPolicy",
      "metadata": {
        "displayName": "AWS IAM ManagedPolicy basic"
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
        }
      }
    }
  }
}
```
