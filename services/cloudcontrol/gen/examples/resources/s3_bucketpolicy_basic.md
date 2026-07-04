A basic AWS S3 BucketPolicy with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource bucketPolicy: aws/s3/bucketPolicy {
    metadata {
        displayName = "AWS S3 BucketPolicy basic"
    }
    spec {
        bucket = "example-bucket"
        policyDocument = {
            exampleKey = "example-value"
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    bucketPolicy:
        type: aws/s3/bucketPolicy
        metadata:
            displayName: AWS S3 BucketPolicy basic
        spec:
            bucket: example-bucket
            policyDocument:
                exampleKey: example-value
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "bucketPolicy": {
      "type": "aws/s3/bucketPolicy",
      "metadata": {
        "displayName": "AWS S3 BucketPolicy basic"
      },
      "spec": {
        "bucket": "example-bucket",
        "policyDocument": {
          "exampleKey": "example-value"
        }
      }
    }
  }
}
```
