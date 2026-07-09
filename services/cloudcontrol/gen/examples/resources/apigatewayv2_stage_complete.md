A AWS ApiGatewayV2 Stage configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource stage: aws/apigatewayv2/stage {
    metadata {
        displayName = "AWS ApiGatewayV2 Stage complete"
    }
    spec {
        accessLogSettings = {
            destinationArn = "example-destination-arn",
            format = "example-format"
        }
        apiId = "example-api-id"
        autoDeploy = false
        clientCertificateId = "example-client-certificate-id"
        defaultRouteSettings = {
            dataTraceEnabled = false,
            detailedMetricsEnabled = false,
            loggingLevel = "example-logging-level",
            throttlingBurstLimit = 1,
            throttlingRateLimit = 1
        }
        deploymentId = "example-deployment-id"
        description = "example-description"
        routeSettings = {
            exampleKey = "example-value"
        }
        stageName = "example-stage-name"
        stageVariables = {
            exampleKey = "example-value"
        }
        tags = {
            exampleKey = "example-value"
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    stage:
        type: aws/apigatewayv2/stage
        metadata:
            displayName: AWS ApiGatewayV2 Stage complete
        spec:
            accessLogSettings:
                destinationArn: example-destination-arn
                format: example-format
            apiId: example-api-id
            autoDeploy: false
            clientCertificateId: example-client-certificate-id
            defaultRouteSettings:
                dataTraceEnabled: false
                detailedMetricsEnabled: false
                loggingLevel: example-logging-level
                throttlingBurstLimit: 1
                throttlingRateLimit: 1
            deploymentId: example-deployment-id
            description: example-description
            routeSettings:
                exampleKey: example-value
            stageName: example-stage-name
            stageVariables:
                exampleKey: example-value
            tags:
                exampleKey: example-value
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "stage": {
      "type": "aws/apigatewayv2/stage",
      "metadata": {
        "displayName": "AWS ApiGatewayV2 Stage complete"
      },
      "spec": {
        "accessLogSettings": {
          "destinationArn": "example-destination-arn",
          "format": "example-format"
        },
        "apiId": "example-api-id",
        "autoDeploy": false,
        "clientCertificateId": "example-client-certificate-id",
        "defaultRouteSettings": {
          "dataTraceEnabled": false,
          "detailedMetricsEnabled": false,
          "loggingLevel": "example-logging-level",
          "throttlingBurstLimit": 1,
          "throttlingRateLimit": 1
        },
        "deploymentId": "example-deployment-id",
        "description": "example-description",
        "routeSettings": {
          "exampleKey": "example-value"
        },
        "stageName": "example-stage-name",
        "stageVariables": {
          "exampleKey": "example-value"
        },
        "tags": {
          "exampleKey": "example-value"
        }
      }
    }
  }
}
```
