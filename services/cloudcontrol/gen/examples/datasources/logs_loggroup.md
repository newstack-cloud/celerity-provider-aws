Look up an existing AWS Logs LogGroup by logGroupName and export its arn.

```blueprintlang
version "2025-11-02"

data exampleLogGroup: aws/logs/logGroup {
    filter "logGroupName" == "example-loggroupname"

    export arn: string
    export bearerTokenAuthenticationEnabled: boolean
    export dataProtectionPolicy: string
}

export exampleLogGroupArn: string {
    field = datasources.exampleLogGroup.arn
}
```

```yaml
version: 2025-11-02

datasources:
  exampleLogGroup:
    type: aws/logs/logGroup
    filter:
      field: logGroupName
      operator: "="
      search: example-loggroupname
    exports:
      arn:
        type: string
      bearerTokenAuthenticationEnabled:
        type: boolean
      dataProtectionPolicy:
        type: string

exports:
  exampleLogGroupArn:
    type: string
    field: datasources.exampleLogGroup.arn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleLogGroup": {
      "type": "aws/logs/logGroup",
      "filter": { "field": "logGroupName", "operator": "=", "search": "example-loggroupname" },
      "exports": {
        "arn": { "type": "string" },
        "bearerTokenAuthenticationEnabled": { "type": "boolean" },
        "dataProtectionPolicy": { "type": "string" }
      }
    }
  }
}
```
