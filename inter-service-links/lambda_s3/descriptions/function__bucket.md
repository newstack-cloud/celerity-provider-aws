## Lambda Function to S3 Bucket

Grants a Lambda function read and/or write access to an S3 bucket.

When a function links to a bucket, the function's execution role is granted access to the bucket and its objects, and (by default) an environment variable holding the bucket's name is added to the function so your code can reference it without manual configuration or hardcoding.

The access level is controlled with the `aws.lambda.s3.<targetBucket>.accessLevel` annotation:

- `read` — `s3:GetObject`, `s3:ListBucket`
- `write` — `s3:PutObject`, `s3:DeleteObject`
- `readwrite` (default) — all of the above

Each statement grants access to both the bucket (`arn:aws:s3:::<bucketName>`) and its objects (`arn:aws:s3:::<bucketName>/*`).

The function's execution role must be defined as a resource in the same blueprint; the link adds the access permission to it.

If the function runs inside a VPC without internet access, the link also activates an S3 gateway VPC endpoint on the linked flex VPC so the function can reach S3 over the private network.

Buckets encrypted with a customer managed KMS key additionally require the key policy to allow the function's execution role (`kms:Decrypt`, and `kms:GenerateDataKey` for writes); this is outside the link's control.

### Example

```blueprintlang
version "2025-11-02"

resource processUploadsFunction: aws/lambda/function {
    metadata {
        labels = {
            bucket = "uploads"
        }
        annotations = {
            "aws.lambda.s3.uploadsBucket.envVarName" = "UPLOADS_BUCKET"
            "aws.lambda.s3.uploadsBucket.accessLevel" = "readwrite"
        }
    }

    select by label {
        bucket = "uploads"
    }

    spec {
        functionName = "process-uploads"
        role = processUploadsFunctionRole.spec.arn
        # ... other function configuration
    }
}

resource uploadsBucket: aws/s3/bucket {
    metadata {
        labels = {
            bucket = "uploads"
        }
    }

    spec {
        bucketName = "my-app-uploads"
    }
}

resource processUploadsFunctionRole: aws/iam/role {
    spec {
        name = "process-uploads-role"
        # ... role configuration
    }
}
```
