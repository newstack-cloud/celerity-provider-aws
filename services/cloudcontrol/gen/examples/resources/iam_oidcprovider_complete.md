A AWS IAM OIDCProvider configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource oIDCProvider: aws/iam/oidcProvider {
    metadata {
        displayName = "AWS IAM OIDCProvider complete"
    }
    spec {
        clientIdList = [
            "example-client-id-list"
        ]
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        thumbprintList = [
            "example-thumbprint-list"
        ]
        url = "example-url"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    oIDCProvider:
        type: aws/iam/oidcProvider
        metadata:
            displayName: AWS IAM OIDCProvider complete
        spec:
            clientIdList:
                - example-client-id-list
            tags:
                - key: example-key
                  value: example-value
            thumbprintList:
                - example-thumbprint-list
            url: example-url
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "oIDCProvider": {
      "type": "aws/iam/oidcProvider",
      "metadata": {
        "displayName": "AWS IAM OIDCProvider complete"
      },
      "spec": {
        "clientIdList": [
          "example-client-id-list"
        ],
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "thumbprintList": [
          "example-thumbprint-list"
        ],
        "url": "example-url"
      }
    }
  }
}
```
