A basic IAM server certificate with a single tag.

```blueprintlang
version "2025-11-02"

resource myServerCertificate: aws/iam/serverCertificate {
    metadata {
        displayName = "My Server Certificate"
        description = "Basic example of an IAM server certificate"
    }
    spec {
        serverCertificateName = "MyServerCertificate"
        certificateBody = "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
        privateKey = "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"
        certificateChain = "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
        path = "/"
        tags = [
            {
                key = "Environment"
                value = "Production"
            }
        ]
    }
}
```

```yaml
version: 2025-11-02

resources:
  myServerCertificate:
    type: aws/iam/serverCertificate
    metadata:
      displayName: My Server Certificate
      description: Basic example of an IAM server certificate
    spec:
      serverCertificateName: MyServerCertificate
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
      path: /
      tags:
        - key: Environment
          value: Production
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "myServerCertificate": {
      "type": "aws/iam/serverCertificate",
      "metadata": {
        "displayName": "My Server Certificate",
        "description": "Basic example of an IAM server certificate"
      },
      "spec": {
        "serverCertificateName": "MyServerCertificate",
        "certificateBody": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
        "privateKey": "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
        "certificateChain": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
        "path": "/",
        "tags": [
          {
            "key": "Environment",
            "value": "Production"
          }
        ]
      }
    }
  }
}
```
