A complete IAM server certificate with multiple tags and a custom path.

```blueprintlang
version "2025-11-02"

resource cloudFrontCert: aws/iam/serverCertificate {
    metadata {
        displayName = "CloudFront Certificate"
        description = "Complete example of an IAM server certificate with multiple tags"
    }
    spec {
        serverCertificateName = "CloudFrontCert"
        certificateBody = "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
        privateKey = "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"
        certificateChain = "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
        path = "/cloudfront/prod/"
        tags = [
            {
                key = "Environment"
                value = "Production"
            },
            {
                key = "Department"
                value = "Engineering"
            },
            {
                key = "ManagedBy"
                value = "Automation"
            }
        ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  cloudFrontCert:
    type: aws/iam/serverCertificate
    metadata:
      displayName: CloudFront Certificate
      description: Complete example of an IAM server certificate with multiple tags
    spec:
      serverCertificateName: CloudFrontCert
      certificateBody: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
      privateKey: |
        -----BEGIN RSA PRIVATE KEY-----
        ...
        -----END RSA PRIVATE KEY-----
      certificateChain: |
        -----BEGIN CERTIFICATE-----
        ...
        -----END CERTIFICATE-----
      path: /cloudfront/prod/
      tags:
        - key: Environment
          value: Production
        - key: Department
          value: Engineering
        - key: ManagedBy
          value: Automation
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "cloudFrontCert": {
      "type": "aws/iam/serverCertificate",
      "metadata": {
        "displayName": "CloudFront Certificate",
        "description": "Complete example of an IAM server certificate with multiple tags"
      },
      "spec": {
        "serverCertificateName": "CloudFrontCert",
        "certificateBody": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
        "privateKey": "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
        "certificateChain": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
        "path": "/cloudfront/prod/",
        "tags": [
          {
            "key": "Environment",
            "value": "Production"
          },
          {
            "key": "Department",
            "value": "Engineering"
          },
          {
            "key": "ManagedBy",
            "value": "Automation"
          }
        ]
      }
    }
  }
}
```
