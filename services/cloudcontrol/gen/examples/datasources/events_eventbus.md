Look up an existing AWS Events EventBus by name and export its arn.

```blueprintlang
version "2025-11-02"

data exampleEventBus: aws/events/eventBus {
    filter "name" == "example-name"

    export arn: string
    export description: string
    export kmsKeyIdentifier: string
}

export exampleEventBusArn: string {
    field = datasources.exampleEventBus.arn
}
```

```yaml
version: 2025-11-02

datasources:
  exampleEventBus:
    type: aws/events/eventBus
    filter:
      field: name
      operator: "="
      search: example-name
    exports:
      arn:
        type: string
      description:
        type: string
      kmsKeyIdentifier:
        type: string

exports:
  exampleEventBusArn:
    type: string
    field: datasources.exampleEventBus.arn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleEventBus": {
      "type": "aws/events/eventBus",
      "filter": { "field": "name", "operator": "=", "search": "example-name" },
      "exports": {
        "arn": { "type": "string" },
        "description": { "type": "string" },
        "kmsKeyIdentifier": { "type": "string" }
      }
    }
  }
}
```
