A AWS ApiGatewayV2 DomainName configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource domainName: aws/apigatewayv2/domainName {
    metadata {
        displayName = "AWS ApiGatewayV2 DomainName complete"
    }
    spec {
        domainName = "example-domain-name"
        domainNameConfigurations = [
            {
                certificateArn = "example-certificate-arn",
                certificateName = "example-certificate-name",
                endpointType = "example-endpoint-type",
                ipAddressType = "example-ip-address-type",
                ownershipVerificationCertificateArn = "example-ownership-verification-certificate-arn",
                securityPolicy = "example-security-policy"
            }
        ]
        mutualTlsAuthentication = {
            truststoreUri = "example-truststore-uri",
            truststoreVersion = "example-truststore-version"
        }
        routingMode = "API_MAPPING_ONLY"
        tags = {
            example = "example-tags"
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    domainName:
        type: aws/apigatewayv2/domainName
        metadata:
            displayName: AWS ApiGatewayV2 DomainName complete
        spec:
            domainName: example-domain-name
            domainNameConfigurations:
                - certificateArn: example-certificate-arn
                  certificateName: example-certificate-name
                  endpointType: example-endpoint-type
                  ipAddressType: example-ip-address-type
                  ownershipVerificationCertificateArn: example-ownership-verification-certificate-arn
                  securityPolicy: example-security-policy
            mutualTlsAuthentication:
                truststoreUri: example-truststore-uri
                truststoreVersion: example-truststore-version
            routingMode: API_MAPPING_ONLY
            tags:
                example: example-tags
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "domainName": {
      "type": "aws/apigatewayv2/domainName",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 DomainName complete"
      },
      "spec": {
        "domainName": "example-domain-name",
        "domainNameConfigurations": [
          {
            "certificateArn": "example-certificate-arn",
            "certificateName": "example-certificate-name",
            "endpointType": "example-endpoint-type",
            "ipAddressType": "example-ip-address-type",
            "ownershipVerificationCertificateArn": "example-ownership-verification-certificate-arn",
            "securityPolicy": "example-security-policy"
          }
        ],
        "mutualTlsAuthentication": {
          "truststoreUri": "example-truststore-uri",
          "truststoreVersion": "example-truststore-version"
        },
        "routingMode": "API_MAPPING_ONLY",
        "tags": {
          "example": "example-tags"
        }
      }
    }
  }
}
```
