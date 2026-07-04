A basic AWS S3 Bucket with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource bucket: aws/s3/bucket {
    metadata {
        displayName = "AWS S3 Bucket basic"
    }
    spec {
        bucketName = "example-bucket-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    bucket:
        type: aws/s3/bucket
        metadata:
            displayName: AWS S3 Bucket basic
        spec:
            bucketName: example-bucket-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "bucket": {
      "type": "aws/s3/bucket",
      "metadata": {
        "displayName": "AWS S3 Bucket basic"
      },
      "spec": {
        "bucketName": "example-bucket-name"
      }
    }
  }
}
```
