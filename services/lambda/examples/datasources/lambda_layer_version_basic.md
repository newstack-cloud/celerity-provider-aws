Look up an existing Lambda layer version by layer name and version number, exporting its ARN and compatible runtimes.

```blueprintlang
version "2025-11-02"

data pythonUtilsLayer: aws/lambda/layerVersion {
    filter "layerName" == "my-python-utils"
    filter "versionNumber" == "1"

    export layerVersionArn: string
    export version: integer
}

export pythonUtilsLayerArn: string {
    field = datasources.pythonUtilsLayer.layerVersionArn
}
```

```yaml
version: 2025-11-02

datasources:
  pythonUtilsLayer:
    type: aws/lambda/layerVersion
    filter:
      - field: layerName
        operator: "="
        search: my-python-utils
      - field: versionNumber
        operator: "="
        search: 1
    exports:
      layerVersionArn:
        type: string
      version:
        type: integer

exports:
  pythonUtilsLayerArn:
    type: string
    field: datasources.pythonUtilsLayer.layerVersionArn
```

```javascript
{
  "version": "2025-11-02",
  "datasources": {
    "pythonUtilsLayer": {
      "type": "aws/lambda/layerVersion",
      "filter": [
        { "field": "layerName", "operator": "=", "search": "my-python-utils" },
        { "field": "versionNumber", "operator": "=", "search": 1 }
      ],
      "exports": {
        "layerVersionArn": { "type": "string" },
        "version": { "type": "integer" }
      }
    }
  },
  "exports": {
    "pythonUtilsLayerArn": {
      "type": "string",
      "field": "datasources.pythonUtilsLayer.layerVersionArn"
    }
  }
}
```
