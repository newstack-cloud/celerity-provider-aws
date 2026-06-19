This example demonstrates how to create a comprehensive IAM SAML provider with all available configuration options.

```blueprintlang
version "2025-11-02"

resource oktaSaml: aws/iam/samlProvider {
    metadata {
        displayName = "Okta SAML Provider"
        description = "SAML provider for Okta SSO integration"
        labels = {
            provider = "okta"
            environment = "production"
        }
    }
    spec {
        name = "OktaSAMLProvider"
        samlMetadataDocument = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<EntityDescriptor xmlns=\"urn:oasis:names:tc:SAML:2.0:metadata\" entityID=\"http://www.okta.com/exk1fxpisXtQDf6v4357\">\n  <IDPSSODescriptor protocolSupportEnumeration=\"urn:oasis:names:tc:SAML:2.0:protocol\">\n    <KeyDescriptor use=\"signing\">\n      <KeyInfo xmlns=\"http://www.w3.org/2000/09/xmldsig#\">\n        <X509Data>\n          <X509Certificate>MIIDpDCCAoygAwIBAgIGAVs7...</X509Certificate>\n        </X509Data>\n      </KeyInfo>\n    </KeyDescriptor>\n    <NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</NameIDFormat>\n    <SingleSignOnService Binding=\"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST\" Location=\"https://company.okta.com/app/company_awsaccountid_1/exk1fxpisXtQDf6v4357/sso/saml\"/>\n  </IDPSSODescriptor>\n</EntityDescriptor>"
        tags = [
            {
                key = "Environment"
                value = "Production"
            },
            {
                key = "Service"
                value = "SSO"
            },
            {
                key = "Provider"
                value = "Okta"
            },
            {
                key = "ManagedBy"
                value = "Security"
            }
        ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  oktaSaml:
    type: aws/iam/samlProvider
    metadata:
      displayName: Okta SAML Provider
      description: SAML provider for Okta SSO integration
      labels:
        provider: okta
        environment: production
    spec:
      name: OktaSAMLProvider
      samlMetadataDocument: |
        <?xml version="1.0" encoding="UTF-8"?>
        <EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata"
                          entityID="http://www.okta.com/exk1fxpisXtQDf6v4357">
          <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
            <KeyDescriptor use="signing">
              <KeyInfo xmlns="http://www.w3.org/2000/09/xmldsig#">
                <X509Data>
                  <X509Certificate>MIIDpDCCAoygAwIBAgIGAVs7...</X509Certificate>
                </X509Data>
              </KeyInfo>
            </KeyDescriptor>
            <NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</NameIDFormat>
            <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
                               Location="https://company.okta.com/app/company_awsaccountid_1/exk1fxpisXtQDf6v4357/sso/saml"/>
          </IDPSSODescriptor>
        </EntityDescriptor>
      tags:
        - key: Environment
          value: Production
        - key: Service
          value: SSO
        - key: Provider
          value: Okta
        - key: ManagedBy
          value: Security
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "oktaSaml": {
      "type": "aws/iam/samlProvider",
      "metadata": {
        "displayName": "Okta SAML Provider",
        "description": "SAML provider for Okta SSO integration",
        "labels": {
          "provider": "okta",
          "environment": "production"
        }
      },
      "spec": {
        "name": "OktaSAMLProvider",
        "samlMetadataDocument": "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<EntityDescriptor xmlns=\"urn:oasis:names:tc:SAML:2.0:metadata\" entityID=\"http://www.okta.com/exk1fxpisXtQDf6v4357\">\n  <IDPSSODescriptor protocolSupportEnumeration=\"urn:oasis:names:tc:SAML:2.0:protocol\">\n    <KeyDescriptor use=\"signing\">\n      <KeyInfo xmlns=\"http://www.w3.org/2000/09/xmldsig#\">\n        <X509Data>\n          <X509Certificate>MIIDpDCCAoygAwIBAgIGAVs7...</X509Certificate>\n        </X509Data>\n      </KeyInfo>\n    </KeyDescriptor>\n    <NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</NameIDFormat>\n    <SingleSignOnService Binding=\"urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST\" Location=\"https://company.okta.com/app/company_awsaccountid_1/exk1fxpisXtQDf6v4357/sso/saml\"/>\n  </IDPSSODescriptor>\n</EntityDescriptor>",
        "tags": [
          {
            "key": "Environment",
            "value": "Production"
          },
          {
            "key": "Service",
            "value": "SSO"
          },
          {
            "key": "Provider",
            "value": "Okta"
          },
          {
            "key": "ManagedBy",
            "value": "Security"
          }
        ]
      }
    }
  }
}
```
