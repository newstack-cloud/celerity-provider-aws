## Lambda Function to SNS Topic

Lets a Lambda function publish messages to an SNS topic.

When a function links to a topic, the function is granted permission to publish to it, and (by default) an environment variable holding the topic's ARN is added to the function so your code can reference it without hardcoding.

The function's execution role must be defined as a resource in the same blueprint; the link adds the publish permission to it.

If the function runs inside a VPC without internet access, the link also sets up the network access it needs to reach SNS.

### Example

```blueprintlang
version "2025-11-02"

resource publishOrderFunction: aws/lambda/function {
    metadata {
        labels = {
            topic = "orders"
        }
        annotations = {
            "aws.lambda.sns.ordersTopic.envVarName" = "ORDERS_TOPIC"
        }
    }

    select by label {
        topic = "orders"
    }

    spec {
        functionName = "publish-order"
        role = publishOrderFunctionRole.spec.arn
        # ... other function configuration
    }
}

resource ordersTopic: aws/sns/topic {
    metadata {
        labels = {
            topic = "orders"
        }
    }

    spec {
        topicName = "orders"
    }
}

resource publishOrderFunctionRole: aws/iam/role {
    spec {
        name = "publish-order-role"
        # ... role configuration
    }
}
```

