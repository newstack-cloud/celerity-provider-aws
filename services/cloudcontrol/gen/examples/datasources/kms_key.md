Look up an existing AWS KMS Key by keyId and export its arn.

```blueprintlang
version "2025-11-02"

data exampleKey: aws/kms/key {
    filter "keyId" == "example-keyid"

    export arn: string
    export description: string
    export enableKeyRotation: boolean
}

export exampleKeyArn: string {
    field = datasources.exampleKey.arn
}
```

```yaml
version: 2025-11-02

datasources:
  exampleKey:
    type: aws/kms/key
    filter:
      field: keyId
      operator: "="
      search: example-keyid
    exports:
      arn:
        type: string
      description:
        type: string
      enableKeyRotation:
        type: boolean

exports:
  exampleKeyArn:
    type: string
    field: datasources.exampleKey.arn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleKey": {
      "type": "aws/kms/key",
      "filter": { "field": "keyId", "operator": "=", "search": "example-keyid" },
      "exports": {
        "arn": { "type": "string" },
        "description": { "type": "string" },
        "enableKeyRotation": { "type": "boolean" }
      }
    }
  }
}
```
