package ssmmock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

type ssmServiceMock struct {
	plugintestutils.MockCalls

	putParameterOutput *ssm.PutParameterOutput
	putParameterError  error

	getParameterOutput *ssm.GetParameterOutput
	getParameterError  error

	deleteParameterOutput *ssm.DeleteParameterOutput
	deleteParameterError  error

	describeParametersOutput *ssm.DescribeParametersOutput
	describeParametersError  error

	addTagsToResourceOutput *ssm.AddTagsToResourceOutput
	addTagsToResourceError  error

	removeTagsFromResourceOutput *ssm.RemoveTagsFromResourceOutput
	removeTagsFromResourceError  error

	listTagsForResourceOutput *ssm.ListTagsForResourceOutput
	listTagsForResourceError  error
}

// CreateSSMServiceMock creates a new instance of the SSM service mock with the provided options.
func CreateSSMServiceMock(options ...SSMServiceMockOption) *ssmServiceMock {
	mock := &ssmServiceMock{}
	for _, option := range options {
		option(mock)
	}
	return mock
}

// CreateSSMServiceMockFactory returns a service factory bound to a single mock instance.
func CreateSSMServiceMockFactory(
	options ...SSMServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) ssmservice.Service {
	mock := CreateSSMServiceMock(options...)
	return func(awsConfig *aws.Config, providerContext provider.Context) ssmservice.Service {
		return mock
	}
}

// SSMServiceMockOption is a function type for configuring the SSM service mock.
type SSMServiceMockOption func(*ssmServiceMock)

func WithPutParameterOutput(output *ssm.PutParameterOutput) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.putParameterOutput = output }
}

func WithPutParameterError(err error) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.putParameterError = err }
}

func WithGetParameterOutput(output *ssm.GetParameterOutput) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.getParameterOutput = output }
}

func WithGetParameterError(err error) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.getParameterError = err }
}

func WithDeleteParameterOutput(output *ssm.DeleteParameterOutput) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.deleteParameterOutput = output }
}

func WithDeleteParameterError(err error) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.deleteParameterError = err }
}

func WithDescribeParametersOutput(output *ssm.DescribeParametersOutput) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.describeParametersOutput = output }
}

func WithDescribeParametersError(err error) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.describeParametersError = err }
}

func WithAddTagsToResourceOutput(output *ssm.AddTagsToResourceOutput) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.addTagsToResourceOutput = output }
}

func WithAddTagsToResourceError(err error) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.addTagsToResourceError = err }
}

func WithRemoveTagsFromResourceOutput(output *ssm.RemoveTagsFromResourceOutput) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.removeTagsFromResourceOutput = output }
}

func WithRemoveTagsFromResourceError(err error) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.removeTagsFromResourceError = err }
}

func WithListTagsForResourceOutput(output *ssm.ListTagsForResourceOutput) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.listTagsForResourceOutput = output }
}

func WithListTagsForResourceError(err error) SSMServiceMockOption {
	return func(mock *ssmServiceMock) { mock.listTagsForResourceError = err }
}

func (m *ssmServiceMock) PutParameter(
	ctx context.Context,
	params *ssm.PutParameterInput,
	optFns ...func(*ssm.Options),
) (*ssm.PutParameterOutput, error) {
	m.RegisterCall(ctx, params)
	if m.putParameterError != nil {
		return nil, m.putParameterError
	}
	return m.putParameterOutput, nil
}

func (m *ssmServiceMock) GetParameter(
	ctx context.Context,
	params *ssm.GetParameterInput,
	optFns ...func(*ssm.Options),
) (*ssm.GetParameterOutput, error) {
	m.RegisterCall(ctx, params)
	if m.getParameterError != nil {
		return nil, m.getParameterError
	}
	return m.getParameterOutput, nil
}

func (m *ssmServiceMock) DeleteParameter(
	ctx context.Context,
	params *ssm.DeleteParameterInput,
	optFns ...func(*ssm.Options),
) (*ssm.DeleteParameterOutput, error) {
	m.RegisterCall(ctx, params)
	if m.deleteParameterError != nil {
		return nil, m.deleteParameterError
	}
	return m.deleteParameterOutput, nil
}

func (m *ssmServiceMock) DescribeParameters(
	ctx context.Context,
	params *ssm.DescribeParametersInput,
	optFns ...func(*ssm.Options),
) (*ssm.DescribeParametersOutput, error) {
	m.RegisterCall(ctx, params)
	if m.describeParametersError != nil {
		return nil, m.describeParametersError
	}
	return m.describeParametersOutput, nil
}

func (m *ssmServiceMock) AddTagsToResource(
	ctx context.Context,
	params *ssm.AddTagsToResourceInput,
	optFns ...func(*ssm.Options),
) (*ssm.AddTagsToResourceOutput, error) {
	m.RegisterCall(ctx, params)
	if m.addTagsToResourceError != nil {
		return nil, m.addTagsToResourceError
	}
	return m.addTagsToResourceOutput, nil
}

func (m *ssmServiceMock) RemoveTagsFromResource(
	ctx context.Context,
	params *ssm.RemoveTagsFromResourceInput,
	optFns ...func(*ssm.Options),
) (*ssm.RemoveTagsFromResourceOutput, error) {
	m.RegisterCall(ctx, params)
	if m.removeTagsFromResourceError != nil {
		return nil, m.removeTagsFromResourceError
	}
	return m.removeTagsFromResourceOutput, nil
}

func (m *ssmServiceMock) ListTagsForResource(
	ctx context.Context,
	params *ssm.ListTagsForResourceInput,
	optFns ...func(*ssm.Options),
) (*ssm.ListTagsForResourceOutput, error) {
	m.RegisterCall(ctx, params)
	if m.listTagsForResourceError != nil {
		return nil, m.listTagsForResourceError
	}
	return m.listTagsForResourceOutput, nil
}
