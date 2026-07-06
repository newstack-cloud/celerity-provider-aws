A basic AWS ElastiCache SubnetGroup with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource subnetGroup: aws/elasticache/subnetGroup {
    metadata {
        displayName = "AWS ElastiCache SubnetGroup basic"
    }
    spec {
        description = "example-description"
        subnetIds = [
            "example-subnet-id"
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    subnetGroup:
        type: aws/elasticache/subnetGroup
        metadata:
            displayName: AWS ElastiCache SubnetGroup basic
        spec:
            description: example-description
            subnetIds:
                - example-subnet-id
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "subnetGroup": {
      "type": "aws/elasticache/subnetGroup",
      "metadata": {
        "displayName": "AWS ElastiCache SubnetGroup basic"
      },
      "spec": {
        "description": "example-description",
        "subnetIds": [
          "example-subnet-id"
        ]
      }
    }
  }
}
```
