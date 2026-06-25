A basic AWS DynamoDB GlobalTable with the minimum configuration.

```blueprintlang
version "2025-11-02"

resource globalTable: aws/dynamodb/globalTable {
    metadata {
        displayName = "AWS DynamoDB GlobalTable basic"
    }
    spec {
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
    }
}
```

```yaml
version: "2025-11-02"
resources:
    globalTable:
        type: aws/dynamodb/globalTable
        metadata:
            displayName: AWS DynamoDB GlobalTable basic
        spec:
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
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "globalTable": {
      "type": "aws/dynamodb/globalTable",
      "metadata": {
        "displayName": "AWS DynamoDB GlobalTable basic"
      },
      "spec": {
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
        ]
      }
    }
  }
}
```
