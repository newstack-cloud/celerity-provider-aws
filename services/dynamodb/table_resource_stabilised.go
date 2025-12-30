package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (a *dynamodbTableResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	dynamodbService, err := a.getDynamoDBService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	tableName := core.StringValue(
		input.ResourceSpec.Fields["tableName"],
	)

	describeOutput, err := dynamodbService.DescribeTable(
		ctx,
		&dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		},
	)
	if err != nil {
		return nil, err
	}

	stabilised := isTableStabilised(describeOutput.Table)

	return &provider.ResourceHasStabilisedOutput{
		Stabilised: stabilised,
	}, nil
}

func isTableStabilised(table *types.TableDescription) bool {
	if table == nil {
		return false
	}

	// Table must be ACTIVE
	if table.TableStatus != types.TableStatusActive {
		return false
	}

	// All GSIs must be ACTIVE
	for _, gsi := range table.GlobalSecondaryIndexes {
		if gsi.IndexStatus != types.IndexStatusActive {
			return false
		}
	}

	// All LSIs are always available once table is ACTIVE, no status to check

	return true
}
