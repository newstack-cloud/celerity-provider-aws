package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// The timeout for waiting for a DynamoDB table to be available.
const tableWaitTimeout = 10 * time.Minute

func (a *dynamodbTableResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	dynamodbService, err := a.getDynamoDBService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[dynamodbservice.Service]{
		&tableCreate{},
	}

	saveOpCtx := pluginutils.SaveOperationContext{
		Data: map[string]any{
			"ResourceDeployInput": input,
		},
	}

	hasSavedValues, saveOpCtx, err := pluginutils.RunSaveOperations(
		ctx,
		saveOpCtx,
		createOperations,
		input,
		dynamodbService,
	)
	if err != nil {
		return nil, err
	}

	if !hasSavedValues {
		return nil, fmt.Errorf("no values were saved during table creation")
	}

	createTableOutput, ok := saveOpCtx.Data["createTableOutput"].(*dynamodb.CreateTableOutput)
	if !ok {
		return nil, fmt.Errorf("createTableOutput not found in save operation context")
	}

	tableName := createTableOutput.TableDescription.TableName

	// Wait for table to become ACTIVE before configuring TTL and PITR
	err = waitForTableActive(ctx, dynamodbService, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed waiting for table to become active: %w", err)
	}

	// Configure TTL if specified
	specData := pluginutils.GetResolvedResourceSpecData(input.Changes)
	err = configureTTLIfSpecified(ctx, dynamodbService, tableName, specData)
	if err != nil {
		return nil, fmt.Errorf("failed to configure TTL: %w", err)
	}

	// Configure Point-in-Time Recovery if specified
	err = configurePITRIfSpecified(ctx, dynamodbService, tableName, specData)
	if err != nil {
		return nil, fmt.Errorf("failed to configure point-in-time recovery: %w", err)
	}

	tableDesc := createTableOutput.TableDescription
	computedFields := map[string]*core.MappingNode{
		"spec.arn":     core.MappingNodeFromString(aws.ToString(tableDesc.TableArn)),
		"spec.tableId": core.MappingNodeFromString(aws.ToString(tableDesc.TableId)),
	}

	if tableDesc.LatestStreamArn != nil {
		computedFields["spec.latestStreamArn"] = core.MappingNodeFromString(
			aws.ToString(tableDesc.LatestStreamArn),
		)
	}

	if tableDesc.LatestStreamLabel != nil {
		computedFields["spec.latestStreamLabel"] = core.MappingNodeFromString(
			aws.ToString(tableDesc.LatestStreamLabel),
		)
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}

func waitForTableActive(
	ctx context.Context,
	service dynamodbservice.Service,
	tableName *string,
) error {
	waiter := dynamodb.NewTableExistsWaiter(service)
	return waiter.Wait(
		ctx,
		&dynamodb.DescribeTableInput{
			TableName: tableName,
		},
		tableWaitTimeout,
	)
}

func configureTTLIfSpecified(
	ctx context.Context,
	service dynamodbservice.Service,
	tableName *string,
	specData *core.MappingNode,
) error {
	ttlSpec, exists := pluginutils.GetValueByPath("$.timeToLiveSpecification", specData)
	if !exists || ttlSpec == nil || ttlSpec.Fields == nil {
		return nil
	}

	enabledNode, hasEnabled := ttlSpec.Fields["enabled"]
	attrNameNode, hasAttrName := ttlSpec.Fields["attributeName"]

	if !hasEnabled || !hasAttrName {
		return nil
	}

	enabled := core.BoolValue(enabledNode)
	attributeName := core.StringValue(attrNameNode)

	_, err := service.UpdateTimeToLive(ctx, &dynamodb.UpdateTimeToLiveInput{
		TableName: tableName,
		TimeToLiveSpecification: &types.TimeToLiveSpecification{
			Enabled:       aws.Bool(enabled),
			AttributeName: aws.String(attributeName),
		},
	})

	return err
}

func configurePITRIfSpecified(
	ctx context.Context,
	service dynamodbservice.Service,
	tableName *string,
	specData *core.MappingNode,
) error {
	pitrSpec, exists := pluginutils.GetValueByPath("$.pointInTimeRecoverySpecification", specData)
	if !exists || pitrSpec == nil || pitrSpec.Fields == nil {
		return nil
	}

	enabledNode, hasEnabled := pitrSpec.Fields["pointInTimeRecoveryEnabled"]
	if !hasEnabled {
		return nil
	}

	enabled := core.BoolValue(enabledNode)

	_, err := service.UpdateContinuousBackups(ctx, &dynamodb.UpdateContinuousBackupsInput{
		TableName: tableName,
		PointInTimeRecoverySpecification: &types.PointInTimeRecoverySpecification{
			PointInTimeRecoveryEnabled: aws.Bool(enabled),
		},
	})

	return err
}
