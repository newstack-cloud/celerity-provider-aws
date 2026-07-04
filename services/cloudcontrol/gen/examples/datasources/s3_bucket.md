Look up an existing AWS S3 Bucket by bucketName and export its arn.

```blueprintlang
version "2025-11-02"

data exampleBucket: aws/s3/bucket {
    filter "bucketName" == "example-bucketname"

    export arn: string
    export abacStatus: string
    export bucketName: string
}

export exampleBucketArn: string {
    field = datasources.exampleBucket.arn
}
```

```yaml
version: 2025-11-02

datasources:
  exampleBucket:
    type: aws/s3/bucket
    filter:
      field: bucketName
      operator: "="
      search: example-bucketname
    exports:
      arn:
        type: string
      abacStatus:
        type: string
      bucketName:
        type: string

exports:
  exampleBucketArn:
    type: string
    field: datasources.exampleBucket.arn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleBucket": {
      "type": "aws/s3/bucket",
      "filter": { "field": "bucketName", "operator": "=", "search": "example-bucketname" },
      "exports": {
        "arn": { "type": "string" },
        "abacStatus": { "type": "string" },
        "bucketName": { "type": "string" }
      }
    }
  }
}
```
