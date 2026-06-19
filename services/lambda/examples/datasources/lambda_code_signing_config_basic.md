Look up an existing Lambda code signing configuration by ARN and export its details.

```blueprintlang
version "2025-11-02"

data codeSigningConfig: aws/lambda/codeSigningConfig {
    filter "arn" == "arn:aws:lambda:us-east-1:123456789012:code-signing-config:csc-0123456789abcdef0"

    export codeSigningConfigId: string
    export description: string
}

export codeSigningConfigId: string {
    field = datasources.codeSigningConfig.codeSigningConfigId
}
```

```yaml
version: 2025-11-02

datasources:
  codeSigningConfig:
    type: aws/lambda/codeSigningConfig
    filter:
      field: arn
      operator: "="
      search: arn:aws:lambda:us-east-1:123456789012:code-signing-config:csc-0123456789abcdef0
    exports:
      codeSigningConfigId:
        type: string
      description:
        type: string

exports:
  codeSigningConfigId:
    type: string
    field: datasources.codeSigningConfig.codeSigningConfigId
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "codeSigningConfig": {
      "type": "aws/lambda/codeSigningConfig",
      "filter": {
        "field": "arn",
        "operator": "=",
        "search": "arn:aws:lambda:us-east-1:123456789012:code-signing-config:csc-0123456789abcdef0"
      },
      "exports": {
        "codeSigningConfigId": { "type": "string" },
        "description": { "type": "string" }
      }
    }
  },
  "exports": {
    "codeSigningConfigId": {
      "type": "string",
      "field": "datasources.codeSigningConfig.codeSigningConfigId"
    }
  }
}
```
