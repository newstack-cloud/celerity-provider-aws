A AWS IAM ServiceLinkedRole configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource serviceLinkedRole: aws/iam/serviceLinkedRole {
    metadata {
        displayName = "AWS IAM ServiceLinkedRole complete"
    }
    spec {
        awsServiceName = "example-aws-service-name"
        customSuffix = "example-custom-suffix"
        description = "example-description"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    serviceLinkedRole:
        type: aws/iam/serviceLinkedRole
        metadata:
            displayName: AWS IAM ServiceLinkedRole complete
        spec:
            awsServiceName: example-aws-service-name
            customSuffix: example-custom-suffix
            description: example-description
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "serviceLinkedRole": {
      "type": "aws/iam/serviceLinkedRole",
      "metadata": {
        "displayName": "AWS IAM ServiceLinkedRole complete"
      },
      "spec": {
        "awsServiceName": "example-aws-service-name",
        "customSuffix": "example-custom-suffix",
        "description": "example-description"
      }
    }
  }
}
```
