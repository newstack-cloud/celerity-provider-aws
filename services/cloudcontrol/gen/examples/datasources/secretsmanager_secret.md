Look up an existing AWS SecretsManager Secret by name and export its description.

```blueprintlang
version "2025-11-02"

data exampleSecret: aws/secretsmanager/secret {
    filter "name" == "example-name"

    export description: string
    export id: string
    export kmsKeyId: string
}

export exampleSecretDescription: string {
    field = datasources.exampleSecret.description
}
```

```yaml
version: 2025-11-02

datasources:
  exampleSecret:
    type: aws/secretsmanager/secret
    filter:
      field: name
      operator: "="
      search: example-name
    exports:
      description:
        type: string
      id:
        type: string
      kmsKeyId:
        type: string

exports:
  exampleSecretDescription:
    type: string
    field: datasources.exampleSecret.description
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleSecret": {
      "type": "aws/secretsmanager/secret",
      "filter": { "field": "name", "operator": "=", "search": "example-name" },
      "exports": {
        "description": { "type": "string" },
        "id": { "type": "string" },
        "kmsKeyId": { "type": "string" }
      }
    }
  }
}
```
