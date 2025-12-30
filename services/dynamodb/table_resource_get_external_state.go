package dynamodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (a *dynamodbTableResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	dynamodbService, err := a.getDynamoDBService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Try to get ARN from current spec first
	tableARN := ""
	tableName := ""
	if input.CurrentResourceSpec != nil && input.CurrentResourceSpec.Fields != nil {
		tableARN = core.StringValue(input.CurrentResourceSpec.Fields["arn"])
		tableName = core.StringValue(input.CurrentResourceSpec.Fields["tableName"])
	}

	// If no ARN, attempt fallback lookup by Bluelink tags
	if tableARN == "" {
		fallbackArn, err := a.lookupTableARNByTags(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to lookup table by tags: %w", err)
		}
		if fallbackArn == "" {
			// Resource doesn't exist yet
			return &provider.ResourceGetExternalStateOutput{
				ResourceSpecState: &core.MappingNode{
					Fields: map[string]*core.MappingNode{},
				},
			}, nil
		}
		tableARN = fallbackArn
	}

	// Use table name if available, otherwise extract from ARN
	if tableName == "" {
		// Extract table name from ARN: arn:aws:dynamodb:region:account:table/tablename
		tableName = extractTableNameFromARN(tableARN)
	}

	// Get table description
	describeOutput, err := dynamodbService.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		// Check if the table doesn't exist
		var notFoundErr *types.ResourceNotFoundException
		if ok := errors.As(err, &notFoundErr); ok {
			return &provider.ResourceGetExternalStateOutput{
				ResourceSpecState: nil,
			}, nil
		}
		return nil, fmt.Errorf("failed to describe table: %w", err)
	}

	// Build the base resource spec state
	resourceSpecState := buildTableResourceSpecState(describeOutput.Table)

	// Get TTL settings
	ttlOutput, err := dynamodbService.DescribeTimeToLive(ctx, &dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe TTL: %w", err)
	}
	addTTLToSpec(ttlOutput, resourceSpecState.Fields)

	// Get PITR settings
	backupsOutput, err := dynamodbService.DescribeContinuousBackups(ctx, &dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe continuous backups: %w", err)
	}
	addPITRToSpec(backupsOutput, resourceSpecState.Fields)

	// Get tags and filter out Bluelink provenance tags
	tagsOutput, err := dynamodbService.ListTagsOfResource(ctx, &dynamodb.ListTagsOfResourceInput{
		ResourceArn: aws.String(tableARN),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}
	addTagsToSpec(tagsOutput, resourceSpecState.Fields, input.ProviderContext)

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}

func (a *dynamodbTableResourceActions) lookupTableARNByTags(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (string, error) {
	// Build the Bluelink tag filters
	tagFilters := utils.BuildBluelinkTagFiltersForLookup(input)
	if tagFilters == nil {
		// Tagging is not enabled, cannot perform fallback lookup
		return "", nil
	}

	resGroupTagService, err := a.getResourceGroupTaggingService(ctx, input.ProviderContext)
	if err != nil {
		return "", err
	}

	// Search for DynamoDB tables with matching tags
	getResourcesOutput, err := resGroupTagService.GetResources(
		ctx,
		&resourcegroupstaggingapi.GetResourcesInput{
			TagFilters: tagFilters,
			ResourceTypeFilters: []string{
				"dynamodb:table",
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to get resources by tags: %w", err)
	}

	if len(getResourcesOutput.ResourceTagMappingList) == 0 {
		return "", nil
	}

	return aws.ToString(getResourcesOutput.ResourceTagMappingList[0].ResourceARN), nil
}

func extractTableNameFromARN(arn string) string {
	// ARN format: arn:aws:dynamodb:region:account:table/tablename
	// We want to extract "tablename"
	for i := len(arn) - 1; i >= 0; i-- {
		if arn[i] == '/' {
			return arn[i+1:]
		}
	}
	return arn
}

func buildTableResourceSpecState(table *types.TableDescription) *core.MappingNode {
	fields := map[string]*core.MappingNode{}

	if table.TableName != nil {
		fields["tableName"] = core.MappingNodeFromString(aws.ToString(table.TableName))
	}

	if table.TableArn != nil {
		fields["arn"] = core.MappingNodeFromString(aws.ToString(table.TableArn))
	}

	if table.TableId != nil {
		fields["tableId"] = core.MappingNodeFromString(aws.ToString(table.TableId))
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
		fields["attributeDefinitions"] = core.MappingNodeItems(attrDefs...)
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
		fields["keySchema"] = core.MappingNodeItems(keySchemaItems...)
	}

	// Billing mode
	if table.BillingModeSummary != nil {
		fields["billingMode"] = core.MappingNodeFromString(string(table.BillingModeSummary.BillingMode))
	}

	// Provisioned throughput
	if table.ProvisionedThroughput != nil {
		fields["provisionedThroughput"] = core.MappingNodeFields(
			"readCapacityUnits", core.MappingNodeFromInt(int(aws.ToInt64(table.ProvisionedThroughput.ReadCapacityUnits))),
			"writeCapacityUnits", core.MappingNodeFromInt(int(aws.ToInt64(table.ProvisionedThroughput.WriteCapacityUnits))),
		)
	}

	// Stream specification
	if table.StreamSpecification != nil {
		fields["streamSpecification"] = core.MappingNodeFields(
			"streamEnabled", core.MappingNodeFromBool(aws.ToBool(table.StreamSpecification.StreamEnabled)),
			"streamViewType", core.MappingNodeFromString(string(table.StreamSpecification.StreamViewType)),
		)
	}

	// Latest stream info (computed fields)
	if table.LatestStreamArn != nil {
		fields["latestStreamArn"] = core.MappingNodeFromString(aws.ToString(table.LatestStreamArn))
	}
	if table.LatestStreamLabel != nil {
		fields["latestStreamLabel"] = core.MappingNodeFromString(aws.ToString(table.LatestStreamLabel))
	}

	// Deletion protection
	if table.DeletionProtectionEnabled != nil {
		fields["deletionProtectionEnabled"] = core.MappingNodeFromBool(aws.ToBool(table.DeletionProtectionEnabled))
	}

	// Table class
	if table.TableClassSummary != nil {
		fields["tableClass"] = core.MappingNodeFromString(string(table.TableClassSummary.TableClass))
	}

	// Global secondary indexes
	if len(table.GlobalSecondaryIndexes) > 0 {
		gsiItems := make([]*core.MappingNode, len(table.GlobalSecondaryIndexes))
		for i, gsi := range table.GlobalSecondaryIndexes {
			gsiItems[i] = buildGSISpec(&gsi)
		}
		fields["globalSecondaryIndexes"] = core.MappingNodeItems(gsiItems...)
	}

	// Local secondary indexes
	if len(table.LocalSecondaryIndexes) > 0 {
		lsiItems := make([]*core.MappingNode, len(table.LocalSecondaryIndexes))
		for i, lsi := range table.LocalSecondaryIndexes {
			lsiItems[i] = buildLSISpec(&lsi)
		}
		fields["localSecondaryIndexes"] = core.MappingNodeItems(lsiItems...)
	}

	return &core.MappingNode{Fields: fields}
}

func buildGSISpec(gsi *types.GlobalSecondaryIndexDescription) *core.MappingNode {
	gsiFields := map[string]*core.MappingNode{}

	if gsi.IndexName != nil {
		gsiFields["indexName"] = core.MappingNodeFromString(aws.ToString(gsi.IndexName))
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
		gsiFields["keySchema"] = core.MappingNodeItems(keySchemaItems...)
	}

	// Projection
	if gsi.Projection != nil {
		projFields := map[string]*core.MappingNode{
			"projectionType": core.MappingNodeFromString(string(gsi.Projection.ProjectionType)),
		}
		if len(gsi.Projection.NonKeyAttributes) > 0 {
			projFields["nonKeyAttributes"] = core.MappingNodeFromStringSlice(gsi.Projection.NonKeyAttributes)
		}
		gsiFields["projection"] = &core.MappingNode{Fields: projFields}
	}

	// Provisioned throughput
	if gsi.ProvisionedThroughput != nil {
		gsiFields["provisionedThroughput"] = core.MappingNodeFields(
			"readCapacityUnits", core.MappingNodeFromInt(int(aws.ToInt64(gsi.ProvisionedThroughput.ReadCapacityUnits))),
			"writeCapacityUnits", core.MappingNodeFromInt(int(aws.ToInt64(gsi.ProvisionedThroughput.WriteCapacityUnits))),
		)
	}

	return &core.MappingNode{Fields: gsiFields}
}

func buildLSISpec(lsi *types.LocalSecondaryIndexDescription) *core.MappingNode {
	lsiFields := map[string]*core.MappingNode{}

	if lsi.IndexName != nil {
		lsiFields["indexName"] = core.MappingNodeFromString(aws.ToString(lsi.IndexName))
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
		lsiFields["keySchema"] = core.MappingNodeItems(keySchemaItems...)
	}

	// Projection
	if lsi.Projection != nil {
		projFields := map[string]*core.MappingNode{
			"projectionType": core.MappingNodeFromString(string(lsi.Projection.ProjectionType)),
		}
		if len(lsi.Projection.NonKeyAttributes) > 0 {
			projFields["nonKeyAttributes"] = core.MappingNodeFromStringSlice(lsi.Projection.NonKeyAttributes)
		}
		lsiFields["projection"] = &core.MappingNode{Fields: projFields}
	}

	return &core.MappingNode{Fields: lsiFields}
}

func addTTLToSpec(ttlOutput *dynamodb.DescribeTimeToLiveOutput, fields map[string]*core.MappingNode) {
	if ttlOutput.TimeToLiveDescription == nil {
		return
	}

	ttl := ttlOutput.TimeToLiveDescription
	// Only add TTL if it's enabled
	if ttl.TimeToLiveStatus == types.TimeToLiveStatusEnabled ||
		ttl.TimeToLiveStatus == types.TimeToLiveStatusEnabling {
		fields["timeToLiveSpecification"] = core.MappingNodeFields(
			"enabled", core.MappingNodeFromBool(true),
			"attributeName", core.MappingNodeFromString(aws.ToString(ttl.AttributeName)),
		)
	}
}

func addPITRToSpec(backupsOutput *dynamodb.DescribeContinuousBackupsOutput, fields map[string]*core.MappingNode) {
	if backupsOutput.ContinuousBackupsDescription == nil {
		return
	}

	pitr := backupsOutput.ContinuousBackupsDescription.PointInTimeRecoveryDescription
	if pitr == nil {
		return
	}

	// Only add PITR spec if it's enabled
	if pitr.PointInTimeRecoveryStatus == types.PointInTimeRecoveryStatusEnabled {
		fields["pointInTimeRecoverySpecification"] = core.MappingNodeFields(
			"pointInTimeRecoveryEnabled", core.MappingNodeFromBool(true),
		)
	}
}

func addTagsToSpec(
	tagsOutput *dynamodb.ListTagsOfResourceOutput,
	fields map[string]*core.MappingNode,
	providerContext provider.Context,
) {
	if len(tagsOutput.Tags) == 0 {
		return
	}

	// Convert tags to a map for filtering
	tagsMap := make(map[string]string)
	for _, tag := range tagsOutput.Tags {
		tagsMap[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	// Filter out Bluelink provenance tags
	userTags := utils.UserTagsToMappingNode(tagsMap, providerContext)
	if userTags != nil && len(userTags.Items) > 0 {
		fields["tags"] = userTags
	}
}
