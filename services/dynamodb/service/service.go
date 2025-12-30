package dynamodbservice

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the Amazon DynamoDB service
// used by the DynamoDB resource implementations.
type Service interface {
	// CreateTable creates a new table in DynamoDB.
	CreateTable(
		ctx context.Context,
		params *dynamodb.CreateTableInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.CreateTableOutput, error)

	// UpdateTable modifies the provisioned throughput settings, global secondary indexes,
	// or DynamoDB Streams settings for a given table.
	UpdateTable(
		ctx context.Context,
		params *dynamodb.UpdateTableInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.UpdateTableOutput, error)

	// DeleteTable deletes a table and all of its items.
	DeleteTable(
		ctx context.Context,
		params *dynamodb.DeleteTableInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.DeleteTableOutput, error)

	// DescribeTable returns information about the table.
	DescribeTable(
		ctx context.Context,
		params *dynamodb.DescribeTableInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.DescribeTableOutput, error)

	// UpdateTimeToLive enables or disables Time to Live (TTL) for the specified table.
	UpdateTimeToLive(
		ctx context.Context,
		params *dynamodb.UpdateTimeToLiveInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.UpdateTimeToLiveOutput, error)

	// DescribeTimeToLive gives a description of the Time to Live (TTL) status on the specified table.
	DescribeTimeToLive(
		ctx context.Context,
		params *dynamodb.DescribeTimeToLiveInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.DescribeTimeToLiveOutput, error)

	// UpdateContinuousBackups enables or disables point in time recovery for the specified table.
	UpdateContinuousBackups(
		ctx context.Context,
		params *dynamodb.UpdateContinuousBackupsInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.UpdateContinuousBackupsOutput, error)

	// DescribeContinuousBackups checks the status of continuous backups and point in time recovery
	// on the specified table.
	DescribeContinuousBackups(
		ctx context.Context,
		params *dynamodb.DescribeContinuousBackupsInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.DescribeContinuousBackupsOutput, error)

	// TagResource associates a set of tags with an Amazon DynamoDB resource.
	TagResource(
		ctx context.Context,
		params *dynamodb.TagResourceInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.TagResourceOutput, error)

	// UntagResource removes the association of tags from an Amazon DynamoDB resource.
	UntagResource(
		ctx context.Context,
		params *dynamodb.UntagResourceInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.UntagResourceOutput, error)

	// ListTagsOfResource lists all tags on an Amazon DynamoDB resource.
	ListTagsOfResource(
		ctx context.Context,
		params *dynamodb.ListTagsOfResourceInput,
		optFns ...func(*dynamodb.Options),
	) (*dynamodb.ListTagsOfResourceOutput, error)
}

// NewService creates a new instance of the AWS DynamoDB service
// based on the provided AWS configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return dynamodb.NewFromConfig(*awsConfig)
}
