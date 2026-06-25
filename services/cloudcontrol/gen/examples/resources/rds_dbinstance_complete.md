A AWS RDS DBInstance configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource dBInstance: aws/rds/dbInstance {
    metadata {
        displayName = "AWS RDS DBInstance complete"
    }
    spec {
        additionalStorageVolumes = [
            {
                allocatedStorage = "example-allocated-storage",
                iops = 1,
                maxAllocatedStorage = 1,
                storageThroughput = 1,
                storageType = "example-storage-type",
                volumeName = "example-volume-name"
            }
        ]
        allocatedStorage = "example-allocated-storage"
        allowMajorVersionUpgrade = false
        applyImmediately = false
        associatedRoles = [
            {
                featureName = "example-feature-name",
                roleArn = "example-role-arn"
            }
        ]
        autoMinorVersionUpgrade = false
        automaticBackupReplicationKmsKeyId = "example-automatic-backup-replication-kms-key-id"
        automaticBackupReplicationRegion = "example-automatic-backup-replication-region"
        automaticBackupReplicationRetentionPeriod = 1
        availabilityZone = "example-availability-zone"
        backupRetentionPeriod = 0
        backupTarget = "example-backup-target"
        caCertificateIdentifier = "example-ca-certificate-identifier"
        certificateRotationRestart = false
        characterSetName = "example-character-set-name"
        copyTagsToSnapshot = false
        customIAMInstanceProfile = "example-custom-i-a-m-instance-profile"
        databaseInsightsMode = "example-database-insights-mode"
        dbClusterIdentifier = "example-db-cluster-identifier"
        dbClusterSnapshotIdentifier = "example-db-cluster-snapshot-identifier"
        dbInstanceClass = "example-db-instance-class"
        dbInstanceIdentifier = "example-db-instance-identifier"
        dbName = "example-db-name"
        dbParameterGroupName = "example-db-parameter-group-name"
        dbSecurityGroups = [
            "example-db-security-group"
        ]
        dbSnapshotIdentifier = "example-db-snapshot-identifier"
        dbSubnetGroupName = "example-db-subnet-group-name"
        dbSystemId = "example-db-system-id"
        dedicatedLogVolume = false
        deleteAutomatedBackups = false
        deletionProtection = false
        domain = "example-domain"
        domainAuthSecretArn = "example-domain-auth-secret-arn"
        domainDnsIps = [
            "example-domain-dns-ip"
        ]
        domainFqdn = "example-domain-fqdn"
        domainIAMRoleName = "example-domain-i-a-m-role-name"
        domainOu = "example-domain-ou"
        enableCloudwatchLogsExports = [
            "example-enable-cloudwatch-logs-export"
        ]
        enableIAMDatabaseAuthentication = false
        enablePerformanceInsights = false
        engine = "example-engine"
        engineLifecycleSupport = "example-engine-lifecycle-support"
        engineVersion = "example-engine-version"
        iops = 1
        kmsKeyId = "example-kms-key-id"
        licenseModel = "example-license-model"
        manageMasterUserPassword = false
        masterUserAuthenticationType = "example-master-user-authentication-type"
        masterUserPassword = "example-master-user-password"
        masterUserSecret = {
            kmsKeyId = "example-kms-key-id"
        }
        masterUsername = "example-master-username"
        maxAllocatedStorage = 1
        monitoringInterval = 1
        monitoringRoleArn = "example-monitoring-role-arn"
        multiAZ = false
        ncharCharacterSetName = "example-nchar-character-set-name"
        networkType = "example-network-type"
        optionGroupName = "example-option-group-name"
        performanceInsightsKMSKeyId = "example-performance-insights-k-m-s-key-id"
        performanceInsightsRetentionPeriod = 1
        port = "example-port"
        preferredBackupWindow = "example-preferred-backup-window"
        preferredMaintenanceWindow = "example-preferred-maintenance-window"
        processorFeatures = [
            {
                name = "coreCount",
                value = "example-value"
            }
        ]
        promotionTier = 0
        publiclyAccessible = false
        replicaMode = "example-replica-mode"
        restoreTime = "example-restore-time"
        sourceDBClusterIdentifier = "example-source-d-b-cluster-identifier"
        sourceDBInstanceAutomatedBackupsArn = "example-source-d-b-instance-automated-backups-arn"
        sourceDBInstanceIdentifier = "example-source-d-b-instance-identifier"
        sourceDbiResourceId = "example-source-dbi-resource-id"
        sourceRegion = "example-source-region"
        storageEncrypted = false
        storageThroughput = 1
        storageType = "example-storage-type"
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        tdeCredentialArn = "example-tde-credential-arn"
        tdeCredentialPassword = "example-tde-credential-password"
        timezone = "example-timezone"
        useDefaultProcessorFeatures = false
        useLatestRestorableTime = false
        vpcSecurityGroups = [
            "example-vpc-security-group"
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    dBInstance:
        type: aws/rds/dbInstance
        metadata:
            displayName: AWS RDS DBInstance complete
        spec:
            additionalStorageVolumes:
                - allocatedStorage: example-allocated-storage
                  iops: 1
                  maxAllocatedStorage: 1
                  storageThroughput: 1
                  storageType: example-storage-type
                  volumeName: example-volume-name
            allocatedStorage: example-allocated-storage
            allowMajorVersionUpgrade: false
            applyImmediately: false
            associatedRoles:
                - featureName: example-feature-name
                  roleArn: example-role-arn
            autoMinorVersionUpgrade: false
            automaticBackupReplicationKmsKeyId: example-automatic-backup-replication-kms-key-id
            automaticBackupReplicationRegion: example-automatic-backup-replication-region
            automaticBackupReplicationRetentionPeriod: 1
            availabilityZone: example-availability-zone
            backupRetentionPeriod: 0
            backupTarget: example-backup-target
            caCertificateIdentifier: example-ca-certificate-identifier
            certificateRotationRestart: false
            characterSetName: example-character-set-name
            copyTagsToSnapshot: false
            customIAMInstanceProfile: example-custom-i-a-m-instance-profile
            databaseInsightsMode: example-database-insights-mode
            dbClusterIdentifier: example-db-cluster-identifier
            dbClusterSnapshotIdentifier: example-db-cluster-snapshot-identifier
            dbInstanceClass: example-db-instance-class
            dbInstanceIdentifier: example-db-instance-identifier
            dbName: example-db-name
            dbParameterGroupName: example-db-parameter-group-name
            dbSecurityGroups:
                - example-db-security-group
            dbSnapshotIdentifier: example-db-snapshot-identifier
            dbSubnetGroupName: example-db-subnet-group-name
            dbSystemId: example-db-system-id
            dedicatedLogVolume: false
            deleteAutomatedBackups: false
            deletionProtection: false
            domain: example-domain
            domainAuthSecretArn: example-domain-auth-secret-arn
            domainDnsIps:
                - example-domain-dns-ip
            domainFqdn: example-domain-fqdn
            domainIAMRoleName: example-domain-i-a-m-role-name
            domainOu: example-domain-ou
            enableCloudwatchLogsExports:
                - example-enable-cloudwatch-logs-export
            enableIAMDatabaseAuthentication: false
            enablePerformanceInsights: false
            engine: example-engine
            engineLifecycleSupport: example-engine-lifecycle-support
            engineVersion: example-engine-version
            iops: 1
            kmsKeyId: example-kms-key-id
            licenseModel: example-license-model
            manageMasterUserPassword: false
            masterUserAuthenticationType: example-master-user-authentication-type
            masterUserPassword: example-master-user-password
            masterUserSecret:
                kmsKeyId: example-kms-key-id
            masterUsername: example-master-username
            maxAllocatedStorage: 1
            monitoringInterval: 1
            monitoringRoleArn: example-monitoring-role-arn
            multiAZ: false
            ncharCharacterSetName: example-nchar-character-set-name
            networkType: example-network-type
            optionGroupName: example-option-group-name
            performanceInsightsKMSKeyId: example-performance-insights-k-m-s-key-id
            performanceInsightsRetentionPeriod: 1
            port: example-port
            preferredBackupWindow: example-preferred-backup-window
            preferredMaintenanceWindow: example-preferred-maintenance-window
            processorFeatures:
                - name: coreCount
                  value: example-value
            promotionTier: 0
            publiclyAccessible: false
            replicaMode: example-replica-mode
            restoreTime: example-restore-time
            sourceDBClusterIdentifier: example-source-d-b-cluster-identifier
            sourceDBInstanceAutomatedBackupsArn: example-source-d-b-instance-automated-backups-arn
            sourceDBInstanceIdentifier: example-source-d-b-instance-identifier
            sourceDbiResourceId: example-source-dbi-resource-id
            sourceRegion: example-source-region
            storageEncrypted: false
            storageThroughput: 1
            storageType: example-storage-type
            tags:
                - key: example-key
                  value: example-value
            tdeCredentialArn: example-tde-credential-arn
            tdeCredentialPassword: example-tde-credential-password
            timezone: example-timezone
            useDefaultProcessorFeatures: false
            useLatestRestorableTime: false
            vpcSecurityGroups:
                - example-vpc-security-group
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "dBInstance": {
      "type": "aws/rds/dbInstance",
      "metadata": {
        "displayName": "AWS RDS DBInstance complete"
      },
      "spec": {
        "additionalStorageVolumes": [
          {
            "allocatedStorage": "example-allocated-storage",
            "iops": 1,
            "maxAllocatedStorage": 1,
            "storageThroughput": 1,
            "storageType": "example-storage-type",
            "volumeName": "example-volume-name"
          }
        ],
        "allocatedStorage": "example-allocated-storage",
        "allowMajorVersionUpgrade": false,
        "applyImmediately": false,
        "associatedRoles": [
          {
            "featureName": "example-feature-name",
            "roleArn": "example-role-arn"
          }
        ],
        "autoMinorVersionUpgrade": false,
        "automaticBackupReplicationKmsKeyId": "example-automatic-backup-replication-kms-key-id",
        "automaticBackupReplicationRegion": "example-automatic-backup-replication-region",
        "automaticBackupReplicationRetentionPeriod": 1,
        "availabilityZone": "example-availability-zone",
        "backupRetentionPeriod": 0,
        "backupTarget": "example-backup-target",
        "caCertificateIdentifier": "example-ca-certificate-identifier",
        "certificateRotationRestart": false,
        "characterSetName": "example-character-set-name",
        "copyTagsToSnapshot": false,
        "customIAMInstanceProfile": "example-custom-i-a-m-instance-profile",
        "databaseInsightsMode": "example-database-insights-mode",
        "dbClusterIdentifier": "example-db-cluster-identifier",
        "dbClusterSnapshotIdentifier": "example-db-cluster-snapshot-identifier",
        "dbInstanceClass": "example-db-instance-class",
        "dbInstanceIdentifier": "example-db-instance-identifier",
        "dbName": "example-db-name",
        "dbParameterGroupName": "example-db-parameter-group-name",
        "dbSecurityGroups": [
          "example-db-security-group"
        ],
        "dbSnapshotIdentifier": "example-db-snapshot-identifier",
        "dbSubnetGroupName": "example-db-subnet-group-name",
        "dbSystemId": "example-db-system-id",
        "dedicatedLogVolume": false,
        "deleteAutomatedBackups": false,
        "deletionProtection": false,
        "domain": "example-domain",
        "domainAuthSecretArn": "example-domain-auth-secret-arn",
        "domainDnsIps": [
          "example-domain-dns-ip"
        ],
        "domainFqdn": "example-domain-fqdn",
        "domainIAMRoleName": "example-domain-i-a-m-role-name",
        "domainOu": "example-domain-ou",
        "enableCloudwatchLogsExports": [
          "example-enable-cloudwatch-logs-export"
        ],
        "enableIAMDatabaseAuthentication": false,
        "enablePerformanceInsights": false,
        "engine": "example-engine",
        "engineLifecycleSupport": "example-engine-lifecycle-support",
        "engineVersion": "example-engine-version",
        "iops": 1,
        "kmsKeyId": "example-kms-key-id",
        "licenseModel": "example-license-model",
        "manageMasterUserPassword": false,
        "masterUserAuthenticationType": "example-master-user-authentication-type",
        "masterUserPassword": "example-master-user-password",
        "masterUserSecret": {
          "kmsKeyId": "example-kms-key-id"
        },
        "masterUsername": "example-master-username",
        "maxAllocatedStorage": 1,
        "monitoringInterval": 1,
        "monitoringRoleArn": "example-monitoring-role-arn",
        "multiAZ": false,
        "ncharCharacterSetName": "example-nchar-character-set-name",
        "networkType": "example-network-type",
        "optionGroupName": "example-option-group-name",
        "performanceInsightsKMSKeyId": "example-performance-insights-k-m-s-key-id",
        "performanceInsightsRetentionPeriod": 1,
        "port": "example-port",
        "preferredBackupWindow": "example-preferred-backup-window",
        "preferredMaintenanceWindow": "example-preferred-maintenance-window",
        "processorFeatures": [
          {
            "name": "coreCount",
            "value": "example-value"
          }
        ],
        "promotionTier": 0,
        "publiclyAccessible": false,
        "replicaMode": "example-replica-mode",
        "restoreTime": "example-restore-time",
        "sourceDBClusterIdentifier": "example-source-d-b-cluster-identifier",
        "sourceDBInstanceAutomatedBackupsArn": "example-source-d-b-instance-automated-backups-arn",
        "sourceDBInstanceIdentifier": "example-source-d-b-instance-identifier",
        "sourceDbiResourceId": "example-source-dbi-resource-id",
        "sourceRegion": "example-source-region",
        "storageEncrypted": false,
        "storageThroughput": 1,
        "storageType": "example-storage-type",
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "tdeCredentialArn": "example-tde-credential-arn",
        "tdeCredentialPassword": "example-tde-credential-password",
        "timezone": "example-timezone",
        "useDefaultProcessorFeatures": false,
        "useLatestRestorableTime": false,
        "vpcSecurityGroups": [
          "example-vpc-security-group"
        ]
      }
    }
  }
}
```
