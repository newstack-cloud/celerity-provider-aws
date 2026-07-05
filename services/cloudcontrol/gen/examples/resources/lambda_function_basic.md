A basic AWS Lambda Function with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource function: aws/lambda/function {
    metadata {
        displayName = "AWS Lambda Function basic"
    }
    spec {
        code = {
            imageUri = "example-image-uri",
            s3Bucket = "example-s3-bucket",
            s3Key = "example-s3-key",
            s3ObjectStorageMode = "COPY",
            s3ObjectVersion = "example-s3-object-version",
            sourceKMSKeyArn = "example-source-k-m-s-key-arn",
            zipFile = "example-zip-file"
        }
        role = "example-role"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    function:
        type: aws/lambda/function
        metadata:
            displayName: AWS Lambda Function basic
        spec:
            code:
                imageUri: example-image-uri
                s3Bucket: example-s3-bucket
                s3Key: example-s3-key
                s3ObjectStorageMode: COPY
                s3ObjectVersion: example-s3-object-version
                sourceKMSKeyArn: example-source-k-m-s-key-arn
                zipFile: example-zip-file
            role: example-role
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "function": {
      "type": "aws/lambda/function",
      "metadata": {
        "displayName": "AWS Lambda Function basic"
      },
      "spec": {
        "code": {
          "imageUri": "example-image-uri",
          "s3Bucket": "example-s3-bucket",
          "s3Key": "example-s3-key",
          "s3ObjectStorageMode": "COPY",
          "s3ObjectVersion": "example-s3-object-version",
          "sourceKMSKeyArn": "example-source-k-m-s-key-arn",
          "zipFile": "example-zip-file"
        },
        "role": "example-role"
      }
    }
  }
}
```
