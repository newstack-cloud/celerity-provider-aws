A AWS RDS DBCluster configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource dBCluster: aws/rds/dbCluster {
    metadata {
        displayName = "AWS RDS DBCluster complete"
    }
    spec {
        allocatedStorage = 1
        associatedRoles = [
            {
                featureName = "example-feature-name",
                roleArn = "example-role-arn"
            }
        ]
        autoMinorVersionUpgrade = false
        availabilityZones = [
            "example-availability-zone"
        ]
        backtrackWindow = 0
        backupRetentionPeriod = 1
        clusterScalabilityType = "example-cluster-scalability-type"
        copyTagsToSnapshot = false
        databaseInsightsMode = "example-database-insights-mode"
        databaseName = "example-database-name"
        dbClusterIdentifier = "example-db-cluster-identifier"
        dbClusterInstanceClass = "example-db-cluster-instance-class"
        dbClusterParameterGroupName = "example-db-cluster-parameter-group-name"
        dbInstanceParameterGroupName = "example-db-instance-parameter-group-name"
        dbSubnetGroupName = "example-db-subnet-group-name"
        dbSystemId = "example-db-system-id"
        deleteAutomatedBackups = false
        deletionProtection = false
        domain = "example-domain"
        domainIAMRoleName = "example-domain-i-a-m-role-name"
        enableCloudwatchLogsExports = [
            "example-enable-cloudwatch-logs-export"
        ]
        enableGlobalWriteForwarding = false
        enableHttpEndpoint = false
        enableIAMDatabaseAuthentication = false
        enableLocalWriteForwarding = false
        engine = "example-engine"
        engineLifecycleSupport = "example-engine-lifecycle-support"
        engineMode = "example-engine-mode"
        engineVersion = "example-engine-version"
        globalClusterIdentifier = "example-global-cluster-identifier"
        iops = 1
        kmsKeyId = "example-kms-key-id"
        manageMasterUserPassword = false
        masterUserAuthenticationType = "example-master-user-authentication-type"
        masterUserPassword = "example-master-user-password"
        masterUserSecret = {
            kmsKeyId = "example-kms-key-id"
        }
        masterUsername = "example-master-username"
        monitoringInterval = 1
        monitoringRoleArn = "example-monitoring-role-arn"
        networkType = "example-network-type"
        performanceInsightsEnabled = false
        performanceInsightsKmsKeyId = "example-performance-insights-kms-key-id"
        performanceInsightsRetentionPeriod = 1
        port = 1
        preferredBackupWindow = "example-preferred-backup-window"
        preferredMaintenanceWindow = "example-preferred-maintenance-window"
        publiclyAccessible = false
        replicationSourceIdentifier = "example-replication-source-identifier"
        restoreToTime = "example-restore-to-time"
        restoreType = "example-restore-type"
        scalingConfiguration = {
            autoPause = false,
            maxCapacity = 1,
            minCapacity = 1,
            secondsBeforeTimeout = 1,
            secondsUntilAutoPause = 1,
            timeoutAction = "example-timeout-action"
        }
        serverlessV2ScalingConfiguration = {
            maxCapacity = 1,
            minCapacity = 1,
            secondsUntilAutoPause = 1
        }
        snapshotIdentifier = "example-snapshot-identifier"
        sourceDBClusterIdentifier = "example-source-d-b-cluster-identifier"
        sourceDbClusterResourceId = "example-source-db-cluster-resource-id"
        sourceRegion = "example-source-region"
        storageEncrypted = false
        storageType = "example-storage-type"
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        useLatestRestorableTime = false
        vpcSecurityGroupIds = [
            "example-vpc-security-group-id"
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    dBCluster:
        type: aws/rds/dbCluster
        metadata:
            displayName: AWS RDS DBCluster complete
        spec:
            allocatedStorage: 1
            associatedRoles:
                - featureName: example-feature-name
                  roleArn: example-role-arn
            autoMinorVersionUpgrade: false
            availabilityZones:
                - example-availability-zone
            backtrackWindow: 0
            backupRetentionPeriod: 1
            clusterScalabilityType: example-cluster-scalability-type
            copyTagsToSnapshot: false
            databaseInsightsMode: example-database-insights-mode
            databaseName: example-database-name
            dbClusterIdentifier: example-db-cluster-identifier
            dbClusterInstanceClass: example-db-cluster-instance-class
            dbClusterParameterGroupName: example-db-cluster-parameter-group-name
            dbInstanceParameterGroupName: example-db-instance-parameter-group-name
            dbSubnetGroupName: example-db-subnet-group-name
            dbSystemId: example-db-system-id
            deleteAutomatedBackups: false
            deletionProtection: false
            domain: example-domain
            domainIAMRoleName: example-domain-i-a-m-role-name
            enableCloudwatchLogsExports:
                - example-enable-cloudwatch-logs-export
            enableGlobalWriteForwarding: false
            enableHttpEndpoint: false
            enableIAMDatabaseAuthentication: false
            enableLocalWriteForwarding: false
            engine: example-engine
            engineLifecycleSupport: example-engine-lifecycle-support
            engineMode: example-engine-mode
            engineVersion: example-engine-version
            globalClusterIdentifier: example-global-cluster-identifier
            iops: 1
            kmsKeyId: example-kms-key-id
            manageMasterUserPassword: false
            masterUserAuthenticationType: example-master-user-authentication-type
            masterUserPassword: example-master-user-password
            masterUserSecret:
                kmsKeyId: example-kms-key-id
            masterUsername: example-master-username
            monitoringInterval: 1
            monitoringRoleArn: example-monitoring-role-arn
            networkType: example-network-type
            performanceInsightsEnabled: false
            performanceInsightsKmsKeyId: example-performance-insights-kms-key-id
            performanceInsightsRetentionPeriod: 1
            port: 1
            preferredBackupWindow: example-preferred-backup-window
            preferredMaintenanceWindow: example-preferred-maintenance-window
            publiclyAccessible: false
            replicationSourceIdentifier: example-replication-source-identifier
            restoreToTime: example-restore-to-time
            restoreType: example-restore-type
            scalingConfiguration:
                autoPause: false
                maxCapacity: 1
                minCapacity: 1
                secondsBeforeTimeout: 1
                secondsUntilAutoPause: 1
                timeoutAction: example-timeout-action
            serverlessV2ScalingConfiguration:
                maxCapacity: 1
                minCapacity: 1
                secondsUntilAutoPause: 1
            snapshotIdentifier: example-snapshot-identifier
            sourceDBClusterIdentifier: example-source-d-b-cluster-identifier
            sourceDbClusterResourceId: example-source-db-cluster-resource-id
            sourceRegion: example-source-region
            storageEncrypted: false
            storageType: example-storage-type
            tags:
                - key: example-key
                  value: example-value
            useLatestRestorableTime: false
            vpcSecurityGroupIds:
                - example-vpc-security-group-id
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "dBCluster": {
      "type": "aws/rds/dbCluster",
      "metadata": {
        "displayName": "AWS RDS DBCluster complete"
      },
      "spec": {
        "allocatedStorage": 1,
        "associatedRoles": [
          {
            "featureName": "example-feature-name",
            "roleArn": "example-role-arn"
          }
        ],
        "autoMinorVersionUpgrade": false,
        "availabilityZones": [
          "example-availability-zone"
        ],
        "backtrackWindow": 0,
        "backupRetentionPeriod": 1,
        "clusterScalabilityType": "example-cluster-scalability-type",
        "copyTagsToSnapshot": false,
        "databaseInsightsMode": "example-database-insights-mode",
        "databaseName": "example-database-name",
        "dbClusterIdentifier": "example-db-cluster-identifier",
        "dbClusterInstanceClass": "example-db-cluster-instance-class",
        "dbClusterParameterGroupName": "example-db-cluster-parameter-group-name",
        "dbInstanceParameterGroupName": "example-db-instance-parameter-group-name",
        "dbSubnetGroupName": "example-db-subnet-group-name",
        "dbSystemId": "example-db-system-id",
        "deleteAutomatedBackups": false,
        "deletionProtection": false,
        "domain": "example-domain",
        "domainIAMRoleName": "example-domain-i-a-m-role-name",
        "enableCloudwatchLogsExports": [
          "example-enable-cloudwatch-logs-export"
        ],
        "enableGlobalWriteForwarding": false,
        "enableHttpEndpoint": false,
        "enableIAMDatabaseAuthentication": false,
        "enableLocalWriteForwarding": false,
        "engine": "example-engine",
        "engineLifecycleSupport": "example-engine-lifecycle-support",
        "engineMode": "example-engine-mode",
        "engineVersion": "example-engine-version",
        "globalClusterIdentifier": "example-global-cluster-identifier",
        "iops": 1,
        "kmsKeyId": "example-kms-key-id",
        "manageMasterUserPassword": false,
        "masterUserAuthenticationType": "example-master-user-authentication-type",
        "masterUserPassword": "example-master-user-password",
        "masterUserSecret": {
          "kmsKeyId": "example-kms-key-id"
        },
        "masterUsername": "example-master-username",
        "monitoringInterval": 1,
        "monitoringRoleArn": "example-monitoring-role-arn",
        "networkType": "example-network-type",
        "performanceInsightsEnabled": false,
        "performanceInsightsKmsKeyId": "example-performance-insights-kms-key-id",
        "performanceInsightsRetentionPeriod": 1,
        "port": 1,
        "preferredBackupWindow": "example-preferred-backup-window",
        "preferredMaintenanceWindow": "example-preferred-maintenance-window",
        "publiclyAccessible": false,
        "replicationSourceIdentifier": "example-replication-source-identifier",
        "restoreToTime": "example-restore-to-time",
        "restoreType": "example-restore-type",
        "scalingConfiguration": {
          "autoPause": false,
          "maxCapacity": 1,
          "minCapacity": 1,
          "secondsBeforeTimeout": 1,
          "secondsUntilAutoPause": 1,
          "timeoutAction": "example-timeout-action"
        },
        "serverlessV2ScalingConfiguration": {
          "maxCapacity": 1,
          "minCapacity": 1,
          "secondsUntilAutoPause": 1
        },
        "snapshotIdentifier": "example-snapshot-identifier",
        "sourceDBClusterIdentifier": "example-source-d-b-cluster-identifier",
        "sourceDbClusterResourceId": "example-source-db-cluster-resource-id",
        "sourceRegion": "example-source-region",
        "storageEncrypted": false,
        "storageType": "example-storage-type",
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "useLatestRestorableTime": false,
        "vpcSecurityGroupIds": [
          "example-vpc-security-group-id"
        ]
      }
    }
  }
}
```
