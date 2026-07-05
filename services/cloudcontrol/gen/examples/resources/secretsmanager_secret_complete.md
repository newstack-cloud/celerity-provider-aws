A AWS SecretsManager Secret configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource secret: aws/secretsmanager/secret {
    metadata {
        displayName = "AWS SecretsManager Secret complete"
    }
    spec {
        description = "example-description"
        generateSecretString = {
            excludeCharacters = "example-exclude-characters",
            excludeLowercase = false,
            excludeNumbers = false,
            excludePunctuation = false,
            excludeUppercase = false,
            generateStringKey = "example-generate-string-key",
            includeSpace = false,
            passwordLength = 1,
            requireEachIncludedType = false,
            secretStringTemplate = "example-secret-string-template"
        }
        kmsKeyId = "example-kms-key-id"
        name = "example-name"
        replicaRegions = [
            {
                kmsKeyId = "example-kms-key-id",
                region = "example-region"
            }
        ]
        secretString = "example-secret-string"
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        type = "example-type"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    secret:
        type: aws/secretsmanager/secret
        metadata:
            displayName: AWS SecretsManager Secret complete
        spec:
            description: example-description
            generateSecretString:
                excludeCharacters: example-exclude-characters
                excludeLowercase: false
                excludeNumbers: false
                excludePunctuation: false
                excludeUppercase: false
                generateStringKey: example-generate-string-key
                includeSpace: false
                passwordLength: 1
                requireEachIncludedType: false
                secretStringTemplate: example-secret-string-template
            kmsKeyId: example-kms-key-id
            name: example-name
            replicaRegions:
                - kmsKeyId: example-kms-key-id
                  region: example-region
            secretString: example-secret-string
            tags:
                - key: example-key
                  value: example-value
            type: example-type
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "secret": {
      "type": "aws/secretsmanager/secret",
      "metadata": {
        "displayName": "AWS SecretsManager Secret complete"
      },
      "spec": {
        "description": "example-description",
        "generateSecretString": {
          "excludeCharacters": "example-exclude-characters",
          "excludeLowercase": false,
          "excludeNumbers": false,
          "excludePunctuation": false,
          "excludeUppercase": false,
          "generateStringKey": "example-generate-string-key",
          "includeSpace": false,
          "passwordLength": 1,
          "requireEachIncludedType": false,
          "secretStringTemplate": "example-secret-string-template"
        },
        "kmsKeyId": "example-kms-key-id",
        "name": "example-name",
        "replicaRegions": [
          {
            "kmsKeyId": "example-kms-key-id",
            "region": "example-region"
          }
        ],
        "secretString": "example-secret-string",
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "type": "example-type"
      }
    }
  }
}
```
