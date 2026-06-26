package cloudcontrolmock

import (
	"context"

	awscc "github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

type cloudControlServiceMock struct {
	plugintestutils.MockCalls

	createResourceOutput *awscc.CreateResourceOutput
	createResourceError  error

	updateResourceOutput *awscc.UpdateResourceOutput
	updateResourceError  error

	getResourceOutput *awscc.GetResourceOutput
	getResourceError  error

	// listResourcesOutputs are returned in order, one per ListResources call, to
	// simulate pagination. When exhausted, the last output is returned.
	listResourcesOutputs []*awscc.ListResourcesOutput
	listResourcesCalls   int
	listResourcesError   error

	deleteResourceOutput *awscc.DeleteResourceOutput
	deleteResourceError  error

	getResourceRequestStatusOutput *awscc.GetResourceRequestStatusOutput
	getResourceRequestStatusError  error
}

// CreateCloudControlServiceMock creates a new Cloud Control service mock configured
// with the provided options.
func CreateCloudControlServiceMock(
	options ...CloudControlServiceMockOption,
) *cloudControlServiceMock {
	mock := &cloudControlServiceMock{}
	for _, option := range options {
		option(mock)
	}
	return mock
}

// CloudControlServiceMockOption configures a Cloud Control service mock.
type CloudControlServiceMockOption func(*cloudControlServiceMock)

func WithCreateResourceOutput(output *awscc.CreateResourceOutput) CloudControlServiceMockOption {
	return func(mock *cloudControlServiceMock) { mock.createResourceOutput = output }
}

func WithCreateResourceError(err error) CloudControlServiceMockOption {
	return func(mock *cloudControlServiceMock) { mock.createResourceError = err }
}

func WithUpdateResourceOutput(output *awscc.UpdateResourceOutput) CloudControlServiceMockOption {
	return func(mock *cloudControlServiceMock) { mock.updateResourceOutput = output }
}

func WithUpdateResourceError(err error) CloudControlServiceMockOption {
	return func(mock *cloudControlServiceMock) { mock.updateResourceError = err }
}

func WithGetResourceOutput(output *awscc.GetResourceOutput) CloudControlServiceMockOption {
	return func(mock *cloudControlServiceMock) { mock.getResourceOutput = output }
}

func WithGetResourceError(err error) CloudControlServiceMockOption {
	return func(mock *cloudControlServiceMock) { mock.getResourceError = err }
}

// WithListResourcesOutput sets a single ListResources output returned on every call.
func WithListResourcesOutput(output *awscc.ListResourcesOutput) CloudControlServiceMockOption {
	return func(mock *cloudControlServiceMock) {
		mock.listResourcesOutputs = []*awscc.ListResourcesOutput{output}
	}
}

// WithListResourcesOutputs sets a sequence of ListResources outputs returned in
// order, one per call, to simulate pagination.
func WithListResourcesOutputs(outputs ...*awscc.ListResourcesOutput) CloudControlServiceMockOption {
	return func(mock *cloudControlServiceMock) { mock.listResourcesOutputs = outputs }
}

func WithListResourcesError(err error) CloudControlServiceMockOption {
	return func(mock *cloudControlServiceMock) { mock.listResourcesError = err }
}

func WithDeleteResourceOutput(output *awscc.DeleteResourceOutput) CloudControlServiceMockOption {
	return func(mock *cloudControlServiceMock) { mock.deleteResourceOutput = output }
}

func WithDeleteResourceError(err error) CloudControlServiceMockOption {
	return func(mock *cloudControlServiceMock) { mock.deleteResourceError = err }
}

func WithGetResourceRequestStatusOutput(
	output *awscc.GetResourceRequestStatusOutput,
) CloudControlServiceMockOption {
	return func(mock *cloudControlServiceMock) { mock.getResourceRequestStatusOutput = output }
}

func WithGetResourceRequestStatusError(err error) CloudControlServiceMockOption {
	return func(mock *cloudControlServiceMock) { mock.getResourceRequestStatusError = err }
}

func (m *cloudControlServiceMock) CreateResource(
	ctx context.Context,
	params *awscc.CreateResourceInput,
	optFns ...func(*awscc.Options),
) (*awscc.CreateResourceOutput, error) {
	m.RegisterCall("CreateResource", params)
	if m.createResourceError != nil {
		return nil, m.createResourceError
	}
	return m.createResourceOutput, nil
}

func (m *cloudControlServiceMock) UpdateResource(
	ctx context.Context,
	params *awscc.UpdateResourceInput,
	optFns ...func(*awscc.Options),
) (*awscc.UpdateResourceOutput, error) {
	m.RegisterCall("UpdateResource", params)
	if m.updateResourceError != nil {
		return nil, m.updateResourceError
	}
	return m.updateResourceOutput, nil
}

func (m *cloudControlServiceMock) GetResource(
	ctx context.Context,
	params *awscc.GetResourceInput,
	optFns ...func(*awscc.Options),
) (*awscc.GetResourceOutput, error) {
	m.RegisterCall("GetResource", params)
	if m.getResourceError != nil {
		return nil, m.getResourceError
	}
	return m.getResourceOutput, nil
}

func (m *cloudControlServiceMock) ListResources(
	ctx context.Context,
	params *awscc.ListResourcesInput,
	optFns ...func(*awscc.Options),
) (*awscc.ListResourcesOutput, error) {
	m.RegisterCall("ListResources", params)
	if m.listResourcesError != nil {
		return nil, m.listResourcesError
	}
	if len(m.listResourcesOutputs) == 0 {
		return &awscc.ListResourcesOutput{}, nil
	}
	index := m.listResourcesCalls
	if index >= len(m.listResourcesOutputs) {
		index = len(m.listResourcesOutputs) - 1
	}
	m.listResourcesCalls++
	return m.listResourcesOutputs[index], nil
}

func (m *cloudControlServiceMock) DeleteResource(
	ctx context.Context,
	params *awscc.DeleteResourceInput,
	optFns ...func(*awscc.Options),
) (*awscc.DeleteResourceOutput, error) {
	m.RegisterCall("DeleteResource", params)
	if m.deleteResourceError != nil {
		return nil, m.deleteResourceError
	}
	return m.deleteResourceOutput, nil
}

func (m *cloudControlServiceMock) GetResourceRequestStatus(
	ctx context.Context,
	params *awscc.GetResourceRequestStatusInput,
	optFns ...func(*awscc.Options),
) (*awscc.GetResourceRequestStatusOutput, error) {
	m.RegisterCall("GetResourceRequestStatus", params)
	if m.getResourceRequestStatusError != nil {
		return nil, m.getResourceRequestStatusError
	}
	return m.getResourceRequestStatusOutput, nil
}
