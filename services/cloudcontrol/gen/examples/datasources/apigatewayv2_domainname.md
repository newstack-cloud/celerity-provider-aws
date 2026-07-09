Look up an existing AWS ApiGatewayV2 DomainName by domainName and export its domainName.

```blueprintlang
version "2025-11-02"

data exampleDomainName: aws/apigatewayv2/domainName {
    filter "domainName" == "example-domainname"

    export domainName: string
    export domainNameArn: string
    export regionalDomainName: string
}

export exampleDomainNameDomainName: string {
    field = datasources.exampleDomainName.domainName
}
```

```yaml
version: 2025-11-02

datasources:
  exampleDomainName:
    type: aws/apigatewayv2/domainName
    filter:
      field: domainName
      operator: "="
      search: example-domainname
    exports:
      domainName:
        type: string
      domainNameArn:
        type: string
      regionalDomainName:
        type: string

exports:
  exampleDomainNameDomainName:
    type: string
    field: datasources.exampleDomainName.domainName
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleDomainName": {
      "type": "aws/apigatewayv2/domainName",
      "filter": { "field": "domainName", "operator": "=", "search": "example-domainname" },
      "exports": {
        "domainName": { "type": "string" },
        "domainNameArn": { "type": "string" },
        "regionalDomainName": { "type": "string" }
      }
    }
  }
}
```
