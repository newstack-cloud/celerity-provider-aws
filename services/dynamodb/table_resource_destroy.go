package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (a *dynamodbTableResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	dynamodbService, err := a.getDynamoDBService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	tableName := core.StringValue(
		input.ResourceState.SpecData.Fields["tableName"],
	)

	_, err = dynamodbService.DeleteTable(
		ctx,
		&dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		},
	)

	return err
}
