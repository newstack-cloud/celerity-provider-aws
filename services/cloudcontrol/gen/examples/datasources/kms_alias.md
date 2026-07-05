Look up an existing AWS KMS Alias by aliasName and export its aliasName.

```blueprintlang
version "2025-11-02"

data exampleAlias: aws/kms/alias {
    filter "aliasName" == "example-aliasname"

    export aliasName: string
    export targetKeyId: string
}

export exampleAliasAliasName: string {
    field = datasources.exampleAlias.aliasName
}
```

```yaml
version: 2025-11-02

datasources:
  exampleAlias:
    type: aws/kms/alias
    filter:
      field: aliasName
      operator: "="
      search: example-aliasname
    exports:
      aliasName:
        type: string
      targetKeyId:
        type: string

exports:
  exampleAliasAliasName:
    type: string
    field: datasources.exampleAlias.aliasName
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleAlias": {
      "type": "aws/kms/alias",
      "filter": { "field": "aliasName", "operator": "=", "search": "example-aliasname" },
      "exports": {
        "aliasName": { "type": "string" },
        "targetKeyId": { "type": "string" }
      }
    }
  }
}
```
