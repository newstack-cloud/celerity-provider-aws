package sqsmock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

type sqsServiceMock struct {
	plugintestutils.MockCalls

	// CreateQueue-related mock fields
	createQueueOutput *sqs.CreateQueueOutput
	createQueueError  error

	// GetQueueUrl-related mock fields
	getQueueUrlOutput *sqs.GetQueueUrlOutput
	getQueueUrlError  error

	// GetQueueAttributes-related mock fields
	getQueueAttributesOutput *sqs.GetQueueAttributesOutput
	getQueueAttributesError  error

	// SetQueueAttributes-related mock fields
	setQueueAttributesOutput *sqs.SetQueueAttributesOutput
	setQueueAttributesError  error

	// DeleteQueue-related mock fields
	deleteQueueOutput *sqs.DeleteQueueOutput
	deleteQueueError  error

	// TagQueue-related mock fields
	tagQueueOutput *sqs.TagQueueOutput
	tagQueueError  error

	// UntagQueue-related mock fields
	untagQueueOutput *sqs.UntagQueueOutput
	untagQueueError  error

	// ListQueueTags-related mock fields
	listQueueTagsOutput *sqs.ListQueueTagsOutput
	listQueueTagsError  error
}

// CreateSQSServiceMock creates a new instance of the SQS service mock with the provided options.
func CreateSQSServiceMock(options ...SqsServiceMockOption) *sqsServiceMock {
	mock := &sqsServiceMock{}
	for _, option := range options {
		option(mock)
	}
	return mock
}

// SqsServiceMockOption is a function type for configuring the SQS service mock.
type SqsServiceMockOption func(*sqsServiceMock)

// WithCreateQueueOutput sets the mock output for CreateQueue.
func WithCreateQueueOutput(output *sqs.CreateQueueOutput) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.createQueueOutput = output
	}
}

// WithCreateQueueError sets the mock error for CreateQueue.
func WithCreateQueueError(err error) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.createQueueError = err
	}
}

// WithGetQueueUrlOutput sets the mock output for GetQueueUrl.
func WithGetQueueUrlOutput(output *sqs.GetQueueUrlOutput) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.getQueueUrlOutput = output
	}
}

// WithGetQueueUrlError sets the mock error for GetQueueUrl.
func WithGetQueueUrlError(err error) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.getQueueUrlError = err
	}
}

// WithGetQueueAttributesOutput sets the mock output for GetQueueAttributes.
func WithGetQueueAttributesOutput(output *sqs.GetQueueAttributesOutput) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.getQueueAttributesOutput = output
	}
}

// WithGetQueueAttributesError sets the mock error for GetQueueAttributes.
func WithGetQueueAttributesError(err error) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.getQueueAttributesError = err
	}
}

// WithSetQueueAttributesOutput sets the mock output for SetQueueAttributes.
func WithSetQueueAttributesOutput(output *sqs.SetQueueAttributesOutput) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.setQueueAttributesOutput = output
	}
}

// WithSetQueueAttributesError sets the mock error for SetQueueAttributes.
func WithSetQueueAttributesError(err error) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.setQueueAttributesError = err
	}
}

// WithDeleteQueueOutput sets the mock output for DeleteQueue.
func WithDeleteQueueOutput(output *sqs.DeleteQueueOutput) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.deleteQueueOutput = output
	}
}

// WithDeleteQueueError sets the mock error for DeleteQueue.
func WithDeleteQueueError(err error) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.deleteQueueError = err
	}
}

// WithTagQueueOutput sets the mock output for TagQueue.
func WithTagQueueOutput(output *sqs.TagQueueOutput) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.tagQueueOutput = output
	}
}

// WithTagQueueError sets the mock error for TagQueue.
func WithTagQueueError(err error) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.tagQueueError = err
	}
}

// WithUntagQueueOutput sets the mock output for UntagQueue.
func WithUntagQueueOutput(output *sqs.UntagQueueOutput) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.untagQueueOutput = output
	}
}

// WithUntagQueueError sets the mock error for UntagQueue.
func WithUntagQueueError(err error) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.untagQueueError = err
	}
}

// WithListQueueTagsOutput sets the mock output for ListQueueTags.
func WithListQueueTagsOutput(output *sqs.ListQueueTagsOutput) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.listQueueTagsOutput = output
	}
}

// WithListQueueTagsError sets the mock error for ListQueueTags.
func WithListQueueTagsError(err error) SqsServiceMockOption {
	return func(mock *sqsServiceMock) {
		mock.listQueueTagsError = err
	}
}

// CreateQueue implements the SQS service interface.
func (m *sqsServiceMock) CreateQueue(
	ctx context.Context,
	params *sqs.CreateQueueInput,
	optFns ...func(*sqs.Options),
) (*sqs.CreateQueueOutput, error) {
	m.RegisterCall("CreateQueue", params)
	if m.createQueueError != nil {
		return nil, m.createQueueError
	}
	return m.createQueueOutput, nil
}

// GetQueueUrl implements the SQS service interface.
func (m *sqsServiceMock) GetQueueUrl(
	ctx context.Context,
	params *sqs.GetQueueUrlInput,
	optFns ...func(*sqs.Options),
) (*sqs.GetQueueUrlOutput, error) {
	m.RegisterCall("GetQueueUrl", params)
	if m.getQueueUrlError != nil {
		return nil, m.getQueueUrlError
	}
	return m.getQueueUrlOutput, nil
}

// GetQueueAttributes implements the SQS service interface.
func (m *sqsServiceMock) GetQueueAttributes(
	ctx context.Context,
	params *sqs.GetQueueAttributesInput,
	optFns ...func(*sqs.Options),
) (*sqs.GetQueueAttributesOutput, error) {
	m.RegisterCall("GetQueueAttributes", params)
	if m.getQueueAttributesError != nil {
		return nil, m.getQueueAttributesError
	}
	return m.getQueueAttributesOutput, nil
}

// SetQueueAttributes implements the SQS service interface.
func (m *sqsServiceMock) SetQueueAttributes(
	ctx context.Context,
	params *sqs.SetQueueAttributesInput,
	optFns ...func(*sqs.Options),
) (*sqs.SetQueueAttributesOutput, error) {
	m.RegisterCall("SetQueueAttributes", params)
	if m.setQueueAttributesError != nil {
		return nil, m.setQueueAttributesError
	}
	return m.setQueueAttributesOutput, nil
}

// DeleteQueue implements the SQS service interface.
func (m *sqsServiceMock) DeleteQueue(
	ctx context.Context,
	params *sqs.DeleteQueueInput,
	optFns ...func(*sqs.Options),
) (*sqs.DeleteQueueOutput, error) {
	m.RegisterCall("DeleteQueue", params)
	if m.deleteQueueError != nil {
		return nil, m.deleteQueueError
	}
	return m.deleteQueueOutput, nil
}

// TagQueue implements the SQS service interface.
func (m *sqsServiceMock) TagQueue(
	ctx context.Context,
	params *sqs.TagQueueInput,
	optFns ...func(*sqs.Options),
) (*sqs.TagQueueOutput, error) {
	m.RegisterCall("TagQueue", params)
	if m.tagQueueError != nil {
		return nil, m.tagQueueError
	}
	return m.tagQueueOutput, nil
}

// UntagQueue implements the SQS service interface.
func (m *sqsServiceMock) UntagQueue(
	ctx context.Context,
	params *sqs.UntagQueueInput,
	optFns ...func(*sqs.Options),
) (*sqs.UntagQueueOutput, error) {
	m.RegisterCall("UntagQueue", params)
	if m.untagQueueError != nil {
		return nil, m.untagQueueError
	}
	return m.untagQueueOutput, nil
}

// ListQueueTags implements the SQS service interface.
func (m *sqsServiceMock) ListQueueTags(
	ctx context.Context,
	params *sqs.ListQueueTagsInput,
	optFns ...func(*sqs.Options),
) (*sqs.ListQueueTagsOutput, error) {
	m.RegisterCall("ListQueueTags", params)
	if m.listQueueTagsError != nil {
		return nil, m.listQueueTagsError
	}
	return m.listQueueTagsOutput, nil
}
