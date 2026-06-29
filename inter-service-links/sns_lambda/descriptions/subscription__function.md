## SNS Subscription to Lambda Function

Lets an SNS topic invoke a Lambda function.

When an SNS subscription delivers to a function, the function is configured to allow the subscription's topic to invoke it. Messages published to the topic then invoke the function, subject to any filtering set on the subscription.

You don't wire this up explicitly: it takes effect automatically when you create an `aws/sns/subscription` whose `endpoint` points at the function. Configure the subscription including message filtering and more on the resource.

### Example

```blueprintlang
version "2025-11-02"

resource ordersTopic: aws/sns/topic {
    spec {
        topicName = "orders"
    }
}

resource processOrderFunction: aws/lambda/function {
    spec {
        functionName = "process-order"
        # ... other function configuration
    }
}

resource processOrderSubscription: aws/sns/subscription {
    spec {
        topicArn = ordersTopic.spec.topicArn
        protocol = "lambda"
        endpoint = processOrderFunction.spec.arn
        filterPolicy = {
            eventType = ["order.created"]
        }
    }
}
```

