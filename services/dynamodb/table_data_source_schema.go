package dynamodb

import "github.com/newstack-cloud/bluelink/libs/blueprint/provider"

func dynamodbTableDataSourceSchema() map[string]*provider.DataSourceSpecSchema {
	return map[string]*provider.DataSourceSpecSchema{
		"arn": {
			Label:       "Table ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The ARN of the DynamoDB table.",
			Nullable:    false,
		},
		"tableId": {
			Label:       "Table ID",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The unique identifier for the DynamoDB table.",
			Nullable:    false,
		},
		"tableName": {
			Label:       "Table Name",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The name of the DynamoDB table.",
			Nullable:    false,
		},
		"billingMode": {
			Label:       "Billing Mode",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The billing mode of the table (PROVISIONED or PAY_PER_REQUEST).",
			Nullable:    true,
		},
		"tableClass": {
			Label:       "Table Class",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The table class of the DynamoDB table (STANDARD or STANDARD_INFREQUENT_ACCESS).",
			Nullable:    true,
		},
		"tableStatus": {
			Label:       "Table Status",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The current status of the table (CREATING, UPDATING, DELETING, ACTIVE, etc.).",
			Nullable:    true,
		},
		"deletionProtectionEnabled": {
			Label:       "Deletion Protection Enabled",
			Type:        provider.DataSourceSpecTypeBoolean,
			Description: "Whether deletion protection is enabled for the table.",
			Nullable:    true,
		},
		"attributeDefinitions": {
			Label:       "Attribute Definitions",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "The attribute definitions for the table.",
			Nullable:    true,
		},
		"keySchema": {
			Label:       "Key Schema",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "The key schema for the table.",
			Nullable:    true,
		},
		"globalSecondaryIndexes": {
			Label:       "Global Secondary Indexes",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "The global secondary indexes for the table.",
			Nullable:    true,
		},
		"localSecondaryIndexes": {
			Label:       "Local Secondary Indexes",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "The local secondary indexes for the table.",
			Nullable:    true,
		},
		"provisionedThroughput.readCapacityUnits": {
			Label:       "Read Capacity Units",
			Type:        provider.DataSourceSpecTypeInteger,
			Description: "The provisioned read capacity units for the table.",
			Nullable:    true,
		},
		"provisionedThroughput.writeCapacityUnits": {
			Label:       "Write Capacity Units",
			Type:        provider.DataSourceSpecTypeInteger,
			Description: "The provisioned write capacity units for the table.",
			Nullable:    true,
		},
		"streamSpecification.streamEnabled": {
			Label:       "Stream Enabled",
			Type:        provider.DataSourceSpecTypeBoolean,
			Description: "Whether DynamoDB Streams is enabled on the table.",
			Nullable:    true,
		},
		"streamSpecification.streamViewType": {
			Label:       "Stream View Type",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The type of data to write to the stream (KEYS_ONLY, NEW_IMAGE, OLD_IMAGE, NEW_AND_OLD_IMAGES).",
			Nullable:    true,
		},
		"latestStreamArn": {
			Label:       "Latest Stream ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The ARN of the latest DynamoDB stream for the table.",
			Nullable:    true,
		},
		"latestStreamLabel": {
			Label:       "Latest Stream Label",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The label of the latest DynamoDB stream for the table.",
			Nullable:    true,
		},
		"timeToLiveSpecification.enabled": {
			Label:       "TTL Enabled",
			Type:        provider.DataSourceSpecTypeBoolean,
			Description: "Whether Time to Live (TTL) is enabled on the table.",
			Nullable:    true,
		},
		"timeToLiveSpecification.attributeName": {
			Label:       "TTL Attribute Name",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The name of the attribute used for TTL.",
			Nullable:    true,
		},
		"pointInTimeRecoverySpecification.pointInTimeRecoveryEnabled": {
			Label:       "Point-in-Time Recovery Enabled",
			Type:        provider.DataSourceSpecTypeBoolean,
			Description: "Whether Point-in-Time Recovery (PITR) is enabled on the table.",
			Nullable:    true,
		},
		"sseSpecification.sseType": {
			Label:       "SSE Type",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The type of server-side encryption (AES256 or KMS).",
			Nullable:    true,
		},
		"sseSpecification.kmsMasterKeyArn": {
			Label:       "KMS Master Key ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The ARN of the KMS key used for server-side encryption.",
			Nullable:    true,
		},
		"itemCount": {
			Label:       "Item Count",
			Type:        provider.DataSourceSpecTypeInteger,
			Description: "The approximate number of items in the table.",
			Nullable:    true,
		},
		"tableSizeBytes": {
			Label:       "Table Size Bytes",
			Type:        provider.DataSourceSpecTypeInteger,
			Description: "The approximate size of the table in bytes.",
			Nullable:    true,
		},
		"tags": {
			Label:       "Tags",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "The tags associated with the table.",
			Nullable:    true,
		},
	}
}
