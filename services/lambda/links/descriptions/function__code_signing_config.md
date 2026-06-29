The link type used to link a lambda function to a code signing config.
This will automatically populate the code signing config ARN of the lambda function
for any resources that match the link selector of the lambda function.

**Example**

```blueprintlang
version "2025-11-02"

resource ordersFunction: aws/lambda/function {
    metadata {
        displayName = "Orders Function"
        labels = {
            app = "orders"
        }
    }

    select by label {
        app = "orders"
    }

    spec {
        handler = "index.handler"
        runtime = "nodejs20.x"
        code = {
            s3Bucket = "my-bucket",
            s3Key = "orders.zip"
        }
    }
}

resource ordersCodeSigningConfig: aws/lambda/codeSigningConfig {
    metadata {
        displayName = "Orders Code Signing Config"
        labels = {
            app = "orders"
        }
    }

    spec {
        allowedPublishers = {
            signingProfileVersionArns = ["arn:aws:signer:us-east-1:123456789012:signing-profile/orders-signing-profile"]
        }
    }
}
```
