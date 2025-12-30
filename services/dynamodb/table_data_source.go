package dynamodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// TableDataSource returns a data source implementation for an AWS DynamoDB Table.
func TableDataSource(
	dynamodbServiceFactory pluginutils.ServiceFactory[*aws.Config, dynamodbservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.DataSource {
	yamlExample, _ := examples.ReadFile("examples/datasources/dynamodb_table_yaml.md")
	jsoncExample, _ := examples.ReadFile("examples/datasources/dynamodb_table_jsonc.md")

	fetcher := &dynamodbTableDataSourceFetcher{
		dynamodbServiceFactory: dynamodbServiceFactory,
		awsConfigStore:         awsConfigStore,
	}

	return &providerv1.DataSourceDefinition{
		Type:             "aws/dynamodb/table",
		Label:            "Amazon DynamoDB Table",
		PlainTextSummary: "A data source for retrieving an Amazon DynamoDB table.",
		FormattedDescription: "The data source type used to define a [DynamoDB table](https://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_DescribeTable.html) " +
			"managed externally in AWS.",
		MarkdownExamples: []string{
			string(yamlExample),
			string(jsoncExample),
		},
		Fields: dynamodbTableDataSourceSchema(),
		FilterFields: map[string]*provider.DataSourceFilterSchema{
			"arn": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The ARN of the DynamoDB table to retrieve.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
				ConflictsWith: []string{"tableName"},
			},
			"tableName": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The name of the DynamoDB table to retrieve.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
				ConflictsWith: []string{"arn"},
			},
			"region": {
				Type:        provider.DataSourceFilterSearchValueTypeString,
				Description: "The region of the DynamoDB table to retrieve. Defaults to the region of the provider.",
				FormattedDescription: "The [region](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints) " +
					"of the DynamoDB table to retrieve. Defaults to the region of the provider.",
				SupportedOperators: []schema.DataSourceFilterOperator{
					schema.DataSourceFilterOperatorEquals,
				},
			},
		},
		FetchFunc: fetcher.Fetch,
	}
}

type dynamodbTableDataSourceFetcher struct {
	dynamodbServiceFactory pluginutils.ServiceFactory[*aws.Config, dynamodbservice.Service]
	awsConfigStore         pluginutils.ServiceConfigStore[*aws.Config]
}

func (f *dynamodbTableDataSourceFetcher) getDynamoDBService(
	ctx context.Context,
	input *provider.DataSourceFetchInput,
) (dynamodbservice.Service, error) {
	meta := map[string]*core.MappingNode{
		"region": extractRegionFromFilters(input.DataSourceWithResolvedSubs.Filter),
	}
	awsConfig, err := f.awsConfigStore.FromProviderContext(
		ctx,
		input.ProviderContext,
		meta,
	)
	if err != nil {
		return nil, err
	}

	return f.dynamodbServiceFactory(awsConfig, input.ProviderContext), nil
}

func (f *dynamodbTableDataSourceFetcher) Fetch(
	ctx context.Context,
	input *provider.DataSourceFetchInput,
) (*provider.DataSourceFetchOutput, error) {
	dynamodbService, err := f.getDynamoDBService(ctx, input)
	if err != nil {
		return nil, err
	}

	tableName := extractTableNameFromFilters(input.DataSourceWithResolvedSubs.Filter)
	if tableName == nil {
		return nil, fmt.Errorf(
			"table name or ARN is required for the DynamoDB table data source",
		)
	}

	// Describe the table
	describeOutput, err := dynamodbService.DescribeTable(
		ctx,
		&dynamodb.DescribeTableInput{
			TableName: aws.String(core.StringValue(tableName)),
		},
	)
	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if errors.As(err, &notFoundErr) {
			return nil, fmt.Errorf("table %s not found", core.StringValue(tableName))
		}
		return nil, err
	}

	data := f.buildTableData(describeOutput.Table)

	// Add TTL info
	err = f.addTTLInfo(ctx, dynamodbService, tableName, data)
	if err != nil {
		return nil, err
	}

	// Add PITR info
	err = f.addPITRInfo(ctx, dynamodbService, tableName, data)
	if err != nil {
		return nil, err
	}

	// Add tags
	if describeOutput.Table.TableArn != nil {
		err = f.addTags(ctx, dynamodbService, describeOutput.Table.TableArn, data)
		if err != nil {
			return nil, err
		}
	}

	return &provider.DataSourceFetchOutput{
		Data: data,
	}, nil
}

func (f *dynamodbTableDataSourceFetcher) buildTableData(
	table *types.TableDescription,
) map[string]*core.MappingNode {
	data := map[string]*core.MappingNode{}

	// Required fields
	if table.TableArn != nil {
		data["arn"] = core.MappingNodeFromString(aws.ToString(table.TableArn))
	}
	if table.TableId != nil {
		data["tableId"] = core.MappingNodeFromString(aws.ToString(table.TableId))
	}
	if table.TableName != nil {
		data["tableName"] = core.MappingNodeFromString(aws.ToString(table.TableName))
	}

	// Table status
	data["tableStatus"] = core.MappingNodeFromString(string(table.TableStatus))

	// Billing mode
	if table.BillingModeSummary != nil {
		data["billingMode"] = core.MappingNodeFromString(string(table.BillingModeSummary.BillingMode))
	}

	// Table class
	if table.TableClassSummary != nil {
		data["tableClass"] = core.MappingNodeFromString(string(table.TableClassSummary.TableClass))
	}

	// Deletion protection
	if table.DeletionProtectionEnabled != nil {
		data["deletionProtectionEnabled"] = core.MappingNodeFromBool(
			aws.ToBool(table.DeletionProtectionEnabled),
		)
	}

	// Attribute definitions
	if len(table.AttributeDefinitions) > 0 {
		attrDefs := make([]*core.MappingNode, len(table.AttributeDefinitions))
		for i, attr := range table.AttributeDefinitions {
			attrDefs[i] = core.MappingNodeFields(
				"attributeName", core.MappingNodeFromString(aws.ToString(attr.AttributeName)),
				"attributeType", core.MappingNodeFromString(string(attr.AttributeType)),
			)
		}
		data["attributeDefinitions"] = core.MappingNodeItems(attrDefs...)
	}

	// Key schema
	if len(table.KeySchema) > 0 {
		keySchemaItems := make([]*core.MappingNode, len(table.KeySchema))
		for i, key := range table.KeySchema {
			keySchemaItems[i] = core.MappingNodeFields(
				"attributeName", core.MappingNodeFromString(aws.ToString(key.AttributeName)),
				"keyType", core.MappingNodeFromString(string(key.KeyType)),
			)
		}
		data["keySchema"] = core.MappingNodeItems(keySchemaItems...)
	}

	// Provisioned throughput
	if table.ProvisionedThroughput != nil {
		data["provisionedThroughput.readCapacityUnits"] = core.MappingNodeFromInt(
			int(aws.ToInt64(table.ProvisionedThroughput.ReadCapacityUnits)),
		)
		data["provisionedThroughput.writeCapacityUnits"] = core.MappingNodeFromInt(
			int(aws.ToInt64(table.ProvisionedThroughput.WriteCapacityUnits)),
		)
	}

	// Stream specification
	if table.StreamSpecification != nil {
		data["streamSpecification.streamEnabled"] = core.MappingNodeFromBool(
			aws.ToBool(table.StreamSpecification.StreamEnabled),
		)
		data["streamSpecification.streamViewType"] = core.MappingNodeFromString(
			string(table.StreamSpecification.StreamViewType),
		)
	}

	// Latest stream info
	if table.LatestStreamArn != nil {
		data["latestStreamArn"] = core.MappingNodeFromString(aws.ToString(table.LatestStreamArn))
	}
	if table.LatestStreamLabel != nil {
		data["latestStreamLabel"] = core.MappingNodeFromString(aws.ToString(table.LatestStreamLabel))
	}

	// SSE specification
	if table.SSEDescription != nil {
		data["sseSpecification.sseType"] = core.MappingNodeFromString(
			string(table.SSEDescription.SSEType),
		)
		if table.SSEDescription.KMSMasterKeyArn != nil {
			data["sseSpecification.kmsMasterKeyArn"] = core.MappingNodeFromString(
				aws.ToString(table.SSEDescription.KMSMasterKeyArn),
			)
		}
	}

	// Item count and size
	if table.ItemCount != nil {
		data["itemCount"] = core.MappingNodeFromInt(int(aws.ToInt64(table.ItemCount)))
	}
	if table.TableSizeBytes != nil {
		data["tableSizeBytes"] = core.MappingNodeFromInt(int(aws.ToInt64(table.TableSizeBytes)))
	}

	// Global secondary indexes
	if len(table.GlobalSecondaryIndexes) > 0 {
		gsiItems := make([]*core.MappingNode, len(table.GlobalSecondaryIndexes))
		for i, gsi := range table.GlobalSecondaryIndexes {
			gsiItems[i] = f.buildGSIData(&gsi)
		}
		data["globalSecondaryIndexes"] = core.MappingNodeItems(gsiItems...)
	}

	// Local secondary indexes
	if len(table.LocalSecondaryIndexes) > 0 {
		lsiItems := make([]*core.MappingNode, len(table.LocalSecondaryIndexes))
		for i, lsi := range table.LocalSecondaryIndexes {
			lsiItems[i] = f.buildLSIData(&lsi)
		}
		data["localSecondaryIndexes"] = core.MappingNodeItems(lsiItems...)
	}

	return data
}

func (f *dynamodbTableDataSourceFetcher) buildGSIData(
	gsi *types.GlobalSecondaryIndexDescription,
) *core.MappingNode {
	fields := map[string]*core.MappingNode{}

	if gsi.IndexName != nil {
		fields["indexName"] = core.MappingNodeFromString(aws.ToString(gsi.IndexName))
	}

	// Key schema
	if len(gsi.KeySchema) > 0 {
		keySchemaItems := make([]*core.MappingNode, len(gsi.KeySchema))
		for i, key := range gsi.KeySchema {
			keySchemaItems[i] = core.MappingNodeFields(
				"attributeName", core.MappingNodeFromString(aws.ToString(key.AttributeName)),
				"keyType", core.MappingNodeFromString(string(key.KeyType)),
			)
		}
		fields["keySchema"] = core.MappingNodeItems(keySchemaItems...)
	}

	// Projection
	if gsi.Projection != nil {
		projFields := map[string]*core.MappingNode{
			"projectionType": core.MappingNodeFromString(string(gsi.Projection.ProjectionType)),
		}
		if len(gsi.Projection.NonKeyAttributes) > 0 {
			projFields["nonKeyAttributes"] = core.MappingNodeFromStringSlice(gsi.Projection.NonKeyAttributes)
		}
		fields["projection"] = &core.MappingNode{Fields: projFields}
	}

	// Provisioned throughput
	if gsi.ProvisionedThroughput != nil {
		fields["provisionedThroughput"] = core.MappingNodeFields(
			"readCapacityUnits", core.MappingNodeFromInt(int(aws.ToInt64(gsi.ProvisionedThroughput.ReadCapacityUnits))),
			"writeCapacityUnits", core.MappingNodeFromInt(int(aws.ToInt64(gsi.ProvisionedThroughput.WriteCapacityUnits))),
		)
	}

	// Index status
	fields["indexStatus"] = core.MappingNodeFromString(string(gsi.IndexStatus))

	return &core.MappingNode{Fields: fields}
}

func (f *dynamodbTableDataSourceFetcher) buildLSIData(
	lsi *types.LocalSecondaryIndexDescription,
) *core.MappingNode {
	fields := map[string]*core.MappingNode{}

	if lsi.IndexName != nil {
		fields["indexName"] = core.MappingNodeFromString(aws.ToString(lsi.IndexName))
	}

	// Key schema
	if len(lsi.KeySchema) > 0 {
		keySchemaItems := make([]*core.MappingNode, len(lsi.KeySchema))
		for i, key := range lsi.KeySchema {
			keySchemaItems[i] = core.MappingNodeFields(
				"attributeName", core.MappingNodeFromString(aws.ToString(key.AttributeName)),
				"keyType", core.MappingNodeFromString(string(key.KeyType)),
			)
		}
		fields["keySchema"] = core.MappingNodeItems(keySchemaItems...)
	}

	// Projection
	if lsi.Projection != nil {
		projFields := map[string]*core.MappingNode{
			"projectionType": core.MappingNodeFromString(string(lsi.Projection.ProjectionType)),
		}
		if len(lsi.Projection.NonKeyAttributes) > 0 {
			projFields["nonKeyAttributes"] = core.MappingNodeFromStringSlice(lsi.Projection.NonKeyAttributes)
		}
		fields["projection"] = &core.MappingNode{Fields: projFields}
	}

	return &core.MappingNode{Fields: fields}
}

func (f *dynamodbTableDataSourceFetcher) addTTLInfo(
	ctx context.Context,
	service dynamodbservice.Service,
	tableName *core.MappingNode,
	data map[string]*core.MappingNode,
) error {
	ttlOutput, err := service.DescribeTimeToLive(ctx, &dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(core.StringValue(tableName)),
	})
	if err != nil {
		return fmt.Errorf("failed to describe TTL: %w", err)
	}

	if ttlOutput.TimeToLiveDescription == nil {
		return nil
	}

	ttl := ttlOutput.TimeToLiveDescription
	enabled := ttl.TimeToLiveStatus == types.TimeToLiveStatusEnabled ||
		ttl.TimeToLiveStatus == types.TimeToLiveStatusEnabling

	data["timeToLiveSpecification.enabled"] = core.MappingNodeFromBool(enabled)
	if ttl.AttributeName != nil {
		data["timeToLiveSpecification.attributeName"] = core.MappingNodeFromString(
			aws.ToString(ttl.AttributeName),
		)
	}

	return nil
}

func (f *dynamodbTableDataSourceFetcher) addPITRInfo(
	ctx context.Context,
	service dynamodbservice.Service,
	tableName *core.MappingNode,
	data map[string]*core.MappingNode,
) error {
	backupsOutput, err := service.DescribeContinuousBackups(ctx, &dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(core.StringValue(tableName)),
	})
	if err != nil {
		return fmt.Errorf("failed to describe continuous backups: %w", err)
	}

	if backupsOutput.ContinuousBackupsDescription == nil {
		return nil
	}

	pitr := backupsOutput.ContinuousBackupsDescription.PointInTimeRecoveryDescription
	if pitr == nil {
		return nil
	}

	enabled := pitr.PointInTimeRecoveryStatus == types.PointInTimeRecoveryStatusEnabled
	data["pointInTimeRecoverySpecification.pointInTimeRecoveryEnabled"] = core.MappingNodeFromBool(enabled)

	return nil
}

func (f *dynamodbTableDataSourceFetcher) addTags(
	ctx context.Context,
	service dynamodbservice.Service,
	tableArn *string,
	data map[string]*core.MappingNode,
) error {
	tagsOutput, err := service.ListTagsOfResource(ctx, &dynamodb.ListTagsOfResourceInput{
		ResourceArn: tableArn,
	})
	if err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	if len(tagsOutput.Tags) == 0 {
		return nil
	}

	tagItems := make([]*core.MappingNode, len(tagsOutput.Tags))
	for i, tag := range tagsOutput.Tags {
		tagItems[i] = core.MappingNodeFields(
			"key", core.MappingNodeFromString(aws.ToString(tag.Key)),
			"value", core.MappingNodeFromString(aws.ToString(tag.Value)),
		)
	}
	data["tags"] = core.MappingNodeItems(tagItems...)

	return nil
}

// extractTableNameFromFilters extracts the table name from the data source filters.
// It checks for either "tableName" or "arn" filter and returns the table name.
func extractTableNameFromFilters(
	filters *provider.ResolvedDataSourceFilters,
) *core.MappingNode {
	// Try tableName first
	tableName := pluginutils.ExtractMatchFromFilters(filters, "tableName")
	if tableName != nil {
		return tableName
	}

	// Try arn and extract table name from it
	arn := pluginutils.ExtractMatchFromFilters(filters, "arn")
	if arn != nil {
		arnValue := core.StringValue(arn)
		tableNameStr := extractTableNameFromARN(arnValue)
		return core.MappingNodeFromString(tableNameStr)
	}

	return nil
}

// extractRegionFromFilters extracts the region from data source filters.
func extractRegionFromFilters(
	filters *provider.ResolvedDataSourceFilters,
) *core.MappingNode {
	return pluginutils.ExtractMatchFromFilters(filters, "region")
}
