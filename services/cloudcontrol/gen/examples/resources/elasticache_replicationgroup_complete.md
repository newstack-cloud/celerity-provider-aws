A AWS ElastiCache ReplicationGroup configured with the full set of available properties.

```blueprintlang
version "2025-11-02"

resource replicationGroup: aws/elasticache/replicationGroup {
    metadata {
        displayName = "AWS ElastiCache ReplicationGroup complete"
    }
    spec {
        atRestEncryptionEnabled = false
        authToken = "example-auth-token"
        autoMinorVersionUpgrade = false
        automaticFailoverEnabled = false
        cacheNodeType = "example-cache-node-type"
        cacheParameterGroupName = "example-cache-parameter-group-name"
        cacheSecurityGroupNames = [
            "example-cache-security-group-name"
        ]
        cacheSubnetGroupName = "example-cache-subnet-group-name"
        clusterMode = "example-cluster-mode"
        dataTieringEnabled = false
        durability = "default"
        engine = "example-engine"
        engineVersion = "example-engine-version"
        globalReplicationGroupId = "example-global-replication-group-id"
        ipDiscovery = "example-ip-discovery"
        kmsKeyId = "example-kms-key-id"
        logDeliveryConfigurations = [
            {
                destinationDetails = {
                    cloudWatchLogsDetails = {
                        logGroup = "example-log-group"
                    },
                    kinesisFirehoseDetails = {
                        deliveryStream = "example-delivery-stream"
                    }
                },
                destinationType = "example-destination-type",
                logFormat = "example-log-format",
                logType = "example-log-type"
            }
        ]
        multiAZEnabled = false
        networkType = "example-network-type"
        nodeGroupConfiguration = [
            {
                nodeGroupId = "example-node-group-id",
                primaryAvailabilityZone = "example-primary-availability-zone",
                replicaAvailabilityZones = [
                    "example-replica-availability-zone"
                ],
                replicaCount = 1,
                slots = "example-slots"
            }
        ]
        notificationTopicArn = "example-notification-topic-arn"
        numCacheClusters = 1
        numNodeGroups = 1
        port = 1
        preferredCacheClusterAZs = [
            "example-preferred-cache-cluster-a-z"
        ]
        preferredMaintenanceWindow = "example-preferred-maintenance-window"
        primaryClusterId = "example-primary-cluster-id"
        replicasPerNodeGroup = 1
        replicationGroupDescription = "example-replication-group-description"
        replicationGroupId = "example-replication-group-id"
        securityGroupIds = [
            "example-security-group-id"
        ]
        snapshotArns = [
            "example-snapshot-arn"
        ]
        snapshotName = "example-snapshot-name"
        snapshotRetentionLimit = 1
        snapshotWindow = "example-snapshot-window"
        snapshottingClusterId = "example-snapshotting-cluster-id"
        tags = [
            {
                key = "example-key",
                value = "example-value"
            }
        ]
        transitEncryptionEnabled = false
        transitEncryptionMode = "example-transit-encryption-mode"
        userGroupIds = [
            "example-user-group-id"
        ]
    }
}
```

```yaml
version: "2025-11-02"
resources:
    replicationGroup:
        type: aws/elasticache/replicationGroup
        metadata:
            displayName: AWS ElastiCache ReplicationGroup complete
        spec:
            atRestEncryptionEnabled: false
            authToken: example-auth-token
            autoMinorVersionUpgrade: false
            automaticFailoverEnabled: false
            cacheNodeType: example-cache-node-type
            cacheParameterGroupName: example-cache-parameter-group-name
            cacheSecurityGroupNames:
                - example-cache-security-group-name
            cacheSubnetGroupName: example-cache-subnet-group-name
            clusterMode: example-cluster-mode
            dataTieringEnabled: false
            durability: default
            engine: example-engine
            engineVersion: example-engine-version
            globalReplicationGroupId: example-global-replication-group-id
            ipDiscovery: example-ip-discovery
            kmsKeyId: example-kms-key-id
            logDeliveryConfigurations:
                - destinationDetails:
                    cloudWatchLogsDetails:
                        logGroup: example-log-group
                    kinesisFirehoseDetails:
                        deliveryStream: example-delivery-stream
                  destinationType: example-destination-type
                  logFormat: example-log-format
                  logType: example-log-type
            multiAZEnabled: false
            networkType: example-network-type
            nodeGroupConfiguration:
                - nodeGroupId: example-node-group-id
                  primaryAvailabilityZone: example-primary-availability-zone
                  replicaAvailabilityZones:
                    - example-replica-availability-zone
                  replicaCount: 1
                  slots: example-slots
            notificationTopicArn: example-notification-topic-arn
            numCacheClusters: 1
            numNodeGroups: 1
            port: 1
            preferredCacheClusterAZs:
                - example-preferred-cache-cluster-a-z
            preferredMaintenanceWindow: example-preferred-maintenance-window
            primaryClusterId: example-primary-cluster-id
            replicasPerNodeGroup: 1
            replicationGroupDescription: example-replication-group-description
            replicationGroupId: example-replication-group-id
            securityGroupIds:
                - example-security-group-id
            snapshotArns:
                - example-snapshot-arn
            snapshotName: example-snapshot-name
            snapshotRetentionLimit: 1
            snapshotWindow: example-snapshot-window
            snapshottingClusterId: example-snapshotting-cluster-id
            tags:
                - key: example-key
                  value: example-value
            transitEncryptionEnabled: false
            transitEncryptionMode: example-transit-encryption-mode
            userGroupIds:
                - example-user-group-id
```

```javascript
{
  "version": "2025-11-02",
  "resources": {
    "replicationGroup": {
      "type": "aws/elasticache/replicationGroup",
      "metadata": {
        "displayName": "AWS ElastiCache ReplicationGroup complete"
      },
      "spec": {
        "atRestEncryptionEnabled": false,
        "authToken": "example-auth-token",
        "autoMinorVersionUpgrade": false,
        "automaticFailoverEnabled": false,
        "cacheNodeType": "example-cache-node-type",
        "cacheParameterGroupName": "example-cache-parameter-group-name",
        "cacheSecurityGroupNames": [
          "example-cache-security-group-name"
        ],
        "cacheSubnetGroupName": "example-cache-subnet-group-name",
        "clusterMode": "example-cluster-mode",
        "dataTieringEnabled": false,
        "durability": "default",
        "engine": "example-engine",
        "engineVersion": "example-engine-version",
        "globalReplicationGroupId": "example-global-replication-group-id",
        "ipDiscovery": "example-ip-discovery",
        "kmsKeyId": "example-kms-key-id",
        "logDeliveryConfigurations": [
          {
            "destinationDetails": {
              "cloudWatchLogsDetails": {
                "logGroup": "example-log-group"
              },
              "kinesisFirehoseDetails": {
                "deliveryStream": "example-delivery-stream"
              }
            },
            "destinationType": "example-destination-type",
            "logFormat": "example-log-format",
            "logType": "example-log-type"
          }
        ],
        "multiAZEnabled": false,
        "networkType": "example-network-type",
        "nodeGroupConfiguration": [
          {
            "nodeGroupId": "example-node-group-id",
            "primaryAvailabilityZone": "example-primary-availability-zone",
            "replicaAvailabilityZones": [
              "example-replica-availability-zone"
            ],
            "replicaCount": 1,
            "slots": "example-slots"
          }
        ],
        "notificationTopicArn": "example-notification-topic-arn",
        "numCacheClusters": 1,
        "numNodeGroups": 1,
        "port": 1,
        "preferredCacheClusterAZs": [
          "example-preferred-cache-cluster-a-z"
        ],
        "preferredMaintenanceWindow": "example-preferred-maintenance-window",
        "primaryClusterId": "example-primary-cluster-id",
        "replicasPerNodeGroup": 1,
        "replicationGroupDescription": "example-replication-group-description",
        "replicationGroupId": "example-replication-group-id",
        "securityGroupIds": [
          "example-security-group-id"
        ],
        "snapshotArns": [
          "example-snapshot-arn"
        ],
        "snapshotName": "example-snapshot-name",
        "snapshotRetentionLimit": 1,
        "snapshotWindow": "example-snapshot-window",
        "snapshottingClusterId": "example-snapshotting-cluster-id",
        "tags": [
          {
            "key": "example-key",
            "value": "example-value"
          }
        ],
        "transitEncryptionEnabled": false,
        "transitEncryptionMode": "example-transit-encryption-mode",
        "userGroupIds": [
          "example-user-group-id"
        ]
      }
    }
  }
}
```
