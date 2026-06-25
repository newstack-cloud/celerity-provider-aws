Look up an existing AWS Events Rule by name and export its arn.

```blueprintlang
version "2025-11-02"

data exampleRule: aws/events/rule {
    filter "name" == "example-name"

    export arn: string
    export description: string
    export eventBusName: string
}

export exampleRuleArn: string {
    field = datasources.exampleRule.arn
}
```

```yaml
version: 2025-11-02

datasources:
  exampleRule:
    type: aws/events/rule
    filter:
      field: name
      operator: "="
      search: example-name
    exports:
      arn:
        type: string
      description:
        type: string
      eventBusName:
        type: string

exports:
  exampleRuleArn:
    type: string
    field: datasources.exampleRule.arn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleRule": {
      "type": "aws/events/rule",
      "filter": { "field": "name", "operator": "=", "search": "example-name" },
      "exports": {
        "arn": { "type": "string" },
        "description": { "type": "string" },
        "eventBusName": { "type": "string" }
      }
    }
  }
}
```
