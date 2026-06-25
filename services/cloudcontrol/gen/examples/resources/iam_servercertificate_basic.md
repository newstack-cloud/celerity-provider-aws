A basic AWS IAM ServerCertificate with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource serverCertificate: aws/iam/serverCertificate {
    metadata {
        displayName = "AWS IAM ServerCertificate basic"
    }
    spec {
        serverCertificateName = "example-server-certificate-name"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    serverCertificate:
        type: aws/iam/serverCertificate
        metadata:
            displayName: AWS IAM ServerCertificate basic
        spec:
            serverCertificateName: example-server-certificate-name
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "serverCertificate": {
      "type": "aws/iam/serverCertificate",
      "metadata": {
        "displayName": "AWS IAM ServerCertificate basic"
      },
      "spec": {
        "serverCertificateName": "example-server-certificate-name"
      }
    }
  }
}
```
