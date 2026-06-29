The link type used to link a lambda function to another lambda function where the first function will be configured to be able to invoke and second function.

This will populate permissions and environment variables for the first function to be able to invoke the second function.

Annotations use the `aws.lambda.invoke.*` pattern where "invoke" is the feature being configured.

**Example for all target functions**

```blueprintlang
version "2025-11-02"

resource ordersFunction: aws/lambda/function {
    metadata {
        displayName = "Orders Function"
        # Disable environment variable population which is enabled by default.
        annotations = {
            "aws.lambda.invoke.populateEnvVars" = false
        }
        labels = {
            app = "orders"
        }
    }

    select by label {
        app = "orders"
        system = "global"
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

resource logOrderEventsFunction: aws/lambda/function {
    metadata {
        displayName = "Log Order Events Function"
        labels = {
            app = "orders"
        }
    }

    spec {
        handler = "index.handler"
    }
}

resource logAllEventsFunction: aws/lambda/function {
    metadata {
        displayName = "Log All Events Function"
        labels = {
            system = "global"
        }
    }

    spec {
        handler = "index.handler"
    }
}
```


**Example for a specific target function**

```blueprintlang
version "2025-11-02"

resource ordersFunction: aws/lambda/function {
    metadata {
        displayName = "Orders Function"
        # These annotations will disable environment variable population for all
        # target functions except for the logOrderEventsFunction.
        # The envVarName annotation sets the environment variable name for the
        # logOrderEventsFunction reference in the ordersFunction.
        annotations = {
            "aws.lambda.invoke.populateEnvVars" = false,
            "aws.lambda.invoke.logOrderEventsFunction.populateEnvVars" = true,
            "aws.lambda.invoke.logOrderEventsFunction.envVarName" = "AWS_LAMBDA_FUNCTION_LOG_ORDER_EVENTS"
        }
    }

    select by label {
        app = "orders"
        system = "global"
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

resource logOrderEventsFunction: aws/lambda/function {
    metadata {
        displayName = "Log Order Events Function"
        labels = {
            app = "orders"
        }
    }

    spec {
        handler = "index.handler"
    }
}

resource logAllEventsFunction: aws/lambda/function {
    metadata {
        displayName = "Log All Events Function"
        labels = {
            system = "global"
        }
    }

    spec {
        handler = "index.handler"
    }
}
```
