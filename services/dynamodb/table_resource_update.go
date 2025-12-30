package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (a *dynamodbTableResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	dynamodbService, err := a.getDynamoDBService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Get current table name and ARN from current state
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(input.Changes)
	tableNameNode, err := core.GetPathValue(
		"$.tableName",
		currentStateSpecData,
		core.MappingNodeMaxTraverseDepth,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get table name from current state: %w", err)
	}
	tableName := core.StringValue(tableNameNode)

	tableARNNode, err := core.GetPathValue(
		"$.arn",
		currentStateSpecData,
		core.MappingNodeMaxTraverseDepth,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get table ARN from current state: %w", err)
	}
	tableARN := core.StringValue(tableARNNode)

	updateOperations := []pluginutils.SaveOperation[dynamodbservice.Service]{
		&tableUpdate{},
		&ttlUpdate{},
		&pitrUpdate{},
		&tagsUpdate{
			pathRoot: "$.tags",
		},
	}

	hasUpdates, _, err := pluginutils.RunSaveOperations(
		ctx,
		pluginutils.SaveOperationContext{
			ProviderUpstreamID: tableName,
			Data: map[string]any{
				"ResourceDeployInput": input,
				"tableARN":            tableARN,
			},
		},
		updateOperations,
		input,
		dynamodbService,
	)
	if err != nil {
		return nil, err
	}

	if hasUpdates {
		// Wait for table to be active after updates
		err = waitForTableActive(ctx, dynamodbService, aws.String(tableName))
		if err != nil {
			return nil, fmt.Errorf("failed waiting for table to become active after update: %w", err)
		}

		// Get updated table description
		describeOutput, err := dynamodbService.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get updated table description: %w", err)
		}

		computedFields := extractComputedFieldsFromTableDescription(describeOutput.Table)
		return &provider.ResourceDeployOutput{
			ComputedFieldValues: computedFields,
		}, nil
	}

	// If no updates were made, return the current computed fields from the current state
	currentStateComputedFields := extractComputedFieldsFromCurrentState(currentStateSpecData)
	return &provider.ResourceDeployOutput{
		ComputedFieldValues: currentStateComputedFields,
	}, nil
}

func extractComputedFieldsFromTableDescription(
	tableDesc *types.TableDescription,
) map[string]*core.MappingNode {
	fields := map[string]*core.MappingNode{}
	if tableDesc == nil {
		return fields
	}

	if tableDesc.TableArn != nil {
		fields["spec.arn"] = core.MappingNodeFromString(aws.ToString(tableDesc.TableArn))
	}
	if tableDesc.TableId != nil {
		fields["spec.tableId"] = core.MappingNodeFromString(aws.ToString(tableDesc.TableId))
	}
	if tableDesc.LatestStreamArn != nil {
		fields["spec.latestStreamArn"] = core.MappingNodeFromString(
			aws.ToString(tableDesc.LatestStreamArn),
		)
	}
	if tableDesc.LatestStreamLabel != nil {
		fields["spec.latestStreamLabel"] = core.MappingNodeFromString(
			aws.ToString(tableDesc.LatestStreamLabel),
		)
	}

	return fields
}

func extractComputedFieldsFromCurrentState(
	currentStateSpecData *core.MappingNode,
) map[string]*core.MappingNode {
	fields := map[string]*core.MappingNode{}
	if v, ok := pluginutils.GetValueByPath("$.arn", currentStateSpecData); ok {
		fields["spec.arn"] = v
	}
	if v, ok := pluginutils.GetValueByPath("$.tableId", currentStateSpecData); ok {
		fields["spec.tableId"] = v
	}
	if v, ok := pluginutils.GetValueByPath("$.latestStreamArn", currentStateSpecData); ok {
		fields["spec.latestStreamArn"] = v
	}
	if v, ok := pluginutils.GetValueByPath("$.latestStreamLabel", currentStateSpecData); ok {
		fields["spec.latestStreamLabel"] = v
	}
	return fields
}
