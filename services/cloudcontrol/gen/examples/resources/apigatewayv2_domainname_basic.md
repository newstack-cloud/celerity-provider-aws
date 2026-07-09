A basic AWS ApiGatewayV2 DomainName with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource domainName: aws/apigatewayv2/domainName {
    metadata {
        displayName = "AWS ApiGatewayV2 DomainName basic"
    }
    spec {
        domainName = "example-domain-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    domainName:
        type: aws/apigatewayv2/domainName
        metadata:
            displayName: AWS ApiGatewayV2 DomainName basic
        spec:
            domainName: example-domain-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "domainName": {
      "type": "aws/apigatewayv2/domainName",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 DomainName basic"
      },
      "spec": {
        "domainName": "example-domain-name"
      }
    }
  }
}
```
