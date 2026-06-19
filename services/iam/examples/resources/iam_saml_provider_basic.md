A basic IAM SAML provider for corporate SSO.

```blueprintlang
version "2025-11-02"

resource corporateSaml: aws/iam/samlProvider {
    metadata {
        displayName = "Corporate SAML Provider"
    }
    spec {
        name = "CorporateSAML"
        samlMetadataDocument = "<?xml version=\"1.0\"?>\n<EntityDescriptor xmlns=\"urn:oasis:names:tc:SAML:2.0:metadata\" entityID=\"http://www.example.com/saml\">\n  <IDPSSODescriptor protocolSupportEnumeration=\"urn:oasis:names:tc:SAML:2.0:protocol\">\n    <SingleSignOnService Binding=\"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST\" Location=\"https://www.example.com/saml/sso\"/>\n  </IDPSSODescriptor>\n</EntityDescriptor>"
    }
}
```

```yaml
version: 2025-11-02

resources:
  corporateSaml:
    type: aws/iam/samlProvider
    metadata:
      displayName: Corporate SAML Provider
    spec:
      name: CorporateSAML
      samlMetadataDocument: |
        <?xml version="1.0"?>
        <EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
                          entityID="http://www.example.com/saml">
          <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
            <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
                               Location="https://www.example.com/saml/sso"/>
          </IDPSSODescriptor>
        </EntityDescriptor>
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "corporateSaml": {
      "type": "aws/iam/samlProvider",
      "metadata": {
        "displayName": "Corporate SAML Provider"
      },
      "spec": {
        "name": "CorporateSAML",
        "samlMetadataDocument": "<?xml version=\"1.0\"?>\n<EntityDescriptor xmlns=\"urn:oasis:names:tc:SAML:2.0:metadata\" entityID=\"http://www.example.com/saml\">\n  <IDPSSODescriptor protocolSupportEnumeration=\"urn:oasis:names:tc:SAML:2.0:protocol\">\n    <SingleSignOnService Binding=\"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST\" Location=\"https://www.example.com/saml/sso\"/>\n  </IDPSSODescriptor>\n</EntityDescriptor>"
      }
    }
  }
}
```
