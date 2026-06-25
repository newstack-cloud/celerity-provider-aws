A AWS Lambda EventSourceMapping configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource eventSourceMapping: aws/lambda/eventSourceMapping {
    metadata {
        displayName = "AWS Lambda EventSourceMapping complete"
    }
    spec {
        amazonManagedKafkaEventSourceConfig = {
            consumerGroupId = "example-consumer-group-id",
            schemaRegistryConfig = {
                accessConfigs = [
                    {
                        type = "BASIC_AUTH",
                        uri = "example-uri"
                    }
                ],
                eventRecordFormat = "JSON",
                schemaRegistryURI = "example-schema-registry-u-r-i",
                schemaValidationConfigs = [
                    {
                        attribute = "KEY"
                    }
                ]
            }
        }
        batchSize = 1
        bisectBatchOnFunctionError = false
        destinationConfig = {
            onFailure = {
                destination = "example-destination"
            }
        }
        documentDBEventSourceConfig = {
            collectionName = "example-collection-name",
            databaseName = "example-database-name",
            fullDocument = "UpdateLookup"
        }
        enabled = false
        eventSourceArn = "example-event-source-arn"
        filterCriteria = {
            filters = [
                {
                    pattern = "example-pattern"
                }
            ]
        }
        functionName = "example-function-name"
        functionResponseTypes = [
            "ReportBatchItemFailures"
        ]
        kmsKeyArn = "example-kms-key-arn"
        loggingConfig = {
            systemLogLevel = "DEBUG"
        }
        maximumBatchingWindowInSeconds = 0
        maximumRecordAgeInSeconds = -1
        maximumRetryAttempts = -1
        metricsConfig = {
            metrics = [
                "EventCount"
            ]
        }
        parallelizationFactor = 1
        provisionedPollerConfig = {
            maximumPollers = 1,
            minimumPollers = 1,
            pollerGroupName = "example-poller-group-name"
        }
        queues = [
            "example-queue"
        ]
        scalingConfig = {
            maximumConcurrency = 2
        }
        selfManagedEventSource = {
            endpoints = {
                kafkaBootstrapServers = [
                    "example-kafka-bootstrap-server"
                ]
            }
        }
        selfManagedKafkaEventSourceConfig = {
            consumerGroupId = "example-consumer-group-id",
            schemaRegistryConfig = {
                accessConfigs = [
                    {
                        type = "BASIC_AUTH",
                        uri = "example-uri"
                    }
                ],
                eventRecordFormat = "JSON",
                schemaRegistryURI = "example-schema-registry-u-r-i",
                schemaValidationConfigs = [
                    {
                        attribute = "KEY"
                    }
                ]
            }
        }
        sourceAccessConfigurations = [
            {
                type = "BASIC_AUTH",
                uri = "example-uri"
            }
        ]
        startingPosition = "example-starting-position"
        startingPositionTimestamp = 1
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        topics = [
            "example-topic"
        ]
        tumblingWindowInSeconds = 0
    }
}
```

```yaml
version: "2025-11-02"
resources:
    eventSourceMapping:
        type: aws/lambda/eventSourceMapping
        metadata:
            displayName: AWS Lambda EventSourceMapping complete
        spec:
            amazonManagedKafkaEventSourceConfig:
                consumerGroupId: example-consumer-group-id
                schemaRegistryConfig:
                    accessConfigs:
                        - type: BASIC_AUTH
                          uri: example-uri
                    eventRecordFormat: JSON
                    schemaRegistryURI: example-schema-registry-u-r-i
                    schemaValidationConfigs:
                        - attribute: KEY
            batchSize: 1
            bisectBatchOnFunctionError: false
            destinationConfig:
                onFailure:
                    destination: example-destination
            documentDBEventSourceConfig:
                collectionName: example-collection-name
                databaseName: example-database-name
                fullDocument: UpdateLookup
            enabled: false
            eventSourceArn: example-event-source-arn
            filterCriteria:
                filters:
                    - pattern: example-pattern
            functionName: example-function-name
            functionResponseTypes:
                - ReportBatchItemFailures
            kmsKeyArn: example-kms-key-arn
            loggingConfig:
                systemLogLevel: DEBUG
            maximumBatchingWindowInSeconds: 0
            maximumRecordAgeInSeconds: -1
            maximumRetryAttempts: -1
            metricsConfig:
                metrics:
                    - EventCount
            parallelizationFactor: 1
            provisionedPollerConfig:
                maximumPollers: 1
                minimumPollers: 1
                pollerGroupName: example-poller-group-name
            queues:
                - example-queue
            scalingConfig:
                maximumConcurrency: 2
            selfManagedEventSource:
                endpoints:
                    kafkaBootstrapServers:
                        - example-kafka-bootstrap-server
            selfManagedKafkaEventSourceConfig:
                consumerGroupId: example-consumer-group-id
                schemaRegistryConfig:
                    accessConfigs:
                        - type: BASIC_AUTH
                          uri: example-uri
                    eventRecordFormat: JSON
                    schemaRegistryURI: example-schema-registry-u-r-i
                    schemaValidationConfigs:
                        - attribute: KEY
            sourceAccessConfigurations:
                - type: BASIC_AUTH
                  uri: example-uri
            startingPosition: example-starting-position
            startingPositionTimestamp: 1
            tags:
                - key: example-key
                  value: example-value
            topics:
                - example-topic
            tumblingWindowInSeconds: 0
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "eventSourceMapping": {
      "type": "aws/lambda/eventSourceMapping",
      "metadata": {
        "displayName": "AWS Lambda EventSourceMapping complete"
      },
      "spec": {
        "amazonManagedKafkaEventSourceConfig": {
          "consumerGroupId": "example-consumer-group-id",
          "schemaRegistryConfig": {
            "accessConfigs": [
              {
                "type": "BASIC_AUTH",
                "uri": "example-uri"
              }
            ],
            "eventRecordFormat": "JSON",
            "schemaRegistryURI": "example-schema-registry-u-r-i",
            "schemaValidationConfigs": [
              {
                "attribute": "KEY"
              }
            ]
          }
        },
        "batchSize": 1,
        "bisectBatchOnFunctionError": false,
        "destinationConfig": {
          "onFailure": {
            "destination": "example-destination"
          }
        },
        "documentDBEventSourceConfig": {
          "collectionName": "example-collection-name",
          "databaseName": "example-database-name",
          "fullDocument": "UpdateLookup"
        },
        "enabled": false,
        "eventSourceArn": "example-event-source-arn",
        "filterCriteria": {
          "filters": [
            {
              "pattern": "example-pattern"
            }
          ]
        },
        "functionName": "example-function-name",
        "functionResponseTypes": [
          "ReportBatchItemFailures"
        ],
        "kmsKeyArn": "example-kms-key-arn",
        "loggingConfig": {
          "systemLogLevel": "DEBUG"
        },
        "maximumBatchingWindowInSeconds": 0,
        "maximumRecordAgeInSeconds": -1,
        "maximumRetryAttempts": -1,
        "metricsConfig": {
          "metrics": [
            "EventCount"
          ]
        },
        "parallelizationFactor": 1,
        "provisionedPollerConfig": {
          "maximumPollers": 1,
          "minimumPollers": 1,
          "pollerGroupName": "example-poller-group-name"
        },
        "queues": [
          "example-queue"
        ],
        "scalingConfig": {
          "maximumConcurrency": 2
        },
        "selfManagedEventSource": {
          "endpoints": {
            "kafkaBootstrapServers": [
              "example-kafka-bootstrap-server"
            ]
          }
        },
        "selfManagedKafkaEventSourceConfig": {
          "consumerGroupId": "example-consumer-group-id",
          "schemaRegistryConfig": {
            "accessConfigs": [
              {
                "type": "BASIC_AUTH",
                "uri": "example-uri"
              }
            ],
            "eventRecordFormat": "JSON",
            "schemaRegistryURI": "example-schema-registry-u-r-i",
            "schemaValidationConfigs": [
              {
                "attribute": "KEY"
              }
            ]
          }
        },
        "sourceAccessConfigurations": [
          {
            "type": "BASIC_AUTH",
            "uri": "example-uri"
          }
        ],
        "startingPosition": "example-starting-position",
        "startingPositionTimestamp": 1,
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "topics": [
          "example-topic"
        ],
        "tumblingWindowInSeconds": 0
      }
    }
  }
}
```
