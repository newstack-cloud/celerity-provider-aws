package dynamodbmock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

type dynamodbServiceMock struct {
	plugintestutils.MockCalls

	// CreateTable-related mock fields
	createTableOutput *dynamodb.CreateTableOutput
	createTableError  error

	// UpdateTable-related mock fields
	updateTableOutput *dynamodb.UpdateTableOutput
	updateTableError  error

	// DeleteTable-related mock fields
	deleteTableOutput *dynamodb.DeleteTableOutput
	deleteTableError  error

	// DescribeTable-related mock fields
	describeTableOutput *dynamodb.DescribeTableOutput
	describeTableError  error

	// UpdateTimeToLive-related mock fields
	updateTimeToLiveOutput *dynamodb.UpdateTimeToLiveOutput
	updateTimeToLiveError  error

	// DescribeTimeToLive-related mock fields
	describeTimeToLiveOutput *dynamodb.DescribeTimeToLiveOutput
	describeTimeToLiveError  error

	// UpdateContinuousBackups-related mock fields
	updateContinuousBackupsOutput *dynamodb.UpdateContinuousBackupsOutput
	updateContinuousBackupsError  error

	// DescribeContinuousBackups-related mock fields
	describeContinuousBackupsOutput *dynamodb.DescribeContinuousBackupsOutput
	describeContinuousBackupsError  error

	// TagResource-related mock fields
	tagResourceOutput *dynamodb.TagResourceOutput
	tagResourceError  error

	// UntagResource-related mock fields
	untagResourceOutput *dynamodb.UntagResourceOutput
	untagResourceError  error

	// ListTagsOfResource-related mock fields
	listTagsOfResourceOutput *dynamodb.ListTagsOfResourceOutput
	listTagsOfResourceError  error

	// UpdateTableReplicaAutoScaling-related mock fields
	updateTableReplicaAutoScalingOutput *dynamodb.UpdateTableReplicaAutoScalingOutput
	updateTableReplicaAutoScalingError  error

	// DescribeTableReplicaAutoScaling-related mock fields
	describeTableReplicaAutoScalingOutput *dynamodb.DescribeTableReplicaAutoScalingOutput
	describeTableReplicaAutoScalingError  error
}

// CreateDynamoDBServiceMock creates a new instance of the DynamoDB service mock with the provided options.
func CreateDynamoDBServiceMock(options ...DynamoDBServiceMockOption) *dynamodbServiceMock {
	mock := &dynamodbServiceMock{}
	for _, option := range options {
		option(mock)
	}
	return mock
}

// DynamoDBServiceMockOption is a function type for configuring the DynamoDB service mock.
type DynamoDBServiceMockOption func(*dynamodbServiceMock)

// WithCreateTableOutput sets the mock output for CreateTable.
func WithCreateTableOutput(output *dynamodb.CreateTableOutput) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.createTableOutput = output
	}
}

// WithCreateTableError sets the mock error for CreateTable.
func WithCreateTableError(err error) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.createTableError = err
	}
}

// WithUpdateTableOutput sets the mock output for UpdateTable.
func WithUpdateTableOutput(output *dynamodb.UpdateTableOutput) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.updateTableOutput = output
	}
}

// WithUpdateTableError sets the mock error for UpdateTable.
func WithUpdateTableError(err error) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.updateTableError = err
	}
}

// WithDeleteTableOutput sets the mock output for DeleteTable.
func WithDeleteTableOutput(output *dynamodb.DeleteTableOutput) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.deleteTableOutput = output
	}
}

// WithDeleteTableError sets the mock error for DeleteTable.
func WithDeleteTableError(err error) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.deleteTableError = err
	}
}

// WithDescribeTableOutput sets the mock output for DescribeTable.
func WithDescribeTableOutput(output *dynamodb.DescribeTableOutput) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.describeTableOutput = output
	}
}

// WithDescribeTableError sets the mock error for DescribeTable.
func WithDescribeTableError(err error) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.describeTableError = err
	}
}

// WithUpdateTimeToLiveOutput sets the mock output for UpdateTimeToLive.
func WithUpdateTimeToLiveOutput(output *dynamodb.UpdateTimeToLiveOutput) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.updateTimeToLiveOutput = output
	}
}

// WithUpdateTimeToLiveError sets the mock error for UpdateTimeToLive.
func WithUpdateTimeToLiveError(err error) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.updateTimeToLiveError = err
	}
}

// WithDescribeTimeToLiveOutput sets the mock output for DescribeTimeToLive.
func WithDescribeTimeToLiveOutput(output *dynamodb.DescribeTimeToLiveOutput) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.describeTimeToLiveOutput = output
	}
}

// WithDescribeTimeToLiveError sets the mock error for DescribeTimeToLive.
func WithDescribeTimeToLiveError(err error) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.describeTimeToLiveError = err
	}
}

// WithUpdateContinuousBackupsOutput sets the mock output for UpdateContinuousBackups.
func WithUpdateContinuousBackupsOutput(output *dynamodb.UpdateContinuousBackupsOutput) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.updateContinuousBackupsOutput = output
	}
}

// WithUpdateContinuousBackupsError sets the mock error for UpdateContinuousBackups.
func WithUpdateContinuousBackupsError(err error) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.updateContinuousBackupsError = err
	}
}

// WithDescribeContinuousBackupsOutput sets the mock output for DescribeContinuousBackups.
func WithDescribeContinuousBackupsOutput(output *dynamodb.DescribeContinuousBackupsOutput) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.describeContinuousBackupsOutput = output
	}
}

// WithDescribeContinuousBackupsError sets the mock error for DescribeContinuousBackups.
func WithDescribeContinuousBackupsError(err error) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.describeContinuousBackupsError = err
	}
}

// WithTagResourceOutput sets the mock output for TagResource.
func WithTagResourceOutput(output *dynamodb.TagResourceOutput) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.tagResourceOutput = output
	}
}

// WithTagResourceError sets the mock error for TagResource.
func WithTagResourceError(err error) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.tagResourceError = err
	}
}

// WithUntagResourceOutput sets the mock output for UntagResource.
func WithUntagResourceOutput(output *dynamodb.UntagResourceOutput) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.untagResourceOutput = output
	}
}

// WithUntagResourceError sets the mock error for UntagResource.
func WithUntagResourceError(err error) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.untagResourceError = err
	}
}

// WithListTagsOfResourceOutput sets the mock output for ListTagsOfResource.
func WithListTagsOfResourceOutput(output *dynamodb.ListTagsOfResourceOutput) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.listTagsOfResourceOutput = output
	}
}

// WithListTagsOfResourceError sets the mock error for ListTagsOfResource.
func WithListTagsOfResourceError(err error) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.listTagsOfResourceError = err
	}
}

// WithUpdateTableReplicaAutoScalingOutput sets the mock output for UpdateTableReplicaAutoScaling.
func WithUpdateTableReplicaAutoScalingOutput(
	output *dynamodb.UpdateTableReplicaAutoScalingOutput,
) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.updateTableReplicaAutoScalingOutput = output
	}
}

// WithUpdateTableReplicaAutoScalingError sets the mock error for UpdateTableReplicaAutoScaling.
func WithUpdateTableReplicaAutoScalingError(err error) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.updateTableReplicaAutoScalingError = err
	}
}

// WithDescribeTableReplicaAutoScalingOutput sets the mock output for DescribeTableReplicaAutoScaling.
func WithDescribeTableReplicaAutoScalingOutput(
	output *dynamodb.DescribeTableReplicaAutoScalingOutput,
) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.describeTableReplicaAutoScalingOutput = output
	}
}

// WithDescribeTableReplicaAutoScalingError sets the mock error for DescribeTableReplicaAutoScaling.
func WithDescribeTableReplicaAutoScalingError(err error) DynamoDBServiceMockOption {
	return func(mock *dynamodbServiceMock) {
		mock.describeTableReplicaAutoScalingError = err
	}
}

// CreateTable implements the DynamoDB service interface.
func (m *dynamodbServiceMock) CreateTable(
	ctx context.Context,
	params *dynamodb.CreateTableInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.CreateTableOutput, error) {
	m.RegisterCall("CreateTable", params)
	if m.createTableError != nil {
		return nil, m.createTableError
	}
	return m.createTableOutput, nil
}

// UpdateTable implements the DynamoDB service interface.
func (m *dynamodbServiceMock) UpdateTable(
	ctx context.Context,
	params *dynamodb.UpdateTableInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.UpdateTableOutput, error) {
	m.RegisterCall("UpdateTable", params)
	if m.updateTableError != nil {
		return nil, m.updateTableError
	}
	return m.updateTableOutput, nil
}

// DeleteTable implements the DynamoDB service interface.
func (m *dynamodbServiceMock) DeleteTable(
	ctx context.Context,
	params *dynamodb.DeleteTableInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.DeleteTableOutput, error) {
	m.RegisterCall("DeleteTable", params)
	if m.deleteTableError != nil {
		return nil, m.deleteTableError
	}
	return m.deleteTableOutput, nil
}

// DescribeTable implements the DynamoDB service interface.
func (m *dynamodbServiceMock) DescribeTable(
	ctx context.Context,
	params *dynamodb.DescribeTableInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.DescribeTableOutput, error) {
	m.RegisterCall("DescribeTable", params)
	if m.describeTableError != nil {
		return nil, m.describeTableError
	}
	return m.describeTableOutput, nil
}

// UpdateTimeToLive implements the DynamoDB service interface.
func (m *dynamodbServiceMock) UpdateTimeToLive(
	ctx context.Context,
	params *dynamodb.UpdateTimeToLiveInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.UpdateTimeToLiveOutput, error) {
	m.RegisterCall("UpdateTimeToLive", params)
	if m.updateTimeToLiveError != nil {
		return nil, m.updateTimeToLiveError
	}
	return m.updateTimeToLiveOutput, nil
}

// DescribeTimeToLive implements the DynamoDB service interface.
func (m *dynamodbServiceMock) DescribeTimeToLive(
	ctx context.Context,
	params *dynamodb.DescribeTimeToLiveInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.DescribeTimeToLiveOutput, error) {
	m.RegisterCall("DescribeTimeToLive", params)
	if m.describeTimeToLiveError != nil {
		return nil, m.describeTimeToLiveError
	}
	return m.describeTimeToLiveOutput, nil
}

// UpdateContinuousBackups implements the DynamoDB service interface.
func (m *dynamodbServiceMock) UpdateContinuousBackups(
	ctx context.Context,
	params *dynamodb.UpdateContinuousBackupsInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.UpdateContinuousBackupsOutput, error) {
	m.RegisterCall("UpdateContinuousBackups", params)
	if m.updateContinuousBackupsError != nil {
		return nil, m.updateContinuousBackupsError
	}
	return m.updateContinuousBackupsOutput, nil
}

// DescribeContinuousBackups implements the DynamoDB service interface.
func (m *dynamodbServiceMock) DescribeContinuousBackups(
	ctx context.Context,
	params *dynamodb.DescribeContinuousBackupsInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.DescribeContinuousBackupsOutput, error) {
	m.RegisterCall("DescribeContinuousBackups", params)
	if m.describeContinuousBackupsError != nil {
		return nil, m.describeContinuousBackupsError
	}
	return m.describeContinuousBackupsOutput, nil
}

// TagResource implements the DynamoDB service interface.
func (m *dynamodbServiceMock) TagResource(
	ctx context.Context,
	params *dynamodb.TagResourceInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.TagResourceOutput, error) {
	m.RegisterCall("TagResource", params)
	if m.tagResourceError != nil {
		return nil, m.tagResourceError
	}
	return m.tagResourceOutput, nil
}

// UntagResource implements the DynamoDB service interface.
func (m *dynamodbServiceMock) UntagResource(
	ctx context.Context,
	params *dynamodb.UntagResourceInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.UntagResourceOutput, error) {
	m.RegisterCall("UntagResource", params)
	if m.untagResourceError != nil {
		return nil, m.untagResourceError
	}
	return m.untagResourceOutput, nil
}

// ListTagsOfResource implements the DynamoDB service interface.
func (m *dynamodbServiceMock) ListTagsOfResource(
	ctx context.Context,
	params *dynamodb.ListTagsOfResourceInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.ListTagsOfResourceOutput, error) {
	m.RegisterCall("ListTagsOfResource", params)
	if m.listTagsOfResourceError != nil {
		return nil, m.listTagsOfResourceError
	}
	return m.listTagsOfResourceOutput, nil
}

// UpdateTableReplicaAutoScaling implements the DynamoDB service interface.
func (m *dynamodbServiceMock) UpdateTableReplicaAutoScaling(
	ctx context.Context,
	params *dynamodb.UpdateTableReplicaAutoScalingInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.UpdateTableReplicaAutoScalingOutput, error) {
	m.RegisterCall("UpdateTableReplicaAutoScaling", params)
	if m.updateTableReplicaAutoScalingError != nil {
		return nil, m.updateTableReplicaAutoScalingError
	}
	return m.updateTableReplicaAutoScalingOutput, nil
}

// DescribeTableReplicaAutoScaling implements the DynamoDB service interface.
func (m *dynamodbServiceMock) DescribeTableReplicaAutoScaling(
	ctx context.Context,
	params *dynamodb.DescribeTableReplicaAutoScalingInput,
	optFns ...func(*dynamodb.Options),
) (*dynamodb.DescribeTableReplicaAutoScalingOutput, error) {
	m.RegisterCall("DescribeTableReplicaAutoScaling", params)
	if m.describeTableReplicaAutoScalingError != nil {
		return nil, m.describeTableReplicaAutoScalingError
	}
	return m.describeTableReplicaAutoScalingOutput, nil
}

// CreateDynamoDBServiceMockFactory creates a factory function that returns the mock service.
// This is used for dependency injection in tests.
func CreateDynamoDBServiceMockFactory(
	options ...DynamoDBServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
	mock := CreateDynamoDBServiceMock(options...)
	return func(awsConfig *aws.Config, providerContext provider.Context) dynamodbservice.Service {
		return mock
	}
}
