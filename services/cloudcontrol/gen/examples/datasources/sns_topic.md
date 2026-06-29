Look up an existing AWS SNS Topic by topicArn and export its archivePolicy.

```blueprintlang
version "2025-11-02"

data exampleTopic: aws/sns/topic {
    filter "topicArn" == "example-topicarn"

    export archivePolicy: string
    export contentBasedDeduplication: boolean
    export dataProtectionPolicy: string
}

export exampleTopicArchivePolicy: string {
    field = datasources.exampleTopic.archivePolicy
}
```

```yaml
version: 2025-11-02

datasources:
  exampleTopic:
    type: aws/sns/topic
    filter:
      field: topicArn
      operator: "="
      search: example-topicarn
    exports:
      archivePolicy:
        type: string
      contentBasedDeduplication:
        type: boolean
      dataProtectionPolicy:
        type: string

exports:
  exampleTopicArchivePolicy:
    type: string
    field: datasources.exampleTopic.archivePolicy
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleTopic": {
      "type": "aws/sns/topic",
      "filter": { "field": "topicArn", "operator": "=", "search": "example-topicarn" },
      "exports": {
        "archivePolicy": { "type": "string" },
        "contentBasedDeduplication": { "type": "boolean" },
        "dataProtectionPolicy": { "type": "string" }
      }
    }
  }
}
```
