A basic AWS DynamoDB Table with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource table: aws/dynamodb/table {
    metadata {
        displayName = "AWS DynamoDB Table basic"
    }
    spec {
        keySchema = [
            {
                attributeName = "example-attribute-name",
                keyType = "example-key-type"
            }
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    table:
        type: aws/dynamodb/table
        metadata:
            displayName: AWS DynamoDB Table basic
        spec:
            keySchema:
                - attributeName: example-attribute-name
                  keyType: example-key-type
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "table": {
      "type": "aws/dynamodb/table",
      "metadata": {
        "displayName": "AWS DynamoDB Table basic"
      },
      "spec": {
        "keySchema": [
          {
            "attributeName": "example-attribute-name",
            "keyType": "example-key-type"
          }
        ]
      }
    }
  }
}
```
