A AWS IAM ServerCertificate configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource serverCertificate: aws/iam/serverCertificate {
    metadata {
        displayName = "AWS IAM ServerCertificate complete"
    }
    spec {
        certificateBody = "example-certificate-body"
        certificateChain = "example-certificate-chain"
        path = "example-path"
        privateKey = "example-private-key"
        serverCertificateName = "example-server-certificate-name"
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
    serverCertificate:
        type: aws/iam/serverCertificate
        metadata:
            displayName: AWS IAM ServerCertificate complete
        spec:
            certificateBody: example-certificate-body
            certificateChain: example-certificate-chain
            path: example-path
            privateKey: example-private-key
            serverCertificateName: example-server-certificate-name
            tags:
                - key: example-key
                  value: example-value
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "serverCertificate": {
      "type": "aws/iam/serverCertificate",
      "metadata": {
        "displayName": "AWS IAM ServerCertificate complete"
      },
      "spec": {
        "certificateBody": "example-certificate-body",
        "certificateChain": "example-certificate-chain",
        "path": "example-path",
        "privateKey": "example-private-key",
        "serverCertificateName": "example-server-certificate-name",
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
