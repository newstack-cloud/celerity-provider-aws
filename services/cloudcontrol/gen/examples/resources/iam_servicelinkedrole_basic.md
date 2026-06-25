A basic AWS IAM ServiceLinkedRole with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource serviceLinkedRole: aws/iam/serviceLinkedRole {
    metadata {
        displayName = "AWS IAM ServiceLinkedRole basic"
    }
    spec {
        awsServiceName = "example-aws-service-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    serviceLinkedRole:
        type: aws/iam/serviceLinkedRole
        metadata:
            displayName: AWS IAM ServiceLinkedRole basic
        spec:
            awsServiceName: example-aws-service-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "serviceLinkedRole": {
      "type": "aws/iam/serviceLinkedRole",
      "metadata": {
        "displayName": "AWS IAM ServiceLinkedRole basic"
      },
      "spec": {
        "awsServiceName": "example-aws-service-name"
      }
    }
  }
}
```
