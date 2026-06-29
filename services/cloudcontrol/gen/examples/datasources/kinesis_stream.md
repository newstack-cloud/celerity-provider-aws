Look up an existing AWS Kinesis Stream by name and export its arn.

```blueprintlang
version "2025-11-02"

data exampleStream: aws/kinesis/stream {
    filter "name" == "example-name"

    export arn: string
    export maxRecordSizeInKiB: integer
    export name: string
}

export exampleStreamArn: string {
    field = datasources.exampleStream.arn
}
```

```yaml
version: 2025-11-02

datasources:
  exampleStream:
    type: aws/kinesis/stream
    filter:
      field: name
      operator: "="
      search: example-name
    exports:
      arn:
        type: string
      maxRecordSizeInKiB:
        type: integer
      name:
        type: string

exports:
  exampleStreamArn:
    type: string
    field: datasources.exampleStream.arn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "exampleStream": {
      "type": "aws/kinesis/stream",
      "filter": { "field": "name", "operator": "=", "search": "example-name" },
      "exports": {
        "arn": { "type": "string" },
        "maxRecordSizeInKiB": { "type": "integer" },
        "name": { "type": "string" }
      }
    }
  }
}
```
