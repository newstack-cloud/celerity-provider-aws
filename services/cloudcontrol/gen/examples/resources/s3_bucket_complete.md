A AWS S3 Bucket configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource bucket: aws/s3/bucket {
    metadata {
        displayName = "AWS S3 Bucket complete"
    }
    spec {
        abacStatus = "Enabled"
        accelerateConfiguration = {
            accelerationStatus = "Enabled"
        }
        accessControl = "AuthenticatedRead"
        analyticsConfigurations = [
            {
                id = "example-id",
                prefix = "example-prefix",
                storageClassAnalysis = {
                    dataExport = {
                        destination = {
                            bucketAccountId = "example-bucket-account-id",
                            bucketArn = "example-bucket-arn",
                            format = "CSV",
                            prefix = "example-prefix"
                        },
                        outputSchemaVersion = "example-output-schema-version"
                    }
                },
                tagFilters = [
                    {
                        key = "example-key",
                        value = "example-value"
                    }
                ]
            }
        ]
        bucketEncryption = {
            serverSideEncryptionConfiguration = [
                {
                    blockedEncryptionTypes = {
                        encryptionType = [
                            "NONE"
                        ]
                    },
                    bucketKeyEnabled = false,
                    serverSideEncryptionByDefault = {
                        kmsMasterKeyID = "example-kms-master-key-i-d",
                        sseAlgorithm = "aws:kms"
                    }
                }
            ]
        }
        bucketName = "example-bucket-name"
        bucketNamePrefix = "example-bucket-name-prefix"
        bucketNamespace = "global"
        corsConfiguration = {
            corsRules = [
                {
                    allowedHeaders = [
                        "example-allowed-header"
                    ],
                    allowedMethods = [
                        "GET"
                    ],
                    allowedOrigins = [
                        "example-allowed-origin"
                    ],
                    exposedHeaders = [
                        "example-exposed-header"
                    ],
                    id = "example-id",
                    maxAge = 0
                }
            ]
        }
        intelligentTieringConfigurations = [
            {
                id = "example-id",
                prefix = "example-prefix",
                status = "Disabled",
                tagFilters = [
                    {
                        key = "example-key",
                        value = "example-value"
                    }
                ],
                tierings = [
                    {
                        accessTier = "ARCHIVE_ACCESS",
                        days = 1
                    }
                ]
            }
        ]
        inventoryConfigurations = [
            {
                destination = {
                    bucketAccountId = "example-bucket-account-id",
                    bucketArn = "example-bucket-arn",
                    format = "CSV",
                    prefix = "example-prefix"
                },
                enabled = false,
                id = "example-id",
                includedObjectVersions = "All",
                optionalFields = [
                    "Size"
                ],
                prefix = "example-prefix",
                scheduleFrequency = "Daily"
            }
        ]
        lifecycleConfiguration = {
            rules = [
                {
                    abortIncompleteMultipartUpload = {
                        daysAfterInitiation = 0
                    },
                    expirationDate = "example-expiration-date",
                    expirationInDays = 1,
                    expiredObjectDeleteMarker = false,
                    id = "example-id",
                    noncurrentVersionExpiration = {
                        newerNoncurrentVersions = 1,
                        noncurrentDays = 1
                    },
                    noncurrentVersionExpirationInDays = 1,
                    noncurrentVersionTransition = {
                        newerNoncurrentVersions = 1,
                        storageClass = "DEEP_ARCHIVE",
                        transitionInDays = 1
                    },
                    noncurrentVersionTransitions = [
                        {
                            newerNoncurrentVersions = 1,
                            storageClass = "DEEP_ARCHIVE",
                            transitionInDays = 1
                        }
                    ],
                    objectSizeGreaterThan = "example-object-size-greater-than",
                    objectSizeLessThan = "example-object-size-less-than",
                    prefix = "example-prefix",
                    status = "Enabled",
                    tagFilters = [
                        {
                            key = "example-key",
                            value = "example-value"
                        }
                    ],
                    transition = {
                        storageClass = "DEEP_ARCHIVE",
                        transitionDate = "example-transition-date",
                        transitionInDays = 1
                    },
                    transitions = [
                        {
                            storageClass = "DEEP_ARCHIVE",
                            transitionDate = "example-transition-date",
                            transitionInDays = 1
                        }
                    ]
                }
            ],
            transitionDefaultMinimumObjectSize = "varies_by_storage_class"
        }
        loggingConfiguration = {
            destinationBucketName = "example-destination-bucket-name",
            logFilePrefix = "example-log-file-prefix",
            targetObjectKeyFormat = {
                simplePrefix = {
                    exampleKey = "example-value"
                }
            }
        }
        metadataConfiguration = {
            annotationTableConfiguration = {
                configurationState = "ENABLED",
                encryptionConfiguration = {
                    kmsKeyArn = "example-kms-key-arn",
                    sseAlgorithm = "aws:kms"
                },
                role = "example-role"
            },
            inventoryTableConfiguration = {
                configurationState = "ENABLED",
                encryptionConfiguration = {
                    kmsKeyArn = "example-kms-key-arn",
                    sseAlgorithm = "aws:kms"
                }
            },
            journalTableConfiguration = {
                encryptionConfiguration = {
                    kmsKeyArn = "example-kms-key-arn",
                    sseAlgorithm = "aws:kms"
                },
                recordExpiration = {
                    days = 1,
                    expiration = "ENABLED"
                }
            }
        }
        metadataTableConfiguration = {
            s3TablesDestination = {
                tableBucketArn = "example-table-bucket-arn",
                tableName = "example-table-name"
            }
        }
        metricsConfigurations = [
            {
                accessPointArn = "example-access-point-arn",
                id = "example-id",
                prefix = "example-prefix",
                tagFilters = [
                    {
                        key = "example-key",
                        value = "example-value"
                    }
                ]
            }
        ]
        notificationConfiguration = {
            eventBridgeConfiguration = {
                eventBridgeEnabled = false
            },
            lambdaConfigurations = [
                {
                    event = "example-event",
                    filter = {
                        s3Key = {
                            rules = [
                                {
                                    name = "example-name",
                                    value = "example-value"
                                }
                            ]
                        }
                    },
                    function = "example-function"
                }
            ],
            queueConfigurations = [
                {
                    event = "example-event",
                    filter = {
                        s3Key = {
                            rules = [
                                {
                                    name = "example-name",
                                    value = "example-value"
                                }
                            ]
                        }
                    },
                    queue = "example-queue"
                }
            ],
            topicConfigurations = [
                {
                    event = "example-event",
                    filter = {
                        s3Key = {
                            rules = [
                                {
                                    name = "example-name",
                                    value = "example-value"
                                }
                            ]
                        }
                    },
                    topic = "example-topic"
                }
            ]
        }
        objectLockConfiguration = {
            objectLockEnabled = "example-object-lock-enabled",
            rule = {
                defaultRetention = {
                    days = 1,
                    mode = "COMPLIANCE",
                    years = 1
                }
            }
        }
        objectLockEnabled = false
        ownershipControls = {
            rules = [
                {
                    objectOwnership = "ObjectWriter"
                }
            ]
        }
        publicAccessBlockConfiguration = {
            blockPublicAcls = false,
            blockPublicPolicy = false,
            ignorePublicAcls = false,
            restrictPublicBuckets = false
        }
        replicationConfiguration = {
            role = "example-role",
            rules = [
                {
                    deleteMarkerReplication = {
                        status = "Disabled"
                    },
                    destination = {
                        accessControlTranslation = {
                            owner = "example-owner"
                        },
                        account = "example-account",
                        bucket = "example-bucket",
                        encryptionConfiguration = {
                            replicaKmsKeyID = "example-replica-kms-key-i-d"
                        },
                        metrics = {
                            eventThreshold = {
                                minutes = 1
                            },
                            status = "Disabled"
                        },
                        replicationTime = {
                            status = "Disabled",
                            time = {
                                minutes = 1
                            }
                        },
                        storageClass = "DEEP_ARCHIVE"
                    },
                    filter = {
                        and = {
                            prefix = "example-prefix",
                            tagFilters = [
                                {
                                    key = "example-key",
                                    value = "example-value"
                                }
                            ]
                        },
                        prefix = "example-prefix",
                        tagFilter = {
                            key = "example-key",
                            value = "example-value"
                        }
                    },
                    id = "example-id",
                    prefix = "example-prefix",
                    priority = 1,
                    sourceSelectionCriteria = {
                        replicaModifications = {
                            status = "Enabled"
                        },
                        sseKmsEncryptedObjects = {
                            status = "Disabled"
                        }
                    },
                    status = "Disabled"
                }
            ]
        }
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        versioningConfiguration = {
            status = "Enabled"
        }
        websiteConfiguration = {
            errorDocument = "example-error-document",
            indexDocument = "example-index-document",
            redirectAllRequestsTo = {
                hostName = "example-host-name",
                protocol = "http"
            },
            routingRules = [
                {
                    redirectRule = {
                        hostName = "example-host-name",
                        httpRedirectCode = "example-http-redirect-code",
                        protocol = "http",
                        replaceKeyPrefixWith = "example-replace-key-prefix-with",
                        replaceKeyWith = "example-replace-key-with"
                    },
                    routingRuleCondition = {
                        httpErrorCodeReturnedEquals = "example-http-error-code-returned-equals",
                        keyPrefixEquals = "example-key-prefix-equals"
                    }
                }
            ]
        }
    }
}
```

```yaml
version: "2025-11-02"
resources:
    bucket:
        type: aws/s3/bucket
        metadata:
            displayName: AWS S3 Bucket complete
        spec:
            abacStatus: Enabled
            accelerateConfiguration:
                accelerationStatus: Enabled
            accessControl: AuthenticatedRead
            analyticsConfigurations:
                - id: example-id
                  prefix: example-prefix
                  storageClassAnalysis:
                    dataExport:
                        destination:
                            bucketAccountId: example-bucket-account-id
                            bucketArn: example-bucket-arn
                            format: CSV
                            prefix: example-prefix
                        outputSchemaVersion: example-output-schema-version
                  tagFilters:
                    - key: example-key
                      value: example-value
            bucketEncryption:
                serverSideEncryptionConfiguration:
                    - blockedEncryptionTypes:
                        encryptionType:
                            - NONE
                      bucketKeyEnabled: false
                      serverSideEncryptionByDefault:
                        kmsMasterKeyID: example-kms-master-key-i-d
                        sseAlgorithm: aws:kms
            bucketName: example-bucket-name
            bucketNamePrefix: example-bucket-name-prefix
            bucketNamespace: global
            corsConfiguration:
                corsRules:
                    - allowedHeaders:
                        - example-allowed-header
                      allowedMethods:
                        - GET
                      allowedOrigins:
                        - example-allowed-origin
                      exposedHeaders:
                        - example-exposed-header
                      id: example-id
                      maxAge: 0
            intelligentTieringConfigurations:
                - id: example-id
                  prefix: example-prefix
                  status: Disabled
                  tagFilters:
                    - key: example-key
                      value: example-value
                  tierings:
                    - accessTier: ARCHIVE_ACCESS
                      days: 1
            inventoryConfigurations:
                - destination:
                    bucketAccountId: example-bucket-account-id
                    bucketArn: example-bucket-arn
                    format: CSV
                    prefix: example-prefix
                  enabled: false
                  id: example-id
                  includedObjectVersions: All
                  optionalFields:
                    - Size
                  prefix: example-prefix
                  scheduleFrequency: Daily
            lifecycleConfiguration:
                rules:
                    - abortIncompleteMultipartUpload:
                        daysAfterInitiation: 0
                      expirationDate: example-expiration-date
                      expirationInDays: 1
                      expiredObjectDeleteMarker: false
                      id: example-id
                      noncurrentVersionExpiration:
                        newerNoncurrentVersions: 1
                        noncurrentDays: 1
                      noncurrentVersionExpirationInDays: 1
                      noncurrentVersionTransition:
                        newerNoncurrentVersions: 1
                        storageClass: DEEP_ARCHIVE
                        transitionInDays: 1
                      noncurrentVersionTransitions:
                        - newerNoncurrentVersions: 1
                          storageClass: DEEP_ARCHIVE
                          transitionInDays: 1
                      objectSizeGreaterThan: example-object-size-greater-than
                      objectSizeLessThan: example-object-size-less-than
                      prefix: example-prefix
                      status: Enabled
                      tagFilters:
                        - key: example-key
                          value: example-value
                      transition:
                        storageClass: DEEP_ARCHIVE
                        transitionDate: example-transition-date
                        transitionInDays: 1
                      transitions:
                        - storageClass: DEEP_ARCHIVE
                          transitionDate: example-transition-date
                          transitionInDays: 1
                transitionDefaultMinimumObjectSize: varies_by_storage_class
            loggingConfiguration:
                destinationBucketName: example-destination-bucket-name
                logFilePrefix: example-log-file-prefix
                targetObjectKeyFormat:
                    simplePrefix:
                        exampleKey: example-value
            metadataConfiguration:
                annotationTableConfiguration:
                    configurationState: ENABLED
                    encryptionConfiguration:
                        kmsKeyArn: example-kms-key-arn
                        sseAlgorithm: aws:kms
                    role: example-role
                inventoryTableConfiguration:
                    configurationState: ENABLED
                    encryptionConfiguration:
                        kmsKeyArn: example-kms-key-arn
                        sseAlgorithm: aws:kms
                journalTableConfiguration:
                    encryptionConfiguration:
                        kmsKeyArn: example-kms-key-arn
                        sseAlgorithm: aws:kms
                    recordExpiration:
                        days: 1
                        expiration: ENABLED
            metadataTableConfiguration:
                s3TablesDestination:
                    tableBucketArn: example-table-bucket-arn
                    tableName: example-table-name
            metricsConfigurations:
                - accessPointArn: example-access-point-arn
                  id: example-id
                  prefix: example-prefix
                  tagFilters:
                    - key: example-key
                      value: example-value
            notificationConfiguration:
                eventBridgeConfiguration:
                    eventBridgeEnabled: false
                lambdaConfigurations:
                    - event: example-event
                      filter:
                        s3Key:
                            rules:
                                - name: example-name
                                  value: example-value
                      function: example-function
                queueConfigurations:
                    - event: example-event
                      filter:
                        s3Key:
                            rules:
                                - name: example-name
                                  value: example-value
                      queue: example-queue
                topicConfigurations:
                    - event: example-event
                      filter:
                        s3Key:
                            rules:
                                - name: example-name
                                  value: example-value
                      topic: example-topic
            objectLockConfiguration:
                objectLockEnabled: example-object-lock-enabled
                rule:
                    defaultRetention:
                        days: 1
                        mode: COMPLIANCE
                        years: 1
            objectLockEnabled: false
            ownershipControls:
                rules:
                    - objectOwnership: ObjectWriter
            publicAccessBlockConfiguration:
                blockPublicAcls: false
                blockPublicPolicy: false
                ignorePublicAcls: false
                restrictPublicBuckets: false
            replicationConfiguration:
                role: example-role
                rules:
                    - deleteMarkerReplication:
                        status: Disabled
                      destination:
                        accessControlTranslation:
                            owner: example-owner
                        account: example-account
                        bucket: example-bucket
                        encryptionConfiguration:
                            replicaKmsKeyID: example-replica-kms-key-i-d
                        metrics:
                            eventThreshold:
                                minutes: 1
                            status: Disabled
                        replicationTime:
                            status: Disabled
                            time:
                                minutes: 1
                        storageClass: DEEP_ARCHIVE
                      filter:
                        and:
                            prefix: example-prefix
                            tagFilters:
                                - key: example-key
                                  value: example-value
                        prefix: example-prefix
                        tagFilter:
                            key: example-key
                            value: example-value
                      id: example-id
                      prefix: example-prefix
                      priority: 1
                      sourceSelectionCriteria:
                        replicaModifications:
                            status: Enabled
                        sseKmsEncryptedObjects:
                            status: Disabled
                      status: Disabled
            tags:
                - key: example-key
                  value: example-value
            versioningConfiguration:
                status: Enabled
            websiteConfiguration:
                errorDocument: example-error-document
                indexDocument: example-index-document
                redirectAllRequestsTo:
                    hostName: example-host-name
                    protocol: http
                routingRules:
                    - redirectRule:
                        hostName: example-host-name
                        httpRedirectCode: example-http-redirect-code
                        protocol: http
                        replaceKeyPrefixWith: example-replace-key-prefix-with
                        replaceKeyWith: example-replace-key-with
                      routingRuleCondition:
                        httpErrorCodeReturnedEquals: example-http-error-code-returned-equals
                        keyPrefixEquals: example-key-prefix-equals
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "bucket": {
      "type": "aws/s3/bucket",
      "metadata": {
        "displayName": "AWS S3 Bucket complete"
      },
      "spec": {
        "abacStatus": "Enabled",
        "accelerateConfiguration": {
          "accelerationStatus": "Enabled"
        },
        "accessControl": "AuthenticatedRead",
        "analyticsConfigurations": [
          {
            "id": "example-id",
            "prefix": "example-prefix",
            "storageClassAnalysis": {
              "dataExport": {
                "destination": {
                  "bucketAccountId": "example-bucket-account-id",
                  "bucketArn": "example-bucket-arn",
                  "format": "CSV",
                  "prefix": "example-prefix"
                },
                "outputSchemaVersion": "example-output-schema-version"
              }
            },
            "tagFilters": [
              {
                "key": "example-key",
                "value": "example-value"
              }
            ]
          }
        ],
        "bucketEncryption": {
          "serverSideEncryptionConfiguration": [
            {
              "blockedEncryptionTypes": {
                "encryptionType": [
                  "NONE"
                ]
              },
              "bucketKeyEnabled": false,
              "serverSideEncryptionByDefault": {
                "kmsMasterKeyID": "example-kms-master-key-i-d",
                "sseAlgorithm": "aws:kms"
              }
            }
          ]
        },
        "bucketName": "example-bucket-name",
        "bucketNamePrefix": "example-bucket-name-prefix",
        "bucketNamespace": "global",
        "corsConfiguration": {
          "corsRules": [
            {
              "allowedHeaders": [
                "example-allowed-header"
              ],
              "allowedMethods": [
                "GET"
              ],
              "allowedOrigins": [
                "example-allowed-origin"
              ],
              "exposedHeaders": [
                "example-exposed-header"
              ],
              "id": "example-id",
              "maxAge": 0
            }
          ]
        },
        "intelligentTieringConfigurations": [
          {
            "id": "example-id",
            "prefix": "example-prefix",
            "status": "Disabled",
            "tagFilters": [
              {
                "key": "example-key",
                "value": "example-value"
              }
            ],
            "tierings": [
              {
                "accessTier": "ARCHIVE_ACCESS",
                "days": 1
              }
            ]
          }
        ],
        "inventoryConfigurations": [
          {
            "destination": {
              "bucketAccountId": "example-bucket-account-id",
              "bucketArn": "example-bucket-arn",
              "format": "CSV",
              "prefix": "example-prefix"
            },
            "enabled": false,
            "id": "example-id",
            "includedObjectVersions": "All",
            "optionalFields": [
              "Size"
            ],
            "prefix": "example-prefix",
            "scheduleFrequency": "Daily"
          }
        ],
        "lifecycleConfiguration": {
          "rules": [
            {
              "abortIncompleteMultipartUpload": {
                "daysAfterInitiation": 0
              },
              "expirationDate": "example-expiration-date",
              "expirationInDays": 1,
              "expiredObjectDeleteMarker": false,
              "id": "example-id",
              "noncurrentVersionExpiration": {
                "newerNoncurrentVersions": 1,
                "noncurrentDays": 1
              },
              "noncurrentVersionExpirationInDays": 1,
              "noncurrentVersionTransition": {
                "newerNoncurrentVersions": 1,
                "storageClass": "DEEP_ARCHIVE",
                "transitionInDays": 1
              },
              "noncurrentVersionTransitions": [
                {
                  "newerNoncurrentVersions": 1,
                  "storageClass": "DEEP_ARCHIVE",
                  "transitionInDays": 1
                }
              ],
              "objectSizeGreaterThan": "example-object-size-greater-than",
              "objectSizeLessThan": "example-object-size-less-than",
              "prefix": "example-prefix",
              "status": "Enabled",
              "tagFilters": [
                {
                  "key": "example-key",
                  "value": "example-value"
                }
              ],
              "transition": {
                "storageClass": "DEEP_ARCHIVE",
                "transitionDate": "example-transition-date",
                "transitionInDays": 1
              },
              "transitions": [
                {
                  "storageClass": "DEEP_ARCHIVE",
                  "transitionDate": "example-transition-date",
                  "transitionInDays": 1
                }
              ]
            }
          ],
          "transitionDefaultMinimumObjectSize": "varies_by_storage_class"
        },
        "loggingConfiguration": {
          "destinationBucketName": "example-destination-bucket-name",
          "logFilePrefix": "example-log-file-prefix",
          "targetObjectKeyFormat": {
            "simplePrefix": {
              "exampleKey": "example-value"
            }
          }
        },
        "metadataConfiguration": {
          "annotationTableConfiguration": {
            "configurationState": "ENABLED",
            "encryptionConfiguration": {
              "kmsKeyArn": "example-kms-key-arn",
              "sseAlgorithm": "aws:kms"
            },
            "role": "example-role"
          },
          "inventoryTableConfiguration": {
            "configurationState": "ENABLED",
            "encryptionConfiguration": {
              "kmsKeyArn": "example-kms-key-arn",
              "sseAlgorithm": "aws:kms"
            }
          },
          "journalTableConfiguration": {
            "encryptionConfiguration": {
              "kmsKeyArn": "example-kms-key-arn",
              "sseAlgorithm": "aws:kms"
            },
            "recordExpiration": {
              "days": 1,
              "expiration": "ENABLED"
            }
          }
        },
        "metadataTableConfiguration": {
          "s3TablesDestination": {
            "tableBucketArn": "example-table-bucket-arn",
            "tableName": "example-table-name"
          }
        },
        "metricsConfigurations": [
          {
            "accessPointArn": "example-access-point-arn",
            "id": "example-id",
            "prefix": "example-prefix",
            "tagFilters": [
              {
                "key": "example-key",
                "value": "example-value"
              }
            ]
          }
        ],
        "notificationConfiguration": {
          "eventBridgeConfiguration": {
            "eventBridgeEnabled": false
          },
          "lambdaConfigurations": [
            {
              "event": "example-event",
              "filter": {
                "s3Key": {
                  "rules": [
                    {
                      "name": "example-name",
                      "value": "example-value"
                    }
                  ]
                }
              },
              "function": "example-function"
            }
          ],
          "queueConfigurations": [
            {
              "event": "example-event",
              "filter": {
                "s3Key": {
                  "rules": [
                    {
                      "name": "example-name",
                      "value": "example-value"
                    }
                  ]
                }
              },
              "queue": "example-queue"
            }
          ],
          "topicConfigurations": [
            {
              "event": "example-event",
              "filter": {
                "s3Key": {
                  "rules": [
                    {
                      "name": "example-name",
                      "value": "example-value"
                    }
                  ]
                }
              },
              "topic": "example-topic"
            }
          ]
        },
        "objectLockConfiguration": {
          "objectLockEnabled": "example-object-lock-enabled",
          "rule": {
            "defaultRetention": {
              "days": 1,
              "mode": "COMPLIANCE",
              "years": 1
            }
          }
        },
        "objectLockEnabled": false,
        "ownershipControls": {
          "rules": [
            {
              "objectOwnership": "ObjectWriter"
            }
          ]
        },
        "publicAccessBlockConfiguration": {
          "blockPublicAcls": false,
          "blockPublicPolicy": false,
          "ignorePublicAcls": false,
          "restrictPublicBuckets": false
        },
        "replicationConfiguration": {
          "role": "example-role",
          "rules": [
            {
              "deleteMarkerReplication": {
                "status": "Disabled"
              },
              "destination": {
                "accessControlTranslation": {
                  "owner": "example-owner"
                },
                "account": "example-account",
                "bucket": "example-bucket",
                "encryptionConfiguration": {
                  "replicaKmsKeyID": "example-replica-kms-key-i-d"
                },
                "metrics": {
                  "eventThreshold": {
                    "minutes": 1
                  },
                  "status": "Disabled"
                },
                "replicationTime": {
                  "status": "Disabled",
                  "time": {
                    "minutes": 1
                  }
                },
                "storageClass": "DEEP_ARCHIVE"
              },
              "filter": {
                "and": {
                  "prefix": "example-prefix",
                  "tagFilters": [
                    {
                      "key": "example-key",
                      "value": "example-value"
                    }
                  ]
                },
                "prefix": "example-prefix",
                "tagFilter": {
                  "key": "example-key",
                  "value": "example-value"
                }
              },
              "id": "example-id",
              "prefix": "example-prefix",
              "priority": 1,
              "sourceSelectionCriteria": {
                "replicaModifications": {
                  "status": "Enabled"
                },
                "sseKmsEncryptedObjects": {
                  "status": "Disabled"
                }
              },
              "status": "Disabled"
            }
          ]
        },
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "versioningConfiguration": {
          "status": "Enabled"
        },
        "websiteConfiguration": {
          "errorDocument": "example-error-document",
          "indexDocument": "example-index-document",
          "redirectAllRequestsTo": {
            "hostName": "example-host-name",
            "protocol": "http"
          },
          "routingRules": [
            {
              "redirectRule": {
                "hostName": "example-host-name",
                "httpRedirectCode": "example-http-redirect-code",
                "protocol": "http",
                "replaceKeyPrefixWith": "example-replace-key-prefix-with",
                "replaceKeyWith": "example-replace-key-with"
              },
              "routingRuleCondition": {
                "httpErrorCodeReturnedEquals": "example-http-error-code-returned-equals",
                "keyPrefixEquals": "example-key-prefix-equals"
              }
            }
          ]
        }
      }
    }
  }
}
```
