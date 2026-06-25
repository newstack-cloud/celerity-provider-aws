Look up an existing AWS SQS Queue by queueName and export its arn.

```blueprintlang
version "2025-11-02"

data exampleQueue: aws/sqs/queue {
    filter "queueName" == "example-queuename"

    export arn: string
    export contentBasedDeduplication: boolean
    export deduplicationScope: string
}

export exampleQueueArn: string {
    field = datasources.exampleQueue.arn
}
```

```yaml
version: 2025-11-02

datasources:
  exampleQueue:
    type: aws/sqs/queue
    filter:
      field: queueName
      operator: "="
      search: example-queuename
    exports:
      arn:
        type: string
      contentBasedDeduplication:
        type: boolean
      deduplicationScope:
        type: string

exports:
  exampleQueueArn:
    type: string
    field: datasources.exampleQueue.arn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleQueue": {
      "type": "aws/sqs/queue",
      "filter": { "field": "queueName", "operator": "=", "search": "example-queuename" },
      "exports": {
        "arn": { "type": "string" },
        "contentBasedDeduplication": { "type": "boolean" },
        "deduplicationScope": { "type": "string" }
      }
    }
  }
}
```
