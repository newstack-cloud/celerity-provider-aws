A AWS DynamoDB GlobalTable configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource globalTable: aws/dynamodb/globalTable {
    metadata {
        displayName = "AWS DynamoDB GlobalTable complete"
    }
    spec {
        attributeDefinitions = [
            {
                attributeName = "example-attribute-name",
                attributeType = "example-attribute-type"
            }
        ]
        billingMode = "example-billing-mode"
        globalSecondaryIndexes = [
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
                },
                readOnDemandThroughputSettings = {
                    maxReadRequestUnits = 1
                },
                readProvisionedThroughputSettings = {
                    readCapacityUnits = 1
                },
                warmThroughput = "example-warm-throughput",
                writeOnDemandThroughputSettings = {
                    maxWriteRequestUnits = 1
                },
                writeProvisionedThroughputSettings = {
                    writeCapacityAutoScalingSettings = {
                        maxCapacity = 1,
                        minCapacity = 1,
                        seedCapacity = 1,
                        targetTrackingScalingPolicyConfiguration = {
                            disableScaleIn = false,
                            scaleInCooldown = 0,
                            scaleOutCooldown = 0,
                            targetValue = 1
                        }
                    }
                }
            }
        ]
        globalTableSourceArn = "example-global-table-source-arn"
        globalTableWitnesses = [
            {
                region = "example-region"
            }
        ]
        keySchema = [
            {
                attributeName = "example-attribute-name",
                keyType = "example-key-type"
            }
        ]
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
        multiRegionConsistency = "EVENTUAL"
        readOnDemandThroughputSettings = {
            maxReadRequestUnits = 1
        }
        readProvisionedThroughputSettings = {
            readCapacityUnits = 1
        }
        replicas = [
            {
                contributorInsightsSpecification = {
                    enabled = false,
                    mode = "ACCESSED_AND_THROTTLED_KEYS"
                },
                deletionProtectionEnabled = false,
                globalSecondaryIndexes = [
                    {
                        contributorInsightsSpecification = {
                            enabled = false,
                            mode = "ACCESSED_AND_THROTTLED_KEYS"
                        },
                        indexName = "example-index-name",
                        readOnDemandThroughputSettings = {
                            maxReadRequestUnits = 1
                        },
                        readProvisionedThroughputSettings = {
                            readCapacityAutoScalingSettings = {
                                maxCapacity = 1,
                                minCapacity = 1,
                                seedCapacity = 1,
                                targetTrackingScalingPolicyConfiguration = {
                                    disableScaleIn = false,
                                    scaleInCooldown = 0,
                                    scaleOutCooldown = 0,
                                    targetValue = 1
                                }
                            },
                            readCapacityUnits = 1
                        }
                    }
                ],
                globalTableSettingsReplicationMode = "ENABLED",
                kinesisStreamSpecification = {
                    approximateCreationDateTimePrecision = "MICROSECOND",
                    streamArn = "example-stream-arn"
                },
                pointInTimeRecoverySpecification = {
                    pointInTimeRecoveryEnabled = false,
                    recoveryPeriodInDays = 1
                },
                readOnDemandThroughputSettings = {
                    maxReadRequestUnits = 1
                },
                readProvisionedThroughputSettings = {
                    readCapacityAutoScalingSettings = {
                        maxCapacity = 1,
                        minCapacity = 1,
                        seedCapacity = 1,
                        targetTrackingScalingPolicyConfiguration = {
                            disableScaleIn = false,
                            scaleInCooldown = 0,
                            scaleOutCooldown = 0,
                            targetValue = 1
                        }
                    },
                    readCapacityUnits = 1
                },
                region = "example-region",
                replicaStreamSpecification = {
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
                },
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
                sseSpecification = {
                    kmsMasterKeyId = "example-kms-master-key-id"
                },
                tableClass = "example-table-class",
                tags = [
                    {
                        key = "example-key",
                        value = "example-value"
                    }
                ]
            }
        ]
        sseSpecification = {
            sseEnabled = false,
            sseType = "example-sse-type"
        }
        streamSpecification = {
            streamViewType = "example-stream-view-type"
        }
        tableName = "example-table-name"
        timeToLiveSpecification = {
            attributeName = "example-attribute-name",
            enabled = false
        }
        warmThroughput = "example-warm-throughput"
        writeOnDemandThroughputSettings = {
            maxWriteRequestUnits = 1
        }
        writeProvisionedThroughputSettings = {
            writeCapacityAutoScalingSettings = {
                maxCapacity = 1,
                minCapacity = 1,
                seedCapacity = 1,
                targetTrackingScalingPolicyConfiguration = {
                    disableScaleIn = false,
                    scaleInCooldown = 0,
                    scaleOutCooldown = 0,
                    targetValue = 1
                }
            }
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    globalTable:
        type: aws/dynamodb/globalTable
        metadata:
            displayName: AWS DynamoDB GlobalTable complete
        spec:
            attributeDefinitions:
                - attributeName: example-attribute-name
                  attributeType: example-attribute-type
            billingMode: example-billing-mode
            globalSecondaryIndexes:
                - indexName: example-index-name
                  keySchema:
                    - attributeName: example-attribute-name
                      keyType: example-key-type
                  projection:
                    nonKeyAttributes:
                        - example-non-key-attribute
                    projectionType: example-projection-type
                  readOnDemandThroughputSettings:
                    maxReadRequestUnits: 1
                  readProvisionedThroughputSettings:
                    readCapacityUnits: 1
                  warmThroughput: example-warm-throughput
                  writeOnDemandThroughputSettings:
                    maxWriteRequestUnits: 1
                  writeProvisionedThroughputSettings:
                    writeCapacityAutoScalingSettings:
                        maxCapacity: 1
                        minCapacity: 1
                        seedCapacity: 1
                        targetTrackingScalingPolicyConfiguration:
                            disableScaleIn: false
                            scaleInCooldown: 0
                            scaleOutCooldown: 0
                            targetValue: 1
            globalTableSourceArn: example-global-table-source-arn
            globalTableWitnesses:
                - region: example-region
            keySchema:
                - attributeName: example-attribute-name
                  keyType: example-key-type
            localSecondaryIndexes:
                - indexName: example-index-name
                  keySchema:
                    - attributeName: example-attribute-name
                      keyType: example-key-type
                  projection:
                    nonKeyAttributes:
                        - example-non-key-attribute
                    projectionType: example-projection-type
            multiRegionConsistency: EVENTUAL
            readOnDemandThroughputSettings:
                maxReadRequestUnits: 1
            readProvisionedThroughputSettings:
                readCapacityUnits: 1
            replicas:
                - contributorInsightsSpecification:
                    enabled: false
                    mode: ACCESSED_AND_THROTTLED_KEYS
                  deletionProtectionEnabled: false
                  globalSecondaryIndexes:
                    - contributorInsightsSpecification:
                        enabled: false
                        mode: ACCESSED_AND_THROTTLED_KEYS
                      indexName: example-index-name
                      readOnDemandThroughputSettings:
                        maxReadRequestUnits: 1
                      readProvisionedThroughputSettings:
                        readCapacityAutoScalingSettings:
                            maxCapacity: 1
                            minCapacity: 1
                            seedCapacity: 1
                            targetTrackingScalingPolicyConfiguration:
                                disableScaleIn: false
                                scaleInCooldown: 0
                                scaleOutCooldown: 0
                                targetValue: 1
                        readCapacityUnits: 1
                  globalTableSettingsReplicationMode: ENABLED
                  kinesisStreamSpecification:
                    approximateCreationDateTimePrecision: MICROSECOND
                    streamArn: example-stream-arn
                  pointInTimeRecoverySpecification:
                    pointInTimeRecoveryEnabled: false
                    recoveryPeriodInDays: 1
                  readOnDemandThroughputSettings:
                    maxReadRequestUnits: 1
                  readProvisionedThroughputSettings:
                    readCapacityAutoScalingSettings:
                        maxCapacity: 1
                        minCapacity: 1
                        seedCapacity: 1
                        targetTrackingScalingPolicyConfiguration:
                            disableScaleIn: false
                            scaleInCooldown: 0
                            scaleOutCooldown: 0
                            targetValue: 1
                    readCapacityUnits: 1
                  region: example-region
                  replicaStreamSpecification:
                    resourcePolicy:
                        policyDocument:
                            statement:
                                - action:
                                    - s3:GetObject
                                  effect: Allow
                                  resource: arn:aws:s3:::example-bucket/*
                            version: "2012-10-17"
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
                  tableClass: example-table-class
                  tags:
                    - key: example-key
                      value: example-value
            sseSpecification:
                sseEnabled: false
                sseType: example-sse-type
            streamSpecification:
                streamViewType: example-stream-view-type
            tableName: example-table-name
            timeToLiveSpecification:
                attributeName: example-attribute-name
                enabled: false
            warmThroughput: example-warm-throughput
            writeOnDemandThroughputSettings:
                maxWriteRequestUnits: 1
            writeProvisionedThroughputSettings:
                writeCapacityAutoScalingSettings:
                    maxCapacity: 1
                    minCapacity: 1
                    seedCapacity: 1
                    targetTrackingScalingPolicyConfiguration:
                        disableScaleIn: false
                        scaleInCooldown: 0
                        scaleOutCooldown: 0
                        targetValue: 1
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "globalTable": {
      "type": "aws/dynamodb/globalTable",
      "metadata": {
        "displayName": "AWS DynamoDB GlobalTable complete"
      },
      "spec": {
        "attributeDefinitions": [
          {
            "attributeName": "example-attribute-name",
            "attributeType": "example-attribute-type"
          }
        ],
        "billingMode": "example-billing-mode",
        "globalSecondaryIndexes": [
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
            },
            "readOnDemandThroughputSettings": {
              "maxReadRequestUnits": 1
            },
            "readProvisionedThroughputSettings": {
              "readCapacityUnits": 1
            },
            "warmThroughput": "example-warm-throughput",
            "writeOnDemandThroughputSettings": {
              "maxWriteRequestUnits": 1
            },
            "writeProvisionedThroughputSettings": {
              "writeCapacityAutoScalingSettings": {
                "maxCapacity": 1,
                "minCapacity": 1,
                "seedCapacity": 1,
                "targetTrackingScalingPolicyConfiguration": {
                  "disableScaleIn": false,
                  "scaleInCooldown": 0,
                  "scaleOutCooldown": 0,
                  "targetValue": 1
                }
              }
            }
          }
        ],
        "globalTableSourceArn": "example-global-table-source-arn",
        "globalTableWitnesses": [
          {
            "region": "example-region"
          }
        ],
        "keySchema": [
          {
            "attributeName": "example-attribute-name",
            "keyType": "example-key-type"
          }
        ],
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
        "multiRegionConsistency": "EVENTUAL",
        "readOnDemandThroughputSettings": {
          "maxReadRequestUnits": 1
        },
        "readProvisionedThroughputSettings": {
          "readCapacityUnits": 1
        },
        "replicas": [
          {
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
                "readOnDemandThroughputSettings": {
                  "maxReadRequestUnits": 1
                },
                "readProvisionedThroughputSettings": {
                  "readCapacityAutoScalingSettings": {
                    "maxCapacity": 1,
                    "minCapacity": 1,
                    "seedCapacity": 1,
                    "targetTrackingScalingPolicyConfiguration": {
                      "disableScaleIn": false,
                      "scaleInCooldown": 0,
                      "scaleOutCooldown": 0,
                      "targetValue": 1
                    }
                  },
                  "readCapacityUnits": 1
                }
              }
            ],
            "globalTableSettingsReplicationMode": "ENABLED",
            "kinesisStreamSpecification": {
              "approximateCreationDateTimePrecision": "MICROSECOND",
              "streamArn": "example-stream-arn"
            },
            "pointInTimeRecoverySpecification": {
              "pointInTimeRecoveryEnabled": false,
              "recoveryPeriodInDays": 1
            },
            "readOnDemandThroughputSettings": {
              "maxReadRequestUnits": 1
            },
            "readProvisionedThroughputSettings": {
              "readCapacityAutoScalingSettings": {
                "maxCapacity": 1,
                "minCapacity": 1,
                "seedCapacity": 1,
                "targetTrackingScalingPolicyConfiguration": {
                  "disableScaleIn": false,
                  "scaleInCooldown": 0,
                  "scaleOutCooldown": 0,
                  "targetValue": 1
                }
              },
              "readCapacityUnits": 1
            },
            "region": "example-region",
            "replicaStreamSpecification": {
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
              }
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
              "kmsMasterKeyId": "example-kms-master-key-id"
            },
            "tableClass": "example-table-class",
            "tags": [
              {
                "key": "example-key",
                "value": "example-value"
              }
            ]
          }
        ],
        "sseSpecification": {
          "sseEnabled": false,
          "sseType": "example-sse-type"
        },
        "streamSpecification": {
          "streamViewType": "example-stream-view-type"
        },
        "tableName": "example-table-name",
        "timeToLiveSpecification": {
          "attributeName": "example-attribute-name",
          "enabled": false
        },
        "warmThroughput": "example-warm-throughput",
        "writeOnDemandThroughputSettings": {
          "maxWriteRequestUnits": 1
        },
        "writeProvisionedThroughputSettings": {
          "writeCapacityAutoScalingSettings": {
            "maxCapacity": 1,
            "minCapacity": 1,
            "seedCapacity": 1,
            "targetTrackingScalingPolicyConfiguration": {
              "disableScaleIn": false,
              "scaleInCooldown": 0,
              "scaleOutCooldown": 0,
              "targetValue": 1
            }
          }
        }
      }
    }
  }
}
```
