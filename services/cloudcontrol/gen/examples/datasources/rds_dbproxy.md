Look up an existing AWS RDS DBProxy by dbProxyName and export its dbProxyArn.

```blueprintlang
version "2025-11-02"

data exampleDbProxy: aws/rds/dbProxy {
    filter "dbProxyName" == "example-dbproxyname"

    export dbProxyArn: string
    export dbProxyName: string
    export debugLogging: boolean
}

export exampleDbProxyDbProxyArn: string {
    field = datasources.exampleDbProxy.dbProxyArn
}
```

```yaml
version: 2025-11-02

datasources:
  exampleDbProxy:
    type: aws/rds/dbProxy
    filter:
      field: dbProxyName
      operator: "="
      search: example-dbproxyname
    exports:
      dbProxyArn:
        type: string
      dbProxyName:
        type: string
      debugLogging:
        type: boolean

exports:
  exampleDbProxyDbProxyArn:
    type: string
    field: datasources.exampleDbProxy.dbProxyArn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleDbProxy": {
      "type": "aws/rds/dbProxy",
      "filter": { "field": "dbProxyName", "operator": "=", "search": "example-dbproxyname" },
      "exports": {
        "dbProxyArn": { "type": "string" },
        "dbProxyName": { "type": "string" },
        "debugLogging": { "type": "boolean" }
      }
    }
  }
}
```
