A AWS DynamoDB Table configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource table: aws/dynamodb/table {
    metadata {
        displayName = "AWS DynamoDB Table complete"
    }
    spec {
        attributeDefinitions = [
            {
                attributeName = "example-attribute-name",
                attributeType = "example-attribute-type"
            }
        ]
        billingMode = "PAY_PER_REQUEST"
        contributorInsightsSpecification = {
            enabled = false,
            mode = "ACCESSED_AND_THROTTLED_KEYS"
        }
        deletionProtectionEnabled = false
        globalSecondaryIndexes = [
            {
                contributorInsightsSpecification = {
                    enabled = false,
                    mode = "ACCESSED_AND_THROTTLED_KEYS"
                },
                indexName = "example-index-name",
                keySchema = [
                    {
                        attributeName = "example-attribute-name",
                        keyType = "example-key-type"
                    }
                ],
                onDemandThroughput = {
                    maxReadRequestUnits = 1,
                    maxWriteRequestUnits = 1
                },
                projection = {
                    nonKeyAttributes = [
                        "example-non-key-attribute"
                    ],
                    projectionType = "example-projection-type"
                },
                provisionedThroughput = {
                    readCapacityUnits = 1,
                    writeCapacityUnits = 1
                },
                warmThroughput = "example-warm-throughput"
            }
        ]
        importSourceSpecification = {
            inputCompressionType = "example-input-compression-type",
            inputFormat = "example-input-format",
            inputFormatOptions = {
                csv = {
                    delimiter = "example-delimiter",
                    headerList = [
                        "example-header-list"
                    ]
                }
            },
            s3BucketSource = {
                s3Bucket = "example-s3-bucket",
                s3BucketOwner = "example-s3-bucket-owner",
                s3KeyPrefix = "example-s3-key-prefix"
            }
        }
        keySchema = [
            {
                attributeName = "example-attribute-name",
                keyType = "example-key-type"
            }
        ]
        kinesisStreamSpecification = {
            approximateCreationDateTimePrecision = "MICROSECOND",
            streamArn = "example-stream-arn"
        }
        localSecondaryIndexes = [
            {
                indexName = "example-index-name",
                keySchema = [
                    {
                        attributeName = "example-attribute-name",
                        keyType = "example-key-type"
                    }
                ],
                projection = {
                    nonKeyAttributes = [
                        "example-non-key-attribute"
                    ],
                    projectionType = "example-projection-type"
                }
            }
        ]
        onDemandThroughput = {
            maxReadRequestUnits = 1,
            maxWriteRequestUnits = 1
        }
        pointInTimeRecoverySpecification = {
            pointInTimeRecoveryEnabled = false,
            recoveryPeriodInDays = 1
        }
        provisionedThroughput = {
            readCapacityUnits = 1,
            writeCapacityUnits = 1
        }
        resourcePolicy = {
            policyDocument = {
                statement = [
                    {
                        action = [
                            "s3:GetObject"
                        ],
                        effect = "Allow",
                        resource = "arn:aws:s3:::example-bucket/*"
                    }
                ],
                version = "2012-10-17"
            }
        }
        sseSpecification = {
            kmsMasterKeyId = "example-kms-master-key-id",
            sseEnabled = false,
            sseType = "example-sse-type"
        }
        streamSpecification = {
            resourcePolicy = {
                policyDocument = {
                    statement = [
                        {
                            action = [
                                "s3:GetObject"
                            ],
                            effect = "Allow",
                            resource = "arn:aws:s3:::example-bucket/*"
                        }
                    ],
                    version = "2012-10-17"
                }
            },
            streamViewType = "example-stream-view-type"
        }
        tableClass = "example-table-class"
        tableName = "orders"
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        timeToLiveSpecification = {
            attributeName = "example-attribute-name",
            enabled = false
        }
        warmThroughput = "example-warm-throughput"
    }
}
```

```yaml
version: "2025-11-02"
resources:
    table:
        type: aws/dynamodb/table
        metadata:
            displayName: AWS DynamoDB Table complete
        spec:
            attributeDefinitions:
                - attributeName: example-attribute-name
                  attributeType: example-attribute-type
            billingMode: PAY_PER_REQUEST
            contributorInsightsSpecification:
                enabled: false
                mode: ACCESSED_AND_THROTTLED_KEYS
            deletionProtectionEnabled: false
            globalSecondaryIndexes:
                - contributorInsightsSpecification:
                    enabled: false
                    mode: ACCESSED_AND_THROTTLED_KEYS
                  indexName: example-index-name
                  keySchema:
                    - attributeName: example-attribute-name
                      keyType: example-key-type
                  onDemandThroughput:
                    maxReadRequestUnits: 1
                    maxWriteRequestUnits: 1
                  projection:
                    nonKeyAttributes:
                        - example-non-key-attribute
                    projectionType: example-projection-type
                  provisionedThroughput:
                    readCapacityUnits: 1
                    writeCapacityUnits: 1
                  warmThroughput: example-warm-throughput
            importSourceSpecification:
                inputCompressionType: example-input-compression-type
                inputFormat: example-input-format
                inputFormatOptions:
                    csv:
                        delimiter: example-delimiter
                        headerList:
                            - example-header-list
                s3BucketSource:
                    s3Bucket: example-s3-bucket
                    s3BucketOwner: example-s3-bucket-owner
                    s3KeyPrefix: example-s3-key-prefix
            keySchema:
                - attributeName: example-attribute-name
                  keyType: example-key-type
            kinesisStreamSpecification:
                approximateCreationDateTimePrecision: MICROSECOND
                streamArn: example-stream-arn
            localSecondaryIndexes:
                - indexName: example-index-name
                  keySchema:
                    - attributeName: example-attribute-name
                      keyType: example-key-type
                  projection:
                    nonKeyAttributes:
                        - example-non-key-attribute
                    projectionType: example-projection-type
            onDemandThroughput:
                maxReadRequestUnits: 1
                maxWriteRequestUnits: 1
            pointInTimeRecoverySpecification:
                pointInTimeRecoveryEnabled: false
                recoveryPeriodInDays: 1
            provisionedThroughput:
                readCapacityUnits: 1
                writeCapacityUnits: 1
            resourcePolicy:
                policyDocument:
                    statement:
                        - action:
                            - s3:GetObject
                          effect: Allow
                          resource: arn:aws:s3:::example-bucket/*
                    version: "2012-10-17"
            sseSpecification:
                kmsMasterKeyId: example-kms-master-key-id
                sseEnabled: false
                sseType: example-sse-type
            streamSpecification:
                resourcePolicy:
                    policyDocument:
                        statement:
                            - action:
                                - s3:GetObject
                              effect: Allow
                              resource: arn:aws:s3:::example-bucket/*
                        version: "2012-10-17"
                streamViewType: example-stream-view-type
            tableClass: example-table-class
            tableName: orders
            tags:
                - key: example-key
                  value: example-value
            timeToLiveSpecification:
                attributeName: example-attribute-name
                enabled: false
            warmThroughput: example-warm-throughput
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "table": {
      "type": "aws/dynamodb/table",
      "metadata": {
        "displayName": "AWS DynamoDB Table complete"
      },
      "spec": {
        "attributeDefinitions": [
          {
            "attributeName": "example-attribute-name",
            "attributeType": "example-attribute-type"
          }
        ],
        "billingMode": "PAY_PER_REQUEST",
        "contributorInsightsSpecification": {
          "enabled": false,
          "mode": "ACCESSED_AND_THROTTLED_KEYS"
        },
        "deletionProtectionEnabled": false,
        "globalSecondaryIndexes": [
          {
            "contributorInsightsSpecification": {
              "enabled": false,
              "mode": "ACCESSED_AND_THROTTLED_KEYS"
            },
            "indexName": "example-index-name",
            "keySchema": [
              {
                "attributeName": "example-attribute-name",
                "keyType": "example-key-type"
              }
            ],
            "onDemandThroughput": {
              "maxReadRequestUnits": 1,
              "maxWriteRequestUnits": 1
            },
            "projection": {
              "nonKeyAttributes": [
                "example-non-key-attribute"
              ],
              "projectionType": "example-projection-type"
            },
            "provisionedThroughput": {
              "readCapacityUnits": 1,
              "writeCapacityUnits": 1
            },
            "warmThroughput": "example-warm-throughput"
          }
        ],
        "importSourceSpecification": {
          "inputCompressionType": "example-input-compression-type",
          "inputFormat": "example-input-format",
          "inputFormatOptions": {
            "csv": {
              "delimiter": "example-delimiter",
              "headerList": [
                "example-header-list"
              ]
            }
          },
          "s3BucketSource": {
            "s3Bucket": "example-s3-bucket",
            "s3BucketOwner": "example-s3-bucket-owner",
            "s3KeyPrefix": "example-s3-key-prefix"
          }
        },
        "keySchema": [
          {
            "attributeName": "example-attribute-name",
            "keyType": "example-key-type"
          }
        ],
        "kinesisStreamSpecification": {
          "approximateCreationDateTimePrecision": "MICROSECOND",
          "streamArn": "example-stream-arn"
        },
        "localSecondaryIndexes": [
          {
            "indexName": "example-index-name",
            "keySchema": [
              {
                "attributeName": "example-attribute-name",
                "keyType": "example-key-type"
              }
            ],
            "projection": {
              "nonKeyAttributes": [
                "example-non-key-attribute"
              ],
              "projectionType": "example-projection-type"
            }
          }
        ],
        "onDemandThroughput": {
          "maxReadRequestUnits": 1,
          "maxWriteRequestUnits": 1
        },
        "pointInTimeRecoverySpecification": {
          "pointInTimeRecoveryEnabled": false,
          "recoveryPeriodInDays": 1
        },
        "provisionedThroughput": {
          "readCapacityUnits": 1,
          "writeCapacityUnits": 1
        },
        "resourcePolicy": {
          "policyDocument": {
            "statement": [
              {
                "action": [
                  "s3:GetObject"
                ],
                "effect": "Allow",
                "resource": "arn:aws:s3:::example-bucket/*"
              }
            ],
            "version": "2012-10-17"
          }
        },
        "sseSpecification": {
          "kmsMasterKeyId": "example-kms-master-key-id",
          "sseEnabled": false,
          "sseType": "example-sse-type"
        },
        "streamSpecification": {
          "resourcePolicy": {
            "policyDocument": {
              "statement": [
                {
                  "action": [
                    "s3:GetObject"
                  ],
                  "effect": "Allow",
                  "resource": "arn:aws:s3:::example-bucket/*"
                }
              ],
              "version": "2012-10-17"
            }
          },
          "streamViewType": "example-stream-view-type"
        },
        "tableClass": "example-table-class",
        "tableName": "orders",
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "timeToLiveSpecification": {
          "attributeName": "example-attribute-name",
          "enabled": false
        },
        "warmThroughput": "example-warm-throughput"
      }
    }
  }
}
```
