Look up an existing SSM parameter by name and export its value and ARN.

```blueprintlang
version "2025-11-02"

data dbHost: aws/ssm/parameter {
    filter "name" == "/my-app/prod/db-host"

    export value: string
    export arn: string
}

export dbHostValue: string {
    field = datasources.dbHost.value
}
```

```yaml
version: 2025-11-02

datasources:
  dbHost:
    type: aws/ssm/parameter
    filter:
      field: name
      operator: "="
      search: /my-app/prod/db-host
    exports:
      value:
        type: string
      arn:
        type: string

exports:
  dbHostValue:
    type: string
    field: datasources.dbHost.value
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "dbHost": {
      "type": "aws/ssm/parameter",
      "filter": {
        "field": "name",
        "operator": "=",
        "search": "/my-app/prod/db-host"
      },
      "exports": {
        "value": {
          "type": "string"
        },
        "arn": {
          "type": "string"
        }
      }
    }
  },
  "exports": {
    "dbHostValue": {
      "type": "string",
      "field": "datasources.dbHost.value"
    }
  }
}
```
