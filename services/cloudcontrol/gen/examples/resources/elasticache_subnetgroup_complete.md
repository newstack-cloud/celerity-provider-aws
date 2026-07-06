A AWS ElastiCache SubnetGroup configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource subnetGroup: aws/elasticache/subnetGroup {
    metadata {
        displayName = "AWS ElastiCache SubnetGroup complete"
    }
    spec {
        cacheSubnetGroupName = "example-cache-subnet-group-name"
        description = "example-description"
        subnetIds = [
            "example-subnet-id"
        ]
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
    subnetGroup:
        type: aws/elasticache/subnetGroup
        metadata:
            displayName: AWS ElastiCache SubnetGroup complete
        spec:
            cacheSubnetGroupName: example-cache-subnet-group-name
            description: example-description
            subnetIds:
                - example-subnet-id
            tags:
                - key: example-key
                  value: example-value
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "subnetGroup": {
      "type": "aws/elasticache/subnetGroup",
      "metadata": {
        "displayName": "AWS ElastiCache SubnetGroup complete"
      },
      "spec": {
        "cacheSubnetGroupName": "example-cache-subnet-group-name",
        "description": "example-description",
        "subnetIds": [
          "example-subnet-id"
        ],
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
