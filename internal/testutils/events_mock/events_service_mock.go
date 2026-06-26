package eventsmock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	eventsservice "github.com/newstack-cloud/bluelink-provider-aws/services/events/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

type eventsServiceMock struct {
	plugintestutils.MockCalls

	createEventBusOutput *eventbridge.CreateEventBusOutput
	createEventBusError  error

	updateEventBusOutput *eventbridge.UpdateEventBusOutput
	updateEventBusError  error

	deleteEventBusOutput *eventbridge.DeleteEventBusOutput
	deleteEventBusError  error

	describeEventBusOutput *eventbridge.DescribeEventBusOutput
	describeEventBusError  error

	putPermissionOutput *eventbridge.PutPermissionOutput
	putPermissionError  error

	removePermissionOutput *eventbridge.RemovePermissionOutput
	removePermissionError  error

	putRuleOutput *eventbridge.PutRuleOutput
	putRuleError  error

	deleteRuleOutput *eventbridge.DeleteRuleOutput
	deleteRuleError  error

	describeRuleOutput *eventbridge.DescribeRuleOutput
	describeRuleError  error

	listRulesOutput *eventbridge.ListRulesOutput
	listRulesError  error

	putTargetsOutput *eventbridge.PutTargetsOutput
	putTargetsError  error

	removeTargetsOutput *eventbridge.RemoveTargetsOutput
	removeTargetsError  error

	listTargetsByRuleOutput *eventbridge.ListTargetsByRuleOutput
	listTargetsByRuleError  error

	tagResourceOutput *eventbridge.TagResourceOutput
	tagResourceError  error

	untagResourceOutput *eventbridge.UntagResourceOutput
	untagResourceError  error

	listTagsForResourceOutput *eventbridge.ListTagsForResourceOutput
	listTagsForResourceError  error

	createArchiveOutput *eventbridge.CreateArchiveOutput
	createArchiveError  error

	updateArchiveOutput *eventbridge.UpdateArchiveOutput
	updateArchiveError  error

	deleteArchiveOutput *eventbridge.DeleteArchiveOutput
	deleteArchiveError  error

	describeArchiveOutput *eventbridge.DescribeArchiveOutput
	describeArchiveError  error

	createConnectionOutput *eventbridge.CreateConnectionOutput
	createConnectionError  error

	updateConnectionOutput *eventbridge.UpdateConnectionOutput
	updateConnectionError  error

	deleteConnectionOutput *eventbridge.DeleteConnectionOutput
	deleteConnectionError  error

	describeConnectionOutput *eventbridge.DescribeConnectionOutput
	describeConnectionError  error

	createApiDestinationOutput *eventbridge.CreateApiDestinationOutput
	createApiDestinationError  error

	updateApiDestinationOutput *eventbridge.UpdateApiDestinationOutput
	updateApiDestinationError  error

	deleteApiDestinationOutput *eventbridge.DeleteApiDestinationOutput
	deleteApiDestinationError  error

	describeApiDestinationOutput *eventbridge.DescribeApiDestinationOutput
	describeApiDestinationError  error
}

// CreateEventsServiceMock creates a new instance of the EventBridge service mock
// with the provided options.
func CreateEventsServiceMock(options ...EventsServiceMockOption) *eventsServiceMock {
	mock := &eventsServiceMock{}
	for _, option := range options {
		option(mock)
	}
	return mock
}

// EventsServiceMockOption is a function type for configuring the EventBridge service mock.
type EventsServiceMockOption func(*eventsServiceMock)

// WithCreateEventBusOutput sets the mock output for CreateEventBus.
func WithCreateEventBusOutput(output *eventbridge.CreateEventBusOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.createEventBusOutput = output
	}
}

// WithCreateEventBusError sets the mock error for CreateEventBus.
func WithCreateEventBusError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.createEventBusError = err
	}
}

// WithUpdateEventBusOutput sets the mock output for UpdateEventBus.
func WithUpdateEventBusOutput(output *eventbridge.UpdateEventBusOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.updateEventBusOutput = output
	}
}

// WithUpdateEventBusError sets the mock error for UpdateEventBus.
func WithUpdateEventBusError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.updateEventBusError = err
	}
}

// WithDeleteEventBusOutput sets the mock output for DeleteEventBus.
func WithDeleteEventBusOutput(output *eventbridge.DeleteEventBusOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.deleteEventBusOutput = output
	}
}

// WithDeleteEventBusError sets the mock error for DeleteEventBus.
func WithDeleteEventBusError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.deleteEventBusError = err
	}
}

// WithDescribeEventBusOutput sets the mock output for DescribeEventBus.
func WithDescribeEventBusOutput(output *eventbridge.DescribeEventBusOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.describeEventBusOutput = output
	}
}

// WithDescribeEventBusError sets the mock error for DescribeEventBus.
func WithDescribeEventBusError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.describeEventBusError = err
	}
}

// WithPutPermissionOutput sets the mock output for PutPermission.
func WithPutPermissionOutput(output *eventbridge.PutPermissionOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.putPermissionOutput = output
	}
}

// WithPutPermissionError sets the mock error for PutPermission.
func WithPutPermissionError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.putPermissionError = err
	}
}

// WithRemovePermissionOutput sets the mock output for RemovePermission.
func WithRemovePermissionOutput(output *eventbridge.RemovePermissionOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.removePermissionOutput = output
	}
}

// WithRemovePermissionError sets the mock error for RemovePermission.
func WithRemovePermissionError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.removePermissionError = err
	}
}

// WithPutRuleOutput sets the mock output for PutRule.
func WithPutRuleOutput(output *eventbridge.PutRuleOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.putRuleOutput = output
	}
}

// WithPutRuleError sets the mock error for PutRule.
func WithPutRuleError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.putRuleError = err
	}
}

// WithDeleteRuleOutput sets the mock output for DeleteRule.
func WithDeleteRuleOutput(output *eventbridge.DeleteRuleOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.deleteRuleOutput = output
	}
}

// WithDeleteRuleError sets the mock error for DeleteRule.
func WithDeleteRuleError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.deleteRuleError = err
	}
}

// WithDescribeRuleOutput sets the mock output for DescribeRule.
func WithDescribeRuleOutput(output *eventbridge.DescribeRuleOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.describeRuleOutput = output
	}
}

// WithDescribeRuleError sets the mock error for DescribeRule.
func WithDescribeRuleError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.describeRuleError = err
	}
}

// WithListRulesOutput sets the mock output for ListRules.
func WithListRulesOutput(output *eventbridge.ListRulesOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.listRulesOutput = output
	}
}

// WithListRulesError sets the mock error for ListRules.
func WithListRulesError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.listRulesError = err
	}
}

// WithPutTargetsOutput sets the mock output for PutTargets.
func WithPutTargetsOutput(output *eventbridge.PutTargetsOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.putTargetsOutput = output
	}
}

// WithPutTargetsError sets the mock error for PutTargets.
func WithPutTargetsError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.putTargetsError = err
	}
}

// WithRemoveTargetsOutput sets the mock output for RemoveTargets.
func WithRemoveTargetsOutput(output *eventbridge.RemoveTargetsOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.removeTargetsOutput = output
	}
}

// WithRemoveTargetsError sets the mock error for RemoveTargets.
func WithRemoveTargetsError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.removeTargetsError = err
	}
}

// WithListTargetsByRuleOutput sets the mock output for ListTargetsByRule.
func WithListTargetsByRuleOutput(output *eventbridge.ListTargetsByRuleOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.listTargetsByRuleOutput = output
	}
}

// WithListTargetsByRuleError sets the mock error for ListTargetsByRule.
func WithListTargetsByRuleError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.listTargetsByRuleError = err
	}
}

// WithTagResourceOutput sets the mock output for TagResource.
func WithTagResourceOutput(output *eventbridge.TagResourceOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.tagResourceOutput = output
	}
}

// WithTagResourceError sets the mock error for TagResource.
func WithTagResourceError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.tagResourceError = err
	}
}

// WithUntagResourceOutput sets the mock output for UntagResource.
func WithUntagResourceOutput(output *eventbridge.UntagResourceOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.untagResourceOutput = output
	}
}

// WithUntagResourceError sets the mock error for UntagResource.
func WithUntagResourceError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.untagResourceError = err
	}
}

// WithListTagsForResourceOutput sets the mock output for ListTagsForResource.
func WithListTagsForResourceOutput(output *eventbridge.ListTagsForResourceOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.listTagsForResourceOutput = output
	}
}

// WithListTagsForResourceError sets the mock error for ListTagsForResource.
func WithListTagsForResourceError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.listTagsForResourceError = err
	}
}

// WithCreateArchiveOutput sets the mock output for CreateArchive.
func WithCreateArchiveOutput(output *eventbridge.CreateArchiveOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.createArchiveOutput = output
	}
}

// WithCreateArchiveError sets the mock error for CreateArchive.
func WithCreateArchiveError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.createArchiveError = err
	}
}

// WithUpdateArchiveOutput sets the mock output for UpdateArchive.
func WithUpdateArchiveOutput(output *eventbridge.UpdateArchiveOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.updateArchiveOutput = output
	}
}

// WithUpdateArchiveError sets the mock error for UpdateArchive.
func WithUpdateArchiveError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.updateArchiveError = err
	}
}

// WithDeleteArchiveOutput sets the mock output for DeleteArchive.
func WithDeleteArchiveOutput(output *eventbridge.DeleteArchiveOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.deleteArchiveOutput = output
	}
}

// WithDeleteArchiveError sets the mock error for DeleteArchive.
func WithDeleteArchiveError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.deleteArchiveError = err
	}
}

// WithDescribeArchiveOutput sets the mock output for DescribeArchive.
func WithDescribeArchiveOutput(output *eventbridge.DescribeArchiveOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.describeArchiveOutput = output
	}
}

// WithDescribeArchiveError sets the mock error for DescribeArchive.
func WithDescribeArchiveError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.describeArchiveError = err
	}
}

// WithCreateConnectionOutput sets the mock output for CreateConnection.
func WithCreateConnectionOutput(output *eventbridge.CreateConnectionOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.createConnectionOutput = output
	}
}

// WithCreateConnectionError sets the mock error for CreateConnection.
func WithCreateConnectionError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.createConnectionError = err
	}
}

// WithUpdateConnectionOutput sets the mock output for UpdateConnection.
func WithUpdateConnectionOutput(output *eventbridge.UpdateConnectionOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.updateConnectionOutput = output
	}
}

// WithUpdateConnectionError sets the mock error for UpdateConnection.
func WithUpdateConnectionError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.updateConnectionError = err
	}
}

// WithDeleteConnectionOutput sets the mock output for DeleteConnection.
func WithDeleteConnectionOutput(output *eventbridge.DeleteConnectionOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.deleteConnectionOutput = output
	}
}

// WithDeleteConnectionError sets the mock error for DeleteConnection.
func WithDeleteConnectionError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.deleteConnectionError = err
	}
}

// WithDescribeConnectionOutput sets the mock output for DescribeConnection.
func WithDescribeConnectionOutput(output *eventbridge.DescribeConnectionOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.describeConnectionOutput = output
	}
}

// WithDescribeConnectionError sets the mock error for DescribeConnection.
func WithDescribeConnectionError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.describeConnectionError = err
	}
}

// WithCreateApiDestinationOutput sets the mock output for CreateApiDestination.
func WithCreateApiDestinationOutput(output *eventbridge.CreateApiDestinationOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.createApiDestinationOutput = output
	}
}

// WithCreateApiDestinationError sets the mock error for CreateApiDestination.
func WithCreateApiDestinationError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.createApiDestinationError = err
	}
}

// WithUpdateApiDestinationOutput sets the mock output for UpdateApiDestination.
func WithUpdateApiDestinationOutput(output *eventbridge.UpdateApiDestinationOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.updateApiDestinationOutput = output
	}
}

// WithUpdateApiDestinationError sets the mock error for UpdateApiDestination.
func WithUpdateApiDestinationError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.updateApiDestinationError = err
	}
}

// WithDeleteApiDestinationOutput sets the mock output for DeleteApiDestination.
func WithDeleteApiDestinationOutput(output *eventbridge.DeleteApiDestinationOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.deleteApiDestinationOutput = output
	}
}

// WithDeleteApiDestinationError sets the mock error for DeleteApiDestination.
func WithDeleteApiDestinationError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.deleteApiDestinationError = err
	}
}

// WithDescribeApiDestinationOutput sets the mock output for DescribeApiDestination.
func WithDescribeApiDestinationOutput(output *eventbridge.DescribeApiDestinationOutput) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.describeApiDestinationOutput = output
	}
}

// WithDescribeApiDestinationError sets the mock error for DescribeApiDestination.
func WithDescribeApiDestinationError(err error) EventsServiceMockOption {
	return func(mock *eventsServiceMock) {
		mock.describeApiDestinationError = err
	}
}

func (m *eventsServiceMock) CreateEventBus(
	ctx context.Context,
	params *eventbridge.CreateEventBusInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.CreateEventBusOutput, error) {
	m.RegisterCall("CreateEventBus", params)
	if m.createEventBusError != nil {
		return nil, m.createEventBusError
	}
	return m.createEventBusOutput, nil
}

func (m *eventsServiceMock) UpdateEventBus(
	ctx context.Context,
	params *eventbridge.UpdateEventBusInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.UpdateEventBusOutput, error) {
	m.RegisterCall("UpdateEventBus", params)
	if m.updateEventBusError != nil {
		return nil, m.updateEventBusError
	}
	return m.updateEventBusOutput, nil
}

func (m *eventsServiceMock) DeleteEventBus(
	ctx context.Context,
	params *eventbridge.DeleteEventBusInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.DeleteEventBusOutput, error) {
	m.RegisterCall("DeleteEventBus", params)
	if m.deleteEventBusError != nil {
		return nil, m.deleteEventBusError
	}
	return m.deleteEventBusOutput, nil
}

func (m *eventsServiceMock) DescribeEventBus(
	ctx context.Context,
	params *eventbridge.DescribeEventBusInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.DescribeEventBusOutput, error) {
	m.RegisterCall("DescribeEventBus", params)
	if m.describeEventBusError != nil {
		return nil, m.describeEventBusError
	}
	return m.describeEventBusOutput, nil
}

func (m *eventsServiceMock) PutPermission(
	ctx context.Context,
	params *eventbridge.PutPermissionInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.PutPermissionOutput, error) {
	m.RegisterCall("PutPermission", params)
	if m.putPermissionError != nil {
		return nil, m.putPermissionError
	}
	return m.putPermissionOutput, nil
}

func (m *eventsServiceMock) RemovePermission(
	ctx context.Context,
	params *eventbridge.RemovePermissionInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.RemovePermissionOutput, error) {
	m.RegisterCall("RemovePermission", params)
	if m.removePermissionError != nil {
		return nil, m.removePermissionError
	}
	return m.removePermissionOutput, nil
}

func (m *eventsServiceMock) PutRule(
	ctx context.Context,
	params *eventbridge.PutRuleInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.PutRuleOutput, error) {
	m.RegisterCall("PutRule", params)
	if m.putRuleError != nil {
		return nil, m.putRuleError
	}
	return m.putRuleOutput, nil
}

func (m *eventsServiceMock) DeleteRule(
	ctx context.Context,
	params *eventbridge.DeleteRuleInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.DeleteRuleOutput, error) {
	m.RegisterCall("DeleteRule", params)
	if m.deleteRuleError != nil {
		return nil, m.deleteRuleError
	}
	return m.deleteRuleOutput, nil
}

func (m *eventsServiceMock) DescribeRule(
	ctx context.Context,
	params *eventbridge.DescribeRuleInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.DescribeRuleOutput, error) {
	m.RegisterCall("DescribeRule", params)
	if m.describeRuleError != nil {
		return nil, m.describeRuleError
	}
	return m.describeRuleOutput, nil
}

func (m *eventsServiceMock) ListRules(
	ctx context.Context,
	params *eventbridge.ListRulesInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.ListRulesOutput, error) {
	m.RegisterCall("ListRules", params)
	if m.listRulesError != nil {
		return nil, m.listRulesError
	}
	return m.listRulesOutput, nil
}

func (m *eventsServiceMock) PutTargets(
	ctx context.Context,
	params *eventbridge.PutTargetsInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.PutTargetsOutput, error) {
	m.RegisterCall("PutTargets", params)
	if m.putTargetsError != nil {
		return nil, m.putTargetsError
	}
	return m.putTargetsOutput, nil
}

func (m *eventsServiceMock) RemoveTargets(
	ctx context.Context,
	params *eventbridge.RemoveTargetsInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.RemoveTargetsOutput, error) {
	m.RegisterCall("RemoveTargets", params)
	if m.removeTargetsError != nil {
		return nil, m.removeTargetsError
	}
	return m.removeTargetsOutput, nil
}

func (m *eventsServiceMock) ListTargetsByRule(
	ctx context.Context,
	params *eventbridge.ListTargetsByRuleInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.ListTargetsByRuleOutput, error) {
	m.RegisterCall("ListTargetsByRule", params)
	if m.listTargetsByRuleError != nil {
		return nil, m.listTargetsByRuleError
	}
	return m.listTargetsByRuleOutput, nil
}

func (m *eventsServiceMock) TagResource(
	ctx context.Context,
	params *eventbridge.TagResourceInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.TagResourceOutput, error) {
	m.RegisterCall("TagResource", params)
	if m.tagResourceError != nil {
		return nil, m.tagResourceError
	}
	return m.tagResourceOutput, nil
}

func (m *eventsServiceMock) UntagResource(
	ctx context.Context,
	params *eventbridge.UntagResourceInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.UntagResourceOutput, error) {
	m.RegisterCall("UntagResource", params)
	if m.untagResourceError != nil {
		return nil, m.untagResourceError
	}
	return m.untagResourceOutput, nil
}

func (m *eventsServiceMock) ListTagsForResource(
	ctx context.Context,
	params *eventbridge.ListTagsForResourceInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.ListTagsForResourceOutput, error) {
	m.RegisterCall("ListTagsForResource", params)
	if m.listTagsForResourceError != nil {
		return nil, m.listTagsForResourceError
	}
	return m.listTagsForResourceOutput, nil
}

func (m *eventsServiceMock) CreateArchive(
	ctx context.Context,
	params *eventbridge.CreateArchiveInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.CreateArchiveOutput, error) {
	m.RegisterCall("CreateArchive", params)
	if m.createArchiveError != nil {
		return nil, m.createArchiveError
	}
	return m.createArchiveOutput, nil
}

func (m *eventsServiceMock) UpdateArchive(
	ctx context.Context,
	params *eventbridge.UpdateArchiveInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.UpdateArchiveOutput, error) {
	m.RegisterCall("UpdateArchive", params)
	if m.updateArchiveError != nil {
		return nil, m.updateArchiveError
	}
	return m.updateArchiveOutput, nil
}

func (m *eventsServiceMock) DeleteArchive(
	ctx context.Context,
	params *eventbridge.DeleteArchiveInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.DeleteArchiveOutput, error) {
	m.RegisterCall("DeleteArchive", params)
	if m.deleteArchiveError != nil {
		return nil, m.deleteArchiveError
	}
	return m.deleteArchiveOutput, nil
}

func (m *eventsServiceMock) DescribeArchive(
	ctx context.Context,
	params *eventbridge.DescribeArchiveInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.DescribeArchiveOutput, error) {
	m.RegisterCall("DescribeArchive", params)
	if m.describeArchiveError != nil {
		return nil, m.describeArchiveError
	}
	return m.describeArchiveOutput, nil
}

func (m *eventsServiceMock) CreateConnection(
	ctx context.Context,
	params *eventbridge.CreateConnectionInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.CreateConnectionOutput, error) {
	m.RegisterCall("CreateConnection", params)
	if m.createConnectionError != nil {
		return nil, m.createConnectionError
	}
	return m.createConnectionOutput, nil
}

func (m *eventsServiceMock) UpdateConnection(
	ctx context.Context,
	params *eventbridge.UpdateConnectionInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.UpdateConnectionOutput, error) {
	m.RegisterCall("UpdateConnection", params)
	if m.updateConnectionError != nil {
		return nil, m.updateConnectionError
	}
	return m.updateConnectionOutput, nil
}

func (m *eventsServiceMock) DeleteConnection(
	ctx context.Context,
	params *eventbridge.DeleteConnectionInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.DeleteConnectionOutput, error) {
	m.RegisterCall("DeleteConnection", params)
	if m.deleteConnectionError != nil {
		return nil, m.deleteConnectionError
	}
	return m.deleteConnectionOutput, nil
}

func (m *eventsServiceMock) DescribeConnection(
	ctx context.Context,
	params *eventbridge.DescribeConnectionInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.DescribeConnectionOutput, error) {
	m.RegisterCall("DescribeConnection", params)
	if m.describeConnectionError != nil {
		return nil, m.describeConnectionError
	}
	return m.describeConnectionOutput, nil
}

func (m *eventsServiceMock) CreateApiDestination(
	ctx context.Context,
	params *eventbridge.CreateApiDestinationInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.CreateApiDestinationOutput, error) {
	m.RegisterCall("CreateApiDestination", params)
	if m.createApiDestinationError != nil {
		return nil, m.createApiDestinationError
	}
	return m.createApiDestinationOutput, nil
}

func (m *eventsServiceMock) UpdateApiDestination(
	ctx context.Context,
	params *eventbridge.UpdateApiDestinationInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.UpdateApiDestinationOutput, error) {
	m.RegisterCall("UpdateApiDestination", params)
	if m.updateApiDestinationError != nil {
		return nil, m.updateApiDestinationError
	}
	return m.updateApiDestinationOutput, nil
}

func (m *eventsServiceMock) DeleteApiDestination(
	ctx context.Context,
	params *eventbridge.DeleteApiDestinationInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.DeleteApiDestinationOutput, error) {
	m.RegisterCall("DeleteApiDestination", params)
	if m.deleteApiDestinationError != nil {
		return nil, m.deleteApiDestinationError
	}
	return m.deleteApiDestinationOutput, nil
}

func (m *eventsServiceMock) DescribeApiDestination(
	ctx context.Context,
	params *eventbridge.DescribeApiDestinationInput,
	optFns ...func(*eventbridge.Options),
) (*eventbridge.DescribeApiDestinationOutput, error) {
	m.RegisterCall("DescribeApiDestination", params)
	if m.describeApiDestinationError != nil {
		return nil, m.describeApiDestinationError
	}
	return m.describeApiDestinationOutput, nil
}

// CreateEventsServiceMockFactory creates a factory function that returns the mock service.
// This is used for dependency injection in tests.
func CreateEventsServiceMockFactory(
	options ...EventsServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) eventsservice.Service {
	mock := CreateEventsServiceMock(options...)
	return func(awsConfig *aws.Config, providerContext provider.Context) eventsservice.Service {
		return mock
	}
}
