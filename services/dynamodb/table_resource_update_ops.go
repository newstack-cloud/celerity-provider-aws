package dynamodb

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type tableUpdate struct {
	input *dynamodb.UpdateTableInput
}

func (u *tableUpdate) Name() string {
	return "table update"
}

func (u *tableUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasUpdates := changesToUpdateTableInput(
		saveOpCtx.ProviderUpstreamID,
		specData,
		changes,
	)
	u.input = input
	return hasUpdates, saveOpCtx, nil
}

func (u *tableUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	dynamodbService dynamodbservice.Service,
) (pluginutils.SaveOperationContext, error) {
	output, err := dynamodbService.UpdateTable(ctx, u.input)
	if err != nil {
		return saveOpCtx, err
	}

	saveOpCtx.Data["updateTableOutput"] = output
	return saveOpCtx, nil
}

type ttlUpdate struct {
	input *dynamodb.UpdateTimeToLiveInput
}

func (u *ttlUpdate) Name() string {
	return "TTL update"
}

func (u *ttlUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	tableName := saveOpCtx.ProviderUpstreamID
	input, hasUpdates := changesToUpdateTTLInput(tableName, specData, changes)
	u.input = input
	return hasUpdates, saveOpCtx, nil
}

func (u *ttlUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	dynamodbService dynamodbservice.Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := dynamodbService.UpdateTimeToLive(ctx, u.input)
	return saveOpCtx, err
}

type pitrUpdate struct {
	input *dynamodb.UpdateContinuousBackupsInput
}

func (u *pitrUpdate) Name() string {
	return "PITR update"
}

func (u *pitrUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	tableName := saveOpCtx.ProviderUpstreamID
	input, hasUpdates := changesToUpdatePITRInput(tableName, specData, changes)
	u.input = input
	return hasUpdates, saveOpCtx, nil
}

func (u *pitrUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	dynamodbService dynamodbservice.Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := dynamodbService.UpdateContinuousBackups(ctx, u.input)
	return saveOpCtx, err
}

type tagsUpdate struct {
	saveTagsInput   *dynamodb.TagResourceInput
	removeTagsInput *dynamodb.UntagResourceInput
	pathRoot        string
}

func (u *tagsUpdate) Name() string {
	return "tags update"
}

func (u *tagsUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	currentResourceStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	newTagsNode, _ := pluginutils.GetValueByPath(u.pathRoot, specData)
	currentTagsNode, _ := pluginutils.GetValueByPath(
		u.pathRoot,
		currentResourceStateSpecData,
	)

	// Get table ARN from context
	tableARN, _ := saveOpCtx.Data["tableARN"].(string)

	// Extract user tags from spec
	userTags := extractTagsMapFromNode(newTagsNode)

	// Merge Bluelink system tags with user-defined tags
	deployInput, ok := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
	if ok && deployInput != nil {
		userTags = mergeBluelinkTagsWithDynamoDBTags(deployInput, userTags)
	}

	input, hasUpdates := changesToTagUpdatesInputWithMergedTags(
		tableARN,
		userTags,
		currentTagsNode,
	)
	u.saveTagsInput = input.saveTagsInput
	u.removeTagsInput = input.removeTagsInput
	return hasUpdates, saveOpCtx, nil
}

func (u *tagsUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	dynamodbService dynamodbservice.Service,
) (pluginutils.SaveOperationContext, error) {
	if len(u.saveTagsInput.Tags) > 0 {
		_, err := dynamodbService.TagResource(ctx, u.saveTagsInput)
		if err != nil {
			return saveOpCtx, err
		}
	}

	if len(u.removeTagsInput.TagKeys) > 0 {
		_, err := dynamodbService.UntagResource(ctx, u.removeTagsInput)
		if err != nil {
			return saveOpCtx, err
		}
	}

	return saveOpCtx, nil
}

// Helper functions for update operations

func changesToUpdateTableInput(
	tableName string,
	specData *core.MappingNode,
	changes *provider.Changes,
) (*dynamodb.UpdateTableInput, bool) {
	modifiedFields := pluginutils.MergeFieldChanges(changes.ModifiedFields, changes.NewFields)
	input := &dynamodb.UpdateTableInput{
		TableName: aws.String(tableName),
	}

	hasUpdates := false

	// Check for billing mode changes
	if hasFieldWithPrefix(modifiedFields, "spec.billingMode") ||
		hasFieldWithPrefix(modifiedFields, "spec.provisionedThroughput") {
		if billingModeNode, exists := pluginutils.GetValueByPath("$.billingMode", specData); exists {
			billingMode := core.StringValue(billingModeNode)
			input.BillingMode = types.BillingMode(billingMode)
			hasUpdates = true

			if billingMode == string(types.BillingModeProvisioned) {
				if ptNode, ptExists := pluginutils.GetValueByPath("$.provisionedThroughput", specData); ptExists {
					input.ProvisionedThroughput = extractProvisionedThroughputForUpdate(ptNode)
				}
			}
		}
	}

	// Check for stream specification changes
	if hasFieldWithPrefix(modifiedFields, "spec.streamSpecification") {
		if streamSpecNode, exists := pluginutils.GetValueByPath("$.streamSpecification", specData); exists {
			input.StreamSpecification = extractStreamSpecificationForUpdate(streamSpecNode)
			hasUpdates = true
		}
	}

	// Check for deletion protection changes
	if hasFieldWithPrefix(modifiedFields, "spec.deletionProtectionEnabled") {
		if deletionProtNode, exists := pluginutils.GetValueByPath("$.deletionProtectionEnabled", specData); exists {
			input.DeletionProtectionEnabled = aws.Bool(core.BoolValue(deletionProtNode))
			hasUpdates = true
		}
	}

	// Check for table class changes
	if hasFieldWithPrefix(modifiedFields, "spec.tableClass") {
		if tableClassNode, exists := pluginutils.GetValueByPath("$.tableClass", specData); exists {
			input.TableClass = types.TableClass(core.StringValue(tableClassNode))
			hasUpdates = true
		}
	}

	return input, hasUpdates
}

func hasFieldWithPrefix(fields []provider.FieldChange, prefix string) bool {
	for _, field := range fields {
		if strings.HasPrefix(field.FieldPath, prefix) {
			return true
		}
	}
	return false
}

func extractProvisionedThroughputForUpdate(node *core.MappingNode) *types.ProvisionedThroughput {
	if node == nil || node.Fields == nil {
		return nil
	}

	pt := &types.ProvisionedThroughput{}
	if rcu, exists := node.Fields["readCapacityUnits"]; exists {
		pt.ReadCapacityUnits = aws.Int64(int64(core.IntValue(rcu)))
	}
	if wcu, exists := node.Fields["writeCapacityUnits"]; exists {
		pt.WriteCapacityUnits = aws.Int64(int64(core.IntValue(wcu)))
	}
	return pt
}

func extractStreamSpecificationForUpdate(node *core.MappingNode) *types.StreamSpecification {
	if node == nil || node.Fields == nil {
		return nil
	}

	spec := &types.StreamSpecification{}
	if enabled, exists := node.Fields["streamEnabled"]; exists {
		spec.StreamEnabled = aws.Bool(core.BoolValue(enabled))
	}
	if viewType, exists := node.Fields["streamViewType"]; exists {
		spec.StreamViewType = types.StreamViewType(core.StringValue(viewType))
	}
	return spec
}

func changesToUpdateTTLInput(
	tableName string,
	specData *core.MappingNode,
	changes *provider.Changes,
) (*dynamodb.UpdateTimeToLiveInput, bool) {
	modifiedFields := pluginutils.MergeFieldChanges(changes.ModifiedFields, changes.NewFields)

	// Check if TTL settings changed
	if !hasFieldWithPrefix(modifiedFields, "spec.timeToLiveSpecification") {
		return nil, false
	}

	ttlSpec, exists := pluginutils.GetValueByPath("$.timeToLiveSpecification", specData)
	if !exists || ttlSpec == nil || ttlSpec.Fields == nil {
		return nil, false
	}

	enabledNode, hasEnabled := ttlSpec.Fields["enabled"]
	attrNameNode, hasAttrName := ttlSpec.Fields["attributeName"]

	if !hasEnabled || !hasAttrName {
		return nil, false
	}

	return &dynamodb.UpdateTimeToLiveInput{
		TableName: aws.String(tableName),
		TimeToLiveSpecification: &types.TimeToLiveSpecification{
			Enabled:       aws.Bool(core.BoolValue(enabledNode)),
			AttributeName: aws.String(core.StringValue(attrNameNode)),
		},
	}, true
}

func changesToUpdatePITRInput(
	tableName string,
	specData *core.MappingNode,
	changes *provider.Changes,
) (*dynamodb.UpdateContinuousBackupsInput, bool) {
	modifiedFields := pluginutils.MergeFieldChanges(changes.ModifiedFields, changes.NewFields)

	// Check if PITR settings changed
	if !hasFieldWithPrefix(modifiedFields, "spec.pointInTimeRecoverySpecification") {
		return nil, false
	}

	pitrSpec, exists := pluginutils.GetValueByPath("$.pointInTimeRecoverySpecification", specData)
	if !exists || pitrSpec == nil || pitrSpec.Fields == nil {
		return nil, false
	}

	enabledNode, hasEnabled := pitrSpec.Fields["pointInTimeRecoveryEnabled"]
	if !hasEnabled {
		return nil, false
	}

	return &dynamodb.UpdateContinuousBackupsInput{
		TableName: aws.String(tableName),
		PointInTimeRecoverySpecification: &types.PointInTimeRecoverySpecification{
			PointInTimeRecoveryEnabled: aws.Bool(core.BoolValue(enabledNode)),
		},
	}, true
}

func extractTagsMapFromNode(node *core.MappingNode) map[string]string {
	tags := make(map[string]string)
	if node == nil || node.Items == nil {
		return tags
	}

	for _, item := range node.Items {
		key := core.StringValue(item.Fields["key"])
		value := core.StringValue(item.Fields["value"])
		tags[key] = value
	}
	return tags
}

type tagUpdatesInput struct {
	saveTagsInput   *dynamodb.TagResourceInput
	removeTagsInput *dynamodb.UntagResourceInput
}

func changesToTagUpdatesInputWithMergedTags(
	arn string,
	mergedTags map[string]string,
	currentTagsNode *core.MappingNode,
) (*tagUpdatesInput, bool) {
	removedTags := []string{}

	var currentTagsNodeItems []*core.MappingNode
	if currentTagsNode != nil {
		currentTagsNodeItems = currentTagsNode.Items
	}

	for _, item := range currentTagsNodeItems {
		key := core.StringValue(item.Fields["key"])
		if _, inNewTags := mergedTags[key]; !inNewTags {
			removedTags = append(removedTags, key)
		}
	}

	// Convert merged tags to DynamoDB tag format
	dynamoDBTags := make([]types.Tag, 0, len(mergedTags))
	for key, value := range mergedTags {
		dynamoDBTags = append(dynamoDBTags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	hasUpdates := len(dynamoDBTags) > 0 || len(removedTags) > 0

	return &tagUpdatesInput{
		saveTagsInput: &dynamodb.TagResourceInput{
			ResourceArn: aws.String(arn),
			Tags:        dynamoDBTags,
		},
		removeTagsInput: &dynamodb.UntagResourceInput{
			ResourceArn: aws.String(arn),
			TagKeys:     removedTags,
		},
	}, hasUpdates
}
