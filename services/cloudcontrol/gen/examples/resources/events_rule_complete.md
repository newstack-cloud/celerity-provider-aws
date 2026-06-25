A AWS Events Rule configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource rule: aws/events/rule {
    metadata {
        displayName = "AWS Events Rule complete"
    }
    spec {
        description = "example-description"
        eventBusName = "example-event-bus-name"
        eventPattern = {
            source = [
                "com.example.orders"
            ]
        }
        name = "example-name"
        roleArn = "example-role-arn"
        scheduleExpression = "example-schedule-expression"
        state = "DISABLED"
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        targets = [
            {
                appSyncParameters = {
                    graphQLOperation = "example-graph-q-l-operation"
                },
                arn = "example-arn",
                batchParameters = {
                    arrayProperties = {
                        size = 1
                    },
                    jobDefinition = "example-job-definition",
                    jobName = "example-job-name",
                    retryStrategy = {
                        attempts = 1
                    }
                },
                deadLetterConfig = {
                    arn = "example-arn"
                },
                ecsParameters = {
                    capacityProviderStrategy = [
                        {
                            base = 1,
                            capacityProvider = "example-capacity-provider",
                            weight = 1
                        }
                    ],
                    enableECSManagedTags = false,
                    enableExecuteCommand = false,
                    group = "example-group",
                    launchType = "example-launch-type",
                    networkConfiguration = {
                        awsVpcConfiguration = {
                            assignPublicIp = "example-assign-public-ip",
                            securityGroups = [
                                "example-security-group"
                            ],
                            subnets = [
                                "example-subnet"
                            ]
                        }
                    },
                    placementConstraints = [
                        {
                            expression = "example-expression",
                            type = "example-type"
                        }
                    ],
                    placementStrategies = [
                        {
                            field = "example-field",
                            type = "example-type"
                        }
                    ],
                    platformVersion = "example-platform-version",
                    propagateTags = "example-propagate-tags",
                    referenceId = "example-reference-id",
                    tagList = [
                        {
                            key = "example-key",
                            value = "example-value"
                        }
                    ],
                    taskCount = 1,
                    taskDefinitionArn = "example-task-definition-arn"
                },
                httpParameters = {
                    headerParameters = {
                        example = "example-header-parameters"
                    },
                    pathParameterValues = [
                        "example-path-parameter-value"
                    ],
                    queryStringParameters = {
                        example = "example-query-string-parameters"
                    }
                },
                id = "example-id",
                input = "example-input",
                inputPath = "example-input-path",
                inputTransformer = {
                    inputPathsMap = {
                        example = "example-input-paths-map"
                    },
                    inputTemplate = "example-input-template"
                },
                kinesisParameters = {
                    partitionKeyPath = "example-partition-key-path"
                },
                redshiftDataParameters = {
                    database = "example-database",
                    dbUser = "example-db-user",
                    secretManagerArn = "example-secret-manager-arn",
                    sql = "example-sql",
                    sqls = [
                        "example-sql"
                    ],
                    statementName = "example-statement-name",
                    withEvent = false
                },
                retryPolicy = {
                    maximumEventAgeInSeconds = 1,
                    maximumRetryAttempts = 1
                },
                roleArn = "example-role-arn",
                runCommandParameters = {
                    runCommandTargets = [
                        {
                            key = "example-key",
                            values = [
                                "example-value"
                            ]
                        }
                    ]
                },
                sageMakerPipelineParameters = {
                    pipelineParameterList = [
                        {
                            name = "example-name",
                            value = "example-value"
                        }
                    ]
                },
                sqsParameters = {
                    messageGroupId = "example-message-group-id"
                }
            }
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    rule:
        type: aws/events/rule
        metadata:
            displayName: AWS Events Rule complete
        spec:
            description: example-description
            eventBusName: example-event-bus-name
            eventPattern:
                source:
                    - com.example.orders
            name: example-name
            roleArn: example-role-arn
            scheduleExpression: example-schedule-expression
            state: DISABLED
            tags:
                - key: example-key
                  value: example-value
            targets:
                - appSyncParameters:
                    graphQLOperation: example-graph-q-l-operation
                  arn: example-arn
                  batchParameters:
                    arrayProperties:
                        size: 1
                    jobDefinition: example-job-definition
                    jobName: example-job-name
                    retryStrategy:
                        attempts: 1
                  deadLetterConfig:
                    arn: example-arn
                  ecsParameters:
                    capacityProviderStrategy:
                        - base: 1
                          capacityProvider: example-capacity-provider
                          weight: 1
                    enableECSManagedTags: false
                    enableExecuteCommand: false
                    group: example-group
                    launchType: example-launch-type
                    networkConfiguration:
                        awsVpcConfiguration:
                            assignPublicIp: example-assign-public-ip
                            securityGroups:
                                - example-security-group
                            subnets:
                                - example-subnet
                    placementConstraints:
                        - expression: example-expression
                          type: example-type
                    placementStrategies:
                        - field: example-field
                          type: example-type
                    platformVersion: example-platform-version
                    propagateTags: example-propagate-tags
                    referenceId: example-reference-id
                    tagList:
                        - key: example-key
                          value: example-value
                    taskCount: 1
                    taskDefinitionArn: example-task-definition-arn
                  httpParameters:
                    headerParameters:
                        example: example-header-parameters
                    pathParameterValues:
                        - example-path-parameter-value
                    queryStringParameters:
                        example: example-query-string-parameters
                  id: example-id
                  input: example-input
                  inputPath: example-input-path
                  inputTransformer:
                    inputPathsMap:
                        example: example-input-paths-map
                    inputTemplate: example-input-template
                  kinesisParameters:
                    partitionKeyPath: example-partition-key-path
                  redshiftDataParameters:
                    database: example-database
                    dbUser: example-db-user
                    secretManagerArn: example-secret-manager-arn
                    sql: example-sql
                    sqls:
                        - example-sql
                    statementName: example-statement-name
                    withEvent: false
                  retryPolicy:
                    maximumEventAgeInSeconds: 1
                    maximumRetryAttempts: 1
                  roleArn: example-role-arn
                  runCommandParameters:
                    runCommandTargets:
                        - key: example-key
                          values:
                            - example-value
                  sageMakerPipelineParameters:
                    pipelineParameterList:
                        - name: example-name
                          value: example-value
                  sqsParameters:
                    messageGroupId: example-message-group-id
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "rule": {
      "type": "aws/events/rule",
      "metadata": {
        "displayName": "AWS Events Rule complete"
      },
      "spec": {
        "description": "example-description",
        "eventBusName": "example-event-bus-name",
        "eventPattern": {
          "source": [
            "com.example.orders"
          ]
        },
        "name": "example-name",
        "roleArn": "example-role-arn",
        "scheduleExpression": "example-schedule-expression",
        "state": "DISABLED",
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "targets": [
          {
            "appSyncParameters": {
              "graphQLOperation": "example-graph-q-l-operation"
            },
            "arn": "example-arn",
            "batchParameters": {
              "arrayProperties": {
                "size": 1
              },
              "jobDefinition": "example-job-definition",
              "jobName": "example-job-name",
              "retryStrategy": {
                "attempts": 1
              }
            },
            "deadLetterConfig": {
              "arn": "example-arn"
            },
            "ecsParameters": {
              "capacityProviderStrategy": [
                {
                  "base": 1,
                  "capacityProvider": "example-capacity-provider",
                  "weight": 1
                }
              ],
              "enableECSManagedTags": false,
              "enableExecuteCommand": false,
              "group": "example-group",
              "launchType": "example-launch-type",
              "networkConfiguration": {
                "awsVpcConfiguration": {
                  "assignPublicIp": "example-assign-public-ip",
                  "securityGroups": [
                    "example-security-group"
                  ],
                  "subnets": [
                    "example-subnet"
                  ]
                }
              },
              "placementConstraints": [
                {
                  "expression": "example-expression",
                  "type": "example-type"
                }
              ],
              "placementStrategies": [
                {
                  "field": "example-field",
                  "type": "example-type"
                }
              ],
              "platformVersion": "example-platform-version",
              "propagateTags": "example-propagate-tags",
              "referenceId": "example-reference-id",
              "tagList": [
                {
                  "key": "example-key",
                  "value": "example-value"
                }
              ],
              "taskCount": 1,
              "taskDefinitionArn": "example-task-definition-arn"
            },
            "httpParameters": {
              "headerParameters": {
                "example": "example-header-parameters"
              },
              "pathParameterValues": [
                "example-path-parameter-value"
              ],
              "queryStringParameters": {
                "example": "example-query-string-parameters"
              }
            },
            "id": "example-id",
            "input": "example-input",
            "inputPath": "example-input-path",
            "inputTransformer": {
              "inputPathsMap": {
                "example": "example-input-paths-map"
              },
              "inputTemplate": "example-input-template"
            },
            "kinesisParameters": {
              "partitionKeyPath": "example-partition-key-path"
            },
            "redshiftDataParameters": {
              "database": "example-database",
              "dbUser": "example-db-user",
              "secretManagerArn": "example-secret-manager-arn",
              "sql": "example-sql",
              "sqls": [
                "example-sql"
              ],
              "statementName": "example-statement-name",
              "withEvent": false
            },
            "retryPolicy": {
              "maximumEventAgeInSeconds": 1,
              "maximumRetryAttempts": 1
            },
            "roleArn": "example-role-arn",
            "runCommandParameters": {
              "runCommandTargets": [
                {
                  "key": "example-key",
                  "values": [
                    "example-value"
                  ]
                }
              ]
            },
            "sageMakerPipelineParameters": {
              "pipelineParameterList": [
                {
                  "name": "example-name",
                  "value": "example-value"
                }
              ]
            },
            "sqsParameters": {
              "messageGroupId": "example-message-group-id"
            }
          }
        ]
      }
    }
  }
}
```
