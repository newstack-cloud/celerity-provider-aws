A AWS IAM SAMLProvider configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource sAMLProvider: aws/iam/samlProvider {
    metadata {
        displayName = "AWS IAM SAMLProvider complete"
    }
    spec {
        addPrivateKey = "example-add-private-key"
        assertionEncryptionMode = "Allowed"
        name = "example-name"
        privateKeyList = [
            {
                keyId = "example-key-id",
                timestamp = "example-timestamp"
            }
        ]
        removePrivateKey = "example-remove-private-key"
        samlMetadataDocument = "example-saml-metadata-document"
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    sAMLProvider:
        type: aws/iam/samlProvider
        metadata:
            displayName: AWS IAM SAMLProvider complete
        spec:
            addPrivateKey: example-add-private-key
            assertionEncryptionMode: Allowed
            name: example-name
            privateKeyList:
                - keyId: example-key-id
                  timestamp: example-timestamp
            removePrivateKey: example-remove-private-key
            samlMetadataDocument: example-saml-metadata-document
            tags:
                - key: example-key
                  value: example-value
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "sAMLProvider": {
      "type": "aws/iam/samlProvider",
      "metadata": {
        "displayName": "AWS IAM SAMLProvider complete"
      },
      "spec": {
        "addPrivateKey": "example-add-private-key",
        "assertionEncryptionMode": "Allowed",
        "name": "example-name",
        "privateKeyList": [
          {
            "keyId": "example-key-id",
            "timestamp": "example-timestamp"
          }
        ],
        "removePrivateKey": "example-remove-private-key",
        "samlMetadataDocument": "example-saml-metadata-document",
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ]
      }
    }
  }
}
```
