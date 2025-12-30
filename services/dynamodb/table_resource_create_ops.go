package dynamodb

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type tableCreate struct {
	input *dynamodb.CreateTableInput
}

func (t *tableCreate) Name() string {
	return "create table"
}

func (t *tableCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues := changesToCreateTableInput(specData)

	deployInput, ok := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
	if ok && deployInput != nil {
		userTags := tagsFromCreateTableInput(input)
		mergedTags := utils.MergeBluelinkTagsWithUserTags(deployInput, userTags)
		input.Tags = mapToSortedDynamoDBTags(mergedTags)
	}

	t.input = input
	return hasValues, saveOpCtx, nil
}

func (t *tableCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	dynamodbService dynamodbservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	createTableOutput, err := dynamodbService.CreateTable(ctx, t.input)
	if err != nil {
		return saveOpCtx, err
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(createTableOutput.TableDescription.TableArn)
	newSaveOpCtx.Data["createTableOutput"] = createTableOutput

	return newSaveOpCtx, nil
}

func changesToCreateTableInput(specData *core.MappingNode) (*dynamodb.CreateTableInput, bool) {
	input := &dynamodb.CreateTableInput{}

	valueSetters := []*pluginutils.ValueSetter[*dynamodb.CreateTableInput]{
		pluginutils.NewValueSetter("$.tableName", setCreateTableName),
		pluginutils.NewValueSetter("$.attributeDefinitions", setCreateAttributeDefinitions),
		pluginutils.NewValueSetter("$.keySchema", setCreateKeySchema),
		pluginutils.NewValueSetter("$.billingMode", setCreateBillingMode),
		pluginutils.NewValueSetter("$.provisionedThroughput", setCreateProvisionedThroughput),
		pluginutils.NewValueSetter("$.globalSecondaryIndexes", setCreateGlobalSecondaryIndexes),
		pluginutils.NewValueSetter("$.localSecondaryIndexes", setCreateLocalSecondaryIndexes),
		pluginutils.NewValueSetter("$.streamSpecification", setCreateStreamSpecification),
		pluginutils.NewValueSetter("$.sseSpecification", setCreateSSESpecification),
		pluginutils.NewValueSetter("$.deletionProtectionEnabled", setCreateDeletionProtectionEnabled),
		pluginutils.NewValueSetter("$.tableClass", setCreateTableClass),
		pluginutils.NewValueSetter("$.tags", setCreateTags),
	}

	hasValues := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
		hasValues = hasValues || valueSetter.DidSet()
	}

	return input, hasValues
}

func setCreateTableName(value *core.MappingNode, input *dynamodb.CreateTableInput) {
	input.TableName = aws.String(core.StringValue(value))
}

func setCreateAttributeDefinitions(value *core.MappingNode, input *dynamodb.CreateTableInput) {
	if value == nil || value.Items == nil {
		return
	}

	definitions := make([]types.AttributeDefinition, 0, len(value.Items))
	for _, item := range value.Items {
		def := types.AttributeDefinition{}
		if nameNode, ok := item.Fields["attributeName"]; ok {
			def.AttributeName = aws.String(core.StringValue(nameNode))
		}
		if typeNode, ok := item.Fields["attributeType"]; ok {
			def.AttributeType = types.ScalarAttributeType(core.StringValue(typeNode))
		}
		definitions = append(definitions, def)
	}
	input.AttributeDefinitions = definitions
}

func setCreateKeySchema(value *core.MappingNode, input *dynamodb.CreateTableInput) {
	if value == nil || value.Items == nil {
		return
	}

	keySchema := make([]types.KeySchemaElement, 0, len(value.Items))
	for _, item := range value.Items {
		element := types.KeySchemaElement{}
		if nameNode, ok := item.Fields["attributeName"]; ok {
			element.AttributeName = aws.String(core.StringValue(nameNode))
		}
		if typeNode, ok := item.Fields["keyType"]; ok {
			element.KeyType = types.KeyType(core.StringValue(typeNode))
		}
		keySchema = append(keySchema, element)
	}
	input.KeySchema = keySchema
}

func setCreateBillingMode(value *core.MappingNode, input *dynamodb.CreateTableInput) {
	input.BillingMode = types.BillingMode(core.StringValue(value))
}

func setCreateProvisionedThroughput(value *core.MappingNode, input *dynamodb.CreateTableInput) {
	if value == nil || value.Fields == nil {
		return
	}

	throughput := &types.ProvisionedThroughput{}
	if rcuNode, ok := value.Fields["readCapacityUnits"]; ok {
		throughput.ReadCapacityUnits = aws.Int64(int64(core.IntValue(rcuNode)))
	}
	if wcuNode, ok := value.Fields["writeCapacityUnits"]; ok {
		throughput.WriteCapacityUnits = aws.Int64(int64(core.IntValue(wcuNode)))
	}
	input.ProvisionedThroughput = throughput
}

func setCreateGlobalSecondaryIndexes(value *core.MappingNode, input *dynamodb.CreateTableInput) {
	if value == nil || value.Items == nil {
		return
	}

	gsis := make([]types.GlobalSecondaryIndex, 0, len(value.Items))
	for _, item := range value.Items {
		gsi := types.GlobalSecondaryIndex{}

		if nameNode, ok := item.Fields["indexName"]; ok {
			gsi.IndexName = aws.String(core.StringValue(nameNode))
		}

		if keySchemaNode, ok := item.Fields["keySchema"]; ok && keySchemaNode.Items != nil {
			gsi.KeySchema = extractKeySchema(keySchemaNode)
		}

		if projectionNode, ok := item.Fields["projection"]; ok && projectionNode.Fields != nil {
			gsi.Projection = extractProjection(projectionNode)
		}

		if throughputNode, ok := item.Fields["provisionedThroughput"]; ok && throughputNode.Fields != nil {
			gsi.ProvisionedThroughput = extractProvisionedThroughput(throughputNode)
		}

		gsis = append(gsis, gsi)
	}
	input.GlobalSecondaryIndexes = gsis
}

func setCreateLocalSecondaryIndexes(value *core.MappingNode, input *dynamodb.CreateTableInput) {
	if value == nil || value.Items == nil {
		return
	}

	lsis := make([]types.LocalSecondaryIndex, 0, len(value.Items))
	for _, item := range value.Items {
		lsi := types.LocalSecondaryIndex{}

		if nameNode, ok := item.Fields["indexName"]; ok {
			lsi.IndexName = aws.String(core.StringValue(nameNode))
		}

		if keySchemaNode, ok := item.Fields["keySchema"]; ok && keySchemaNode.Items != nil {
			lsi.KeySchema = extractKeySchema(keySchemaNode)
		}

		if projectionNode, ok := item.Fields["projection"]; ok && projectionNode.Fields != nil {
			lsi.Projection = extractProjection(projectionNode)
		}

		lsis = append(lsis, lsi)
	}
	input.LocalSecondaryIndexes = lsis
}

func setCreateStreamSpecification(value *core.MappingNode, input *dynamodb.CreateTableInput) {
	if value == nil || value.Fields == nil {
		return
	}

	spec := &types.StreamSpecification{}
	if enabledNode, ok := value.Fields["streamEnabled"]; ok {
		spec.StreamEnabled = aws.Bool(core.BoolValue(enabledNode))
	}
	if viewTypeNode, ok := value.Fields["streamViewType"]; ok {
		spec.StreamViewType = types.StreamViewType(core.StringValue(viewTypeNode))
	}
	input.StreamSpecification = spec
}

func setCreateSSESpecification(value *core.MappingNode, input *dynamodb.CreateTableInput) {
	if value == nil || value.Fields == nil {
		return
	}

	spec := &types.SSESpecification{}
	if enabledNode, ok := value.Fields["enabled"]; ok {
		spec.Enabled = aws.Bool(core.BoolValue(enabledNode))
	}
	if sseTypeNode, ok := value.Fields["sseType"]; ok {
		spec.SSEType = types.SSEType(core.StringValue(sseTypeNode))
	}
	if kmsKeyNode, ok := value.Fields["kmsMasterKeyId"]; ok {
		spec.KMSMasterKeyId = aws.String(core.StringValue(kmsKeyNode))
	}
	input.SSESpecification = spec
}

func setCreateDeletionProtectionEnabled(value *core.MappingNode, input *dynamodb.CreateTableInput) {
	input.DeletionProtectionEnabled = aws.Bool(core.BoolValue(value))
}

func setCreateTableClass(value *core.MappingNode, input *dynamodb.CreateTableInput) {
	input.TableClass = types.TableClass(core.StringValue(value))
}

func setCreateTags(value *core.MappingNode, input *dynamodb.CreateTableInput) {
	if value == nil || value.Items == nil {
		return
	}
	input.Tags = mappingNodeToSortedDynamoDBTags(value)
}

// Helper functions

func extractKeySchema(node *core.MappingNode) []types.KeySchemaElement {
	elements := make([]types.KeySchemaElement, 0, len(node.Items))
	for _, item := range node.Items {
		element := types.KeySchemaElement{}
		if nameNode, ok := item.Fields["attributeName"]; ok {
			element.AttributeName = aws.String(core.StringValue(nameNode))
		}
		if typeNode, ok := item.Fields["keyType"]; ok {
			element.KeyType = types.KeyType(core.StringValue(typeNode))
		}
		elements = append(elements, element)
	}
	return elements
}

func extractProjection(node *core.MappingNode) *types.Projection {
	projection := &types.Projection{}
	if typeNode, ok := node.Fields["projectionType"]; ok {
		projection.ProjectionType = types.ProjectionType(core.StringValue(typeNode))
	}
	if nonKeyAttrsNode, ok := node.Fields["nonKeyAttributes"]; ok && nonKeyAttrsNode.Items != nil {
		attrs := make([]string, 0, len(nonKeyAttrsNode.Items))
		for _, item := range nonKeyAttrsNode.Items {
			attrs = append(attrs, core.StringValue(item))
		}
		projection.NonKeyAttributes = attrs
	}
	return projection
}

func extractProvisionedThroughput(node *core.MappingNode) *types.ProvisionedThroughput {
	throughput := &types.ProvisionedThroughput{}
	if rcuNode, ok := node.Fields["readCapacityUnits"]; ok {
		throughput.ReadCapacityUnits = aws.Int64(int64(core.IntValue(rcuNode)))
	}
	if wcuNode, ok := node.Fields["writeCapacityUnits"]; ok {
		throughput.WriteCapacityUnits = aws.Int64(int64(core.IntValue(wcuNode)))
	}
	return throughput
}

func tagsFromCreateTableInput(input *dynamodb.CreateTableInput) map[string]string {
	tags := make(map[string]string)
	for _, tag := range input.Tags {
		if tag.Key != nil && tag.Value != nil {
			tags[*tag.Key] = *tag.Value
		}
	}
	return tags
}

func mapToSortedDynamoDBTags(tagsMap map[string]string) []types.Tag {
	keys := make([]string, 0, len(tagsMap))
	for k := range tagsMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tags := make([]types.Tag, 0, len(tagsMap))
	for _, k := range keys {
		tags = append(tags, types.Tag{
			Key:   aws.String(k),
			Value: aws.String(tagsMap[k]),
		})
	}
	return tags
}

func mappingNodeToSortedDynamoDBTags(value *core.MappingNode) []types.Tag {
	if value == nil || value.Items == nil {
		return nil
	}

	tagsMap := make(map[string]string)
	for _, item := range value.Items {
		key := core.StringValue(item.Fields["key"])
		val := core.StringValue(item.Fields["value"])
		tagsMap[key] = val
	}

	return mapToSortedDynamoDBTags(tagsMap)
}
