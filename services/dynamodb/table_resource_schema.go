package dynamodb

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func dynamodbTableResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "DynamoDBTableDefinition",
		Description: "The definition of an AWS DynamoDB table.",
		Required:    []string{"tableName", "attributeDefinitions", "keySchema"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"tableName": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The name of the DynamoDB table.",
				FormattedDescription: "The name of the DynamoDB table. " +
					"Table names must be unique within a region and can contain alphanumeric characters, " +
					"underscores, hyphens, and periods.",
				Pattern:      `^[a-zA-Z0-9_.-]+$`,
				MinLength:    3,
				MaxLength:    255,
				MustRecreate: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("my-table"),
					core.MappingNodeFromString("users"),
					core.MappingNodeFromString("orders-2024"),
				},
			},
			"attributeDefinitions": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of attributes that describe the key schema for the table and indexes.",
				FormattedDescription: "A list of attributes that describe the key schema for the table and indexes. " +
					"Each attribute definition specifies the attribute name and data type.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:     provider.ResourceDefinitionsSchemaTypeObject,
					Label:    "AttributeDefinition",
					Required: []string{"attributeName", "attributeType"},
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"attributeName": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The name of the attribute.",
							MinLength:   1,
							MaxLength:   255,
							Examples: []*core.MappingNode{
								core.MappingNodeFromString("id"),
								core.MappingNodeFromString("userId"),
								core.MappingNodeFromString("createdAt"),
							},
						},
						"attributeType": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The data type for the attribute. S = String, N = Number, B = Binary.",
							AllowedValues: []*core.MappingNode{
								core.MappingNodeFromString("S"),
								core.MappingNodeFromString("N"),
								core.MappingNodeFromString("B"),
							},
							Examples: []*core.MappingNode{
								core.MappingNodeFromString("S"),
								core.MappingNodeFromString("N"),
							},
						},
					},
				},
				Examples: []*core.MappingNode{
					{
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"attributeName": core.MappingNodeFromString("id"),
									"attributeType": core.MappingNodeFromString("S"),
								},
							},
						},
					},
				},
			},
			"keySchema": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "Specifies the attributes that make up the primary key for the table.",
				FormattedDescription: "Specifies the attributes that make up the primary key for the table. " +
					"The first element is the partition key (HASH), and the optional second element is the sort key (RANGE).",
				Items: &provider.ResourceDefinitionsSchema{
					Type:     provider.ResourceDefinitionsSchemaTypeObject,
					Label:    "KeySchemaElement",
					Required: []string{"attributeName", "keyType"},
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"attributeName": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The name of the key attribute.",
							MinLength:   1,
							MaxLength:   255,
							Examples: []*core.MappingNode{
								core.MappingNodeFromString("id"),
								core.MappingNodeFromString("sortKey"),
							},
						},
						"keyType": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The role of the key attribute. HASH = partition key, RANGE = sort key.",
							AllowedValues: []*core.MappingNode{
								core.MappingNodeFromString("HASH"),
								core.MappingNodeFromString("RANGE"),
							},
							Examples: []*core.MappingNode{
								core.MappingNodeFromString("HASH"),
								core.MappingNodeFromString("RANGE"),
							},
						},
					},
				},
			},
			"billingMode": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "Controls how you are charged for read and write throughput.",
				FormattedDescription: "Controls how you are charged for read and write throughput. " +
					"PROVISIONED = you specify the read and write capacity. " +
					"PAY_PER_REQUEST = on-demand capacity mode.",
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("PROVISIONED"),
					core.MappingNodeFromString("PAY_PER_REQUEST"),
				},
				Default:  core.MappingNodeFromString("PAY_PER_REQUEST"),
				Nullable: true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("PAY_PER_REQUEST"),
					core.MappingNodeFromString("PROVISIONED"),
				},
			},
			"provisionedThroughput": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "Provisioned throughput settings for the table.",
				FormattedDescription: "Provisioned throughput settings for the table. " +
					"Required when billingMode is PROVISIONED.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"readCapacityUnits": {
						Type:        provider.ResourceDefinitionsSchemaTypeInteger,
						Description: "The maximum number of strongly consistent reads per second.",
						Minimum:     core.ScalarFromInt(1),
						Examples: []*core.MappingNode{
							core.MappingNodeFromInt(5),
							core.MappingNodeFromInt(100),
						},
					},
					"writeCapacityUnits": {
						Type:        provider.ResourceDefinitionsSchemaTypeInteger,
						Description: "The maximum number of writes per second.",
						Minimum:     core.ScalarFromInt(1),
						Examples: []*core.MappingNode{
							core.MappingNodeFromInt(5),
							core.MappingNodeFromInt(100),
						},
					},
				},
				Nullable: true,
			},
			"globalSecondaryIndexes": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "Global secondary indexes to be created on the table.",
				FormattedDescription: "Global secondary indexes to be created on the table. " +
					"Each GSI has its own key schema and can have different provisioned throughput.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:     provider.ResourceDefinitionsSchemaTypeObject,
					Label:    "GlobalSecondaryIndex",
					Required: []string{"indexName", "keySchema", "projection"},
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"indexName": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The name of the global secondary index.",
							MinLength:   3,
							MaxLength:   255,
							Examples: []*core.MappingNode{
								core.MappingNodeFromString("userId-index"),
								core.MappingNodeFromString("email-index"),
							},
						},
						"keySchema": {
							Type:        provider.ResourceDefinitionsSchemaTypeArray,
							Description: "The key schema for the global secondary index.",
							Items: &provider.ResourceDefinitionsSchema{
								Type:     provider.ResourceDefinitionsSchemaTypeObject,
								Label:    "KeySchemaElement",
								Required: []string{"attributeName", "keyType"},
								Attributes: map[string]*provider.ResourceDefinitionsSchema{
									"attributeName": {
										Type:        provider.ResourceDefinitionsSchemaTypeString,
										Description: "The name of the key attribute.",
									},
									"keyType": {
										Type:        provider.ResourceDefinitionsSchemaTypeString,
										Description: "The role of the key attribute.",
										AllowedValues: []*core.MappingNode{
											core.MappingNodeFromString("HASH"),
											core.MappingNodeFromString("RANGE"),
										},
									},
								},
							},
						},
						"projection": {
							Type:        provider.ResourceDefinitionsSchemaTypeObject,
							Description: "Represents attributes that are copied from the table into the index.",
							Required:    []string{"projectionType"},
							Attributes: map[string]*provider.ResourceDefinitionsSchema{
								"projectionType": {
									Type:        provider.ResourceDefinitionsSchemaTypeString,
									Description: "The set of attributes that are projected into the index.",
									AllowedValues: []*core.MappingNode{
										core.MappingNodeFromString("ALL"),
										core.MappingNodeFromString("KEYS_ONLY"),
										core.MappingNodeFromString("INCLUDE"),
									},
								},
								"nonKeyAttributes": {
									Type:        provider.ResourceDefinitionsSchemaTypeArray,
									Description: "A list of non-key attribute names to project into the index.",
									Items: &provider.ResourceDefinitionsSchema{
										Type: provider.ResourceDefinitionsSchemaTypeString,
									},
									Nullable: true,
								},
							},
						},
						"provisionedThroughput": {
							Type:        provider.ResourceDefinitionsSchemaTypeObject,
							Description: "Provisioned throughput for the index.",
							Attributes: map[string]*provider.ResourceDefinitionsSchema{
								"readCapacityUnits": {
									Type:    provider.ResourceDefinitionsSchemaTypeInteger,
									Minimum: core.ScalarFromInt(1),
								},
								"writeCapacityUnits": {
									Type:    provider.ResourceDefinitionsSchemaTypeInteger,
									Minimum: core.ScalarFromInt(1),
								},
							},
							Nullable: true,
						},
					},
				},
				Nullable: true,
			},
			"localSecondaryIndexes": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "Local secondary indexes to be created on the table.",
				FormattedDescription: "Local secondary indexes to be created on the table. " +
					"LSIs share the same partition key as the table but have a different sort key.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:     provider.ResourceDefinitionsSchemaTypeObject,
					Label:    "LocalSecondaryIndex",
					Required: []string{"indexName", "keySchema", "projection"},
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"indexName": {
							Type:      provider.ResourceDefinitionsSchemaTypeString,
							MinLength: 3,
							MaxLength: 255,
						},
						"keySchema": {
							Type: provider.ResourceDefinitionsSchemaTypeArray,
							Items: &provider.ResourceDefinitionsSchema{
								Type:     provider.ResourceDefinitionsSchemaTypeObject,
								Required: []string{"attributeName", "keyType"},
								Attributes: map[string]*provider.ResourceDefinitionsSchema{
									"attributeName": {
										Type: provider.ResourceDefinitionsSchemaTypeString,
									},
									"keyType": {
										Type: provider.ResourceDefinitionsSchemaTypeString,
										AllowedValues: []*core.MappingNode{
											core.MappingNodeFromString("HASH"),
											core.MappingNodeFromString("RANGE"),
										},
									},
								},
							},
						},
						"projection": {
							Type:     provider.ResourceDefinitionsSchemaTypeObject,
							Required: []string{"projectionType"},
							Attributes: map[string]*provider.ResourceDefinitionsSchema{
								"projectionType": {
									Type: provider.ResourceDefinitionsSchemaTypeString,
									AllowedValues: []*core.MappingNode{
										core.MappingNodeFromString("ALL"),
										core.MappingNodeFromString("KEYS_ONLY"),
										core.MappingNodeFromString("INCLUDE"),
									},
								},
								"nonKeyAttributes": {
									Type: provider.ResourceDefinitionsSchemaTypeArray,
									Items: &provider.ResourceDefinitionsSchema{
										Type: provider.ResourceDefinitionsSchemaTypeString,
									},
									Nullable: true,
								},
							},
						},
					},
				},
				MustRecreate: true,
				Nullable:     true,
			},
			"streamSpecification": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "The settings for DynamoDB Streams on the table.",
				FormattedDescription: "The settings for DynamoDB Streams on the table. " +
					"When enabled, you can use the stream to trigger Lambda functions.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"streamEnabled": {
						Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
						Description: "Indicates whether DynamoDB Streams is enabled on the table.",
						Examples: []*core.MappingNode{
							core.MappingNodeFromBool(true),
							core.MappingNodeFromBool(false),
						},
					},
					"streamViewType": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "When an item in the table is modified, StreamViewType determines what information is written to the stream.",
						AllowedValues: []*core.MappingNode{
							core.MappingNodeFromString("KEYS_ONLY"),
							core.MappingNodeFromString("NEW_IMAGE"),
							core.MappingNodeFromString("OLD_IMAGE"),
							core.MappingNodeFromString("NEW_AND_OLD_IMAGES"),
						},
						Nullable: true,
						Examples: []*core.MappingNode{
							core.MappingNodeFromString("NEW_AND_OLD_IMAGES"),
							core.MappingNodeFromString("NEW_IMAGE"),
						},
					},
				},
				Nullable: true,
			},
			"sseSpecification": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "Server-side encryption settings for the table.",
				FormattedDescription: "Server-side encryption settings for the table. " +
					"You can use an AWS owned key or a customer managed key.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"enabled": {
						Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
						Description: "Indicates whether server-side encryption is enabled.",
						Default:     core.MappingNodeFromBool(true),
					},
					"sseType": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "Server-side encryption type. KMS = customer managed key, AES256 = AWS owned key.",
						AllowedValues: []*core.MappingNode{
							core.MappingNodeFromString("KMS"),
							core.MappingNodeFromString("AES256"),
						},
						Nullable: true,
					},
					"kmsMasterKeyId": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The KMS key ARN or ID to use for encryption.",
						Nullable:    true,
						Examples: []*core.MappingNode{
							core.MappingNodeFromString("arn:aws:kms:us-west-2:123456789012:key/12345678-1234-1234-1234-123456789012"),
							core.MappingNodeFromString("alias/my-key"),
						},
					},
				},
				Nullable: true,
			},
			"timeToLiveSpecification": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "Time to Live (TTL) settings for the table.",
				FormattedDescription: "Time to Live (TTL) settings for the table. " +
					"TTL automatically deletes items after their expiration time.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"enabled": {
						Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
						Description: "Indicates whether TTL is enabled on the table.",
					},
					"attributeName": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The name of the TTL attribute used to store the expiration time.",
						Examples: []*core.MappingNode{
							core.MappingNodeFromString("ttl"),
							core.MappingNodeFromString("expiresAt"),
						},
					},
				},
				Nullable: true,
			},
			"pointInTimeRecoverySpecification": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Description: "Point in time recovery settings for the table.",
				FormattedDescription: "Point in time recovery settings for the table. " +
					"When enabled, you can restore the table to any point in time within the last 35 days.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"pointInTimeRecoveryEnabled": {
						Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
						Description: "Indicates whether point in time recovery is enabled.",
						Default:     core.MappingNodeFromBool(false),
					},
				},
				Nullable: true,
			},
			"deletionProtectionEnabled": {
				Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
				Description: "Indicates whether deletion protection is enabled on the table.",
				FormattedDescription: "Indicates whether deletion protection is enabled on the table. " +
					"When enabled, the table cannot be deleted.",
				Default:  core.MappingNodeFromBool(false),
				Nullable: true,
			},
			"tableClass": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The table class of the DynamoDB table.",
				FormattedDescription: "The table class of the DynamoDB table. " +
					"STANDARD = default, optimized for frequently accessed data. " +
					"STANDARD_INFREQUENT_ACCESS = optimized for infrequently accessed data.",
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("STANDARD"),
					core.MappingNodeFromString("STANDARD_INFREQUENT_ACCESS"),
				},
				Default:  core.MappingNodeFromString("STANDARD"),
				Nullable: true,
			},
			"tags": {
				Type:             provider.ResourceDefinitionsSchemaTypeArray,
				SortArrayByField: "key",
				Description:      "A list of tags to associate with the DynamoDB table.",
				FormattedDescription: "A list of tags to associate with the DynamoDB table. " +
					"Tags are key-value pairs that help you organize and identify your AWS resources.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:     provider.ResourceDefinitionsSchemaTypeObject,
					Label:    "Tag",
					Required: []string{"key", "value"},
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"key": {
							Type:      provider.ResourceDefinitionsSchemaTypeString,
							MinLength: 1,
							MaxLength: 128,
						},
						"value": {
							Type:      provider.ResourceDefinitionsSchemaTypeString,
							MinLength: 0,
							MaxLength: 256,
						},
					},
				},
				Nullable: true,
			},

			// Computed fields
			"arn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the DynamoDB table.",
				Computed:    true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:dynamodb:us-west-2:123456789012:table/my-table"),
				},
			},
			"tableId": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The unique identifier for the table.",
				Computed:    true,
			},
			"latestStreamArn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The ARN of the latest stream for the table.",
				FormattedDescription: "The ARN of the latest stream for the table. " +
					"Only present if streams are enabled.",
				Computed: true,
				Nullable: true,
			},
			"latestStreamLabel": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "A timestamp label for the latest stream.",
				Computed:    true,
				Nullable:    true,
			},
		},
	}
}
