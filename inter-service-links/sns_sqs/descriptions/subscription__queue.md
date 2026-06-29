## SNS Subscription to SQS Queue

Lets an SNS topic deliver messages to an SQS queue.

When an SNS subscription delivers to a queue, the queue's access policy is configured to allow the subscription's topic to send messages to it. Messages published to the topic are then delivered to the queue, subject to any filtering set on the subscription.

You don't wire this up explicitly: it takes effect automatically when you create an `aws/sns/subscription` whose `endpoint` points at the queue. Configure the subscription with the protocol, message filtering, raw message delivery and so on the resource.

### Encryption

If the queue is encrypted with a customer managed KMS key, that key's policy must allow SNS (`sns.amazonaws.com`) to use it (`kms:GenerateDataKey` and `kms:Decrypt`).

### Example

```blueprintlang
version "2025-11-02"

resource ordersTopic: aws/sns/topic {
    spec {
        topicName = "orders"
    }
}

resource ordersConsumerQueue: aws/sqs/queue {
    spec {
        queueName = "orders-consumer"
    }
}

resource ordersConsumerSubscription: aws/sns/subscription {
    spec {
        topicArn = ordersTopic.spec.topicArn
        protocol = "sqs"
        endpoint = ordersConsumerQueue.spec.arn
        rawMessageDelivery = true
        filterPolicy = {
            eventType = ["order.created", "order.updated"]
        }
    }
}
```
